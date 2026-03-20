package resource

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"text/template"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/customvalidator"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/providerdata"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	fwattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const gatewayConfigFilename = "gateway-config"

const (
	defaultSSHGatewayUsername    = "gateway"
	defaultSSHGatewayKeyType     = "ed25519"
	defaultSSHGatewayHostCertTTL = "24h"
	defaultSSHGatewayUserCertTTL = "5m"

	defaultPort        = 8443
	defaultMetricsPort = 9090

	defaultTLSCertificateFile = "/etc/gateway/tls.crt"
	defaultTLSPrivateKeyFile  = "/etc/gateway/tls.key"

	defaultVaultCABundleFile = "/etc/ssl/vault-ca.crt"
	defaultVaultMount        = "ssh"
	defaultVaultRole         = "gateway"
)

var (
	ErrExtractSSHResources        = errors.New("failed to extract ssh_resources")
	ErrExtractKubernetesResources = errors.New("failed to extract kubernetes_resources")
	ErrExtractSSHGateway          = errors.New("failed to extract ssh_gateway")
	ErrExtractTLS                 = errors.New("failed to extract tls")
	ErrExtractSSHCA               = errors.New("failed to extract ssh_ca")
)

//go:embed gateway-config.tmpl.yaml
var gatewayConfigTemplate string

var _ resource.Resource = &gatewayConfig{}

func NewGatewayConfigResource() resource.Resource {
	return &gatewayConfig{}
}

type gatewayConfig struct {
	ProviderConfig providerdata.Config
}

type gatewayConfigModel struct {
	ID                  types.String `tfsdk:"id"`
	Port                types.Int64  `tfsdk:"port"`
	MetricsPort         types.Int64  `tfsdk:"metrics_port"`
	TLS                 types.Object `tfsdk:"tls"`
	SshGateway          types.Object `tfsdk:"ssh_gateway"`
	SshCA               types.Object `tfsdk:"ssh_ca"`
	SSHResources        types.List   `tfsdk:"ssh_resources"`
	KubernetesResources types.List   `tfsdk:"kubernetes_resources"`
	Content             types.String `tfsdk:"content"`
}

type tlsModel struct {
	CertificateFile types.String `tfsdk:"certificate_file"`
	PrivateKeyFile  types.String `tfsdk:"private_key_file"`
}

type sshGatewayModel struct {
	Username    types.String `tfsdk:"username"`
	KeyType     types.String `tfsdk:"key_type"`
	HostCertTTL types.String `tfsdk:"host_cert_ttl"`
	UserCertTTL types.String `tfsdk:"user_cert_ttl"`
}

type sshCAModel struct {
	VaultAddr         types.String `tfsdk:"vault_addr"`
	PrivateKeyFile    types.String `tfsdk:"private_key_file"`
	VaultCABundleFile types.String `tfsdk:"vault_ca_bundle_file"`
	VaultMount        types.String `tfsdk:"vault_mount"`
	VaultRole         types.String `tfsdk:"vault_role"`
	VaultAuthToken    types.String `tfsdk:"vault_auth_token"`
}

type sshResourceRef struct {
	Name     types.String `tfsdk:"name"`
	Address  types.String `tfsdk:"address"`
	Username types.String `tfsdk:"username"`
}

type kubernetesResourceRef struct {
	Name      types.String `tfsdk:"name"`
	Address   types.String `tfsdk:"address"`
	InCluster types.Bool   `tfsdk:"in_cluster"`
}

type gatewayConfigData struct {
	TwingateNetwork     string
	TwingateHost        string
	Port                int64
	MetricsPort         int64
	TLS                 tlsData
	SshGateway          sshGatewayData
	SshCA               sshCAData
	SSHResources        []sshResourceData
	KubernetesResources []kubernetesResourceData
}

type tlsData struct {
	CertificateFile string
	PrivateKeyFile  string
}

type sshGatewayData struct {
	Username    string
	KeyType     string
	HostCertTTL string
	UserCertTTL string
}

type sshCAData struct {
	VaultAddr         string
	PrivateKeyFile    string
	VaultCABundleFile string
	VaultMount        string
	VaultRole         string
	VaultAuthToken    string
}

type sshResourceData struct {
	Name     string
	Address  string
	Username string
}

type kubernetesResourceData struct {
	Name      string
	Address   string
	InCluster bool
}

func tlsAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.CertificateFile: types.StringType,
		attr.PrivateKeyFile:  types.StringType,
	}
}

func sshGatewayAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.Username:    types.StringType,
		attr.KeyType:     types.StringType,
		attr.HostCertTTL: types.StringType,
		attr.UserCertTTL: types.StringType,
	}
}

func sshCAAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.VaultAddr:         types.StringType,
		attr.PrivateKeyFile:    types.StringType,
		attr.VaultCABundleFile: types.StringType,
		attr.VaultMount:        types.StringType,
		attr.VaultRole:         types.StringType,
		attr.VaultAuthToken:    types.StringType,
	}
}

func (r *gatewayConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateGatewayConfig
}

func (r *gatewayConfig) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerdata.ProviderData)
	if !ok {
		return
	}

	r.ProviderConfig = providerData.Config
}

//nolint:funlen
func (r *gatewayConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a Gateway configuration YAML from SSH and Kubernetes resources.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "SHA-256 hash of the generated config content.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.SshCA: schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SSH CA configuration. Specify either vault_addr or private_key_file, not both.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(sshCAAttrTypes(), map[string]fwattr.Value{
					attr.VaultAddr:         types.StringNull(),
					attr.PrivateKeyFile:    types.StringNull(),
					attr.VaultCABundleFile: types.StringValue(defaultVaultCABundleFile),
					attr.VaultMount:        types.StringValue(defaultVaultMount),
					attr.VaultRole:         types.StringValue(defaultVaultRole),
					attr.VaultAuthToken:    types.StringNull(),
				})),
				Attributes: map[string]schema.Attribute{
					attr.VaultAddr: schema.StringAttribute{
						Optional:    true,
						Description: "The Vault server address. Can't be used together with private_key_file.",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(attr.PrivateKeyFile)),
						},
					},
					attr.PrivateKeyFile: schema.StringAttribute{
						Optional:    true,
						Description: "Path to the SSH CA private key file. Can't be used together with vault_addr.",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(attr.VaultAddr)),
						},
					},
					attr.VaultCABundleFile: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("Path to the Vault CA bundle file. Default: %q.", defaultVaultCABundleFile),
						Default:     stringdefault.StaticString(defaultVaultCABundleFile),
					},
					attr.VaultMount: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("Vault SSH secrets engine mount path. Default: %q.", defaultVaultMount),
						Default:     stringdefault.StaticString(defaultVaultMount),
					},
					attr.VaultRole: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("Vault role for signing certificates. Default: %q.", defaultVaultRole),
						Default:     stringdefault.StaticString(defaultVaultRole),
					},
					attr.VaultAuthToken: schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Vault token used for authentication.",
					},
				},
			},
			attr.Port: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Gateway listen port. Default: %d.", defaultPort),
				Default:     int64default.StaticInt64(defaultPort),
			},
			attr.MetricsPort: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Gateway metrics port. Default: %d.", defaultMetricsPort),
				Default:     int64default.StaticInt64(defaultMetricsPort),
			},
			attr.TLS: schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "TLS configuration for the gateway.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(tlsAttrTypes(), map[string]fwattr.Value{
					attr.CertificateFile: types.StringValue(defaultTLSCertificateFile),
					attr.PrivateKeyFile:  types.StringValue(defaultTLSPrivateKeyFile),
				})),
				Attributes: map[string]schema.Attribute{
					attr.CertificateFile: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("Path to the TLS certificate file. Default: %q.", defaultTLSCertificateFile),
						Default:     stringdefault.StaticString(defaultTLSCertificateFile),
					},
					attr.PrivateKeyFile: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("Path to the TLS private key file. Default: %q.", defaultTLSPrivateKeyFile),
						Default:     stringdefault.StaticString(defaultTLSPrivateKeyFile),
					},
				},
			},
			attr.SSHResources: schema.ListAttribute{
				Optional:    true,
				Description: "List of SSH resources. Accepts full twingate_ssh_resource references.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]fwattr.Type{
						attr.Name:     types.StringType,
						attr.Address:  types.StringType,
						attr.Username: types.StringType,
					},
				},
				Validators: []validator.List{
					customvalidator.AtLeastOneNonEmptyWith(path.MatchRoot(attr.KubernetesResources)),
				},
			},
			attr.KubernetesResources: schema.ListAttribute{
				Optional:    true,
				Description: "List of Kubernetes resources. Accepts full twingate_kubernetes_resource references.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]fwattr.Type{
						attr.Name:      types.StringType,
						attr.Address:   types.StringType,
						attr.InCluster: types.BoolType,
					},
				},
				Validators: []validator.List{
					customvalidator.AtLeastOneNonEmptyWith(path.MatchRoot(attr.SSHResources)),
				},
			},
			attr.SshGateway: schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SSH gateway settings. All fields are optional and fall back to built-in defaults.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(sshGatewayAttrTypes(), map[string]fwattr.Value{
					attr.Username:    types.StringValue(defaultSSHGatewayUsername),
					attr.KeyType:     types.StringValue(defaultSSHGatewayKeyType),
					attr.HostCertTTL: types.StringValue(defaultSSHGatewayHostCertTTL),
					attr.UserCertTTL: types.StringValue(defaultSSHGatewayUserCertTTL),
				})),
				Attributes: map[string]schema.Attribute{
					attr.Username: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("SSH gateway username. Default: %q.", defaultSSHGatewayUsername),
						Default:     stringdefault.StaticString(defaultSSHGatewayUsername),
					},
					attr.KeyType: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("SSH key type. Default: %q.", defaultSSHGatewayKeyType),
						Default:     stringdefault.StaticString(defaultSSHGatewayKeyType),
					},
					attr.HostCertTTL: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("Host certificate TTL. Default: %q.", defaultSSHGatewayHostCertTTL),
						Default:     stringdefault.StaticString(defaultSSHGatewayHostCertTTL),
					},
					attr.UserCertTTL: schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: fmt.Sprintf("User certificate TTL. Default: %q.", defaultSSHGatewayUserCertTTL),
						Default:     stringdefault.StaticString(defaultSSHGatewayUserCertTTL),
					},
				},
			},
			attr.Content: schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The generated YAML configuration content.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

//nolint:funlen
func (gateway *gatewayConfigModel) generateContent(ctx context.Context, config providerdata.Config) (string, error) {
	var sshRefs []sshResourceRef
	if diags := gateway.SSHResources.ElementsAs(ctx, &sshRefs, false); diags.HasError() {
		return "", ErrExtractSSHResources
	}

	var k8sRefs []kubernetesResourceRef
	if diags := gateway.KubernetesResources.ElementsAs(ctx, &k8sRefs, false); diags.HasError() {
		return "", ErrExtractKubernetesResources
	}

	sshItems := make([]sshResourceData, 0, len(sshRefs))
	for _, s := range sshRefs {
		sshItems = append(sshItems, sshResourceData{
			Name:     s.Name.ValueString(),
			Address:  s.Address.ValueString(),
			Username: s.Username.ValueString(),
		})
	}

	k8sItems := make([]kubernetesResourceData, 0, len(k8sRefs))
	for _, k := range k8sRefs {
		k8sItems = append(k8sItems, kubernetesResourceData{
			Name:      k.Name.ValueString(),
			Address:   k.Address.ValueString(),
			InCluster: k.InCluster.ValueBool(),
		})
	}

	var tlsGW tlsModel
	if diags := gateway.TLS.As(ctx, &tlsGW, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractTLS
	}

	var sshGW sshGatewayModel
	if diags := gateway.SshGateway.As(ctx, &sshGW, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractSSHGateway
	}

	var sshCA sshCAModel
	if diags := gateway.SshCA.As(ctx, &sshCA, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractSSHCA
	}

	data := gatewayConfigData{
		TwingateNetwork: config.Network,
		TwingateHost:    config.URL,
		Port:            gateway.Port.ValueInt64(),
		MetricsPort:     gateway.MetricsPort.ValueInt64(),
		TLS: tlsData{
			CertificateFile: tlsGW.CertificateFile.ValueString(),
			PrivateKeyFile:  tlsGW.PrivateKeyFile.ValueString(),
		},
		SshGateway: sshGatewayData{
			Username:    sshGW.Username.ValueString(),
			KeyType:     sshGW.KeyType.ValueString(),
			HostCertTTL: sshGW.HostCertTTL.ValueString(),
			UserCertTTL: sshGW.UserCertTTL.ValueString(),
		},
		SshCA: sshCAData{
			VaultAddr:         sshCA.VaultAddr.ValueString(),
			PrivateKeyFile:    sshCA.PrivateKeyFile.ValueString(),
			VaultCABundleFile: sshCA.VaultCABundleFile.ValueString(),
			VaultMount:        sshCA.VaultMount.ValueString(),
			VaultRole:         sshCA.VaultRole.ValueString(),
			VaultAuthToken:    sshCA.VaultAuthToken.ValueString(),
		},
		SSHResources:        sshItems,
		KubernetesResources: k8sItems,
	}

	tmpl, err := template.New(gatewayConfigFilename).Parse(gatewayConfigTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse gateway config template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render gateway config template: %w", err)
	}

	return buf.String(), nil
}

func (r *gatewayConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.storeContent(ctx, req.Plan, &resp.State, &resp.Diagnostics, operationCreate)
}

func (r *gatewayConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.storeContent(ctx, req.State, &resp.State, &resp.Diagnostics, operationRead)
}

func (r *gatewayConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.storeContent(ctx, req.Plan, &resp.State, &resp.Diagnostics, operationUpdate)
}

func (r *gatewayConfig) storeContent(ctx context.Context, getter Getter, setter Setter, diagnostics *diag.Diagnostics, operation string) {
	var state gatewayConfigModel

	diagnostics.Append(getter.Get(ctx, &state)...)

	if diagnostics.HasError() {
		return
	}

	content, err := state.generateContent(ctx, r.ProviderConfig)
	if err != nil {
		addErr(diagnostics, err, operation, TwingateGatewayConfig)

		return
	}

	state.Content = types.StringValue(content)
	state.ID = types.StringValue(fmt.Sprintf("%x", sha256.Sum256([]byte(content))))

	diagnostics.Append(setter.Set(ctx, &state)...)
}

func (r *gatewayConfig) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nothing to delete - purely local resource.
}

type Getter interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}

type Setter interface {
	Set(ctx context.Context, val any) diag.Diagnostics
}
