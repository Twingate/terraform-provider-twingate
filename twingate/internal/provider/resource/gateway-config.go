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

	defaultGCPMount = "gcp"
)

var (
	ErrExtractSSH                       = errors.New("failed to extract ssh")
	ErrExtractSSHResources              = errors.New("failed to extract ssh.resources")
	ErrExtractKubernetes                = errors.New("failed to extract kubernetes")
	ErrExtractKubernetesResources       = errors.New("failed to extract kubernetes.resources")
	ErrExtractSSHGateway                = errors.New("failed to extract ssh.gateway")
	ErrExtractTLS                       = errors.New("failed to extract tls")
	ErrExtractSSHCA                     = errors.New("failed to extract ssh.ca")
	ErrExtractVault                     = errors.New("failed to extract vault")
	ErrExtractAuth                      = errors.New("failed to extract auth")
	ErrExtractGCP                       = errors.New("failed to extract gcp")
	ErrFailedDecodeVault                = errors.New("failed to decode ssh.ca.vault configuration")
	ErrAtLeastOnePrivateKeyOrAddressSet = errors.New(`At least one of "ssh.ca.private_key_file" or "ssh.ca.vault.address" must be set.`)
	ErrAuthNotSet                       = errors.New("ssh.ca.vault.auth must be set")
)

//go:embed gateway-config.tmpl.yaml
var gatewayConfigTemplate string

var _ resource.Resource = &gatewayConfig{}
var _ resource.ResourceWithValidateConfig = &gatewayConfig{}

func NewGatewayConfigResource() resource.Resource {
	return &gatewayConfig{}
}

type gatewayConfig struct {
	ProviderConfig providerdata.Config
}

type gatewayConfigModel struct {
	ID          types.String `tfsdk:"id"`
	Port        types.Int64  `tfsdk:"port"`
	MetricsPort types.Int64  `tfsdk:"metrics_port"`
	TLS         types.Object `tfsdk:"tls"`
	SSH         types.Object `tfsdk:"ssh"`
	Kubernetes  types.Object `tfsdk:"kubernetes"`
	Content     types.String `tfsdk:"content"`
}

type kubernetesModel struct {
	Resources types.List `tfsdk:"resources"`
}

func (m *kubernetesModel) IsEmptyResources() bool {
	if m == nil {
		return true
	}

	return m.Resources.IsNull() || m.Resources.IsUnknown() || len(m.Resources.Elements()) == 0
}

type sshModel struct {
	Gateway   types.Object `tfsdk:"gateway"`
	CA        types.Object `tfsdk:"ca"`
	Resources types.List   `tfsdk:"resources"`
}

func (m *sshModel) IsEmptyResources() bool {
	if m == nil {
		return true
	}

	return m.Resources.IsNull() || m.Resources.IsUnknown() || len(m.Resources.Elements()) == 0
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
	PrivateKeyFile types.String `tfsdk:"private_key_file"`
	Vault          types.Object `tfsdk:"vault"`
}

func (m *sshCAModel) Validate(ctx context.Context) error {
	if m == nil {
		return nil
	}

	privateKeySet := !m.PrivateKeyFile.IsNull() && !m.PrivateKeyFile.IsUnknown() && m.PrivateKeyFile.ValueString() != ""

	var vaultConf vaultModel
	if !m.Vault.IsNull() && !m.Vault.IsUnknown() {
		if diags := m.Vault.As(ctx, &vaultConf, basetypes.ObjectAsOptions{}); diags.HasError() {
			return ErrFailedDecodeVault
		}
	}

	if !privateKeySet && !vaultConf.IsAddressSet() {
		return ErrAtLeastOnePrivateKeyOrAddressSet
	}

	return nil
}

type vaultModel struct {
	Address      types.String `tfsdk:"address"`
	CABundleFile types.String `tfsdk:"ca_bundle_file"`
	Mount        types.String `tfsdk:"mount"`
	Role         types.String `tfsdk:"role"`
	Auth         types.Object `tfsdk:"auth"`
}

func (m *vaultModel) IsAddressSet() bool {
	if m == nil {
		return false
	}

	return !m.Address.IsNull() && !m.Address.IsUnknown() && m.Address.ValueString() != ""
}

func (m *vaultModel) Validate() error {
	if m == nil {
		return nil
	}

	if m.IsAddressSet() && (m.Auth.IsNull() || m.Auth.IsUnknown()) {
		return ErrAuthNotSet
	}

	return nil
}

type authModel struct {
	Token types.String `tfsdk:"token"`
	GCP   types.Object `tfsdk:"gcp"`
}

type gcpModel struct {
	Role                types.String `tfsdk:"role"`
	Type                types.String `tfsdk:"type"`
	Mount               types.String `tfsdk:"mount"`
	ServiceAccountEmail types.String `tfsdk:"service_account_email"`
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
	TwingateNetwork string
	TwingateHost    string
	Port            int64
	MetricsPort     int64
	TLS             tlsData
	SSH             sshData
	Kubernetes      kubernetesData
}

type tlsData struct {
	CertificateFile string
	PrivateKeyFile  string
}

type sshData struct {
	Gateway   sshGatewayData
	CA        sshCAData
	Resources []sshResourceData
}

type kubernetesData struct {
	Resources []kubernetesResourceData
}

type sshGatewayData struct {
	Username    string
	KeyType     string
	HostCertTTL string
	UserCertTTL string
}

type sshCAData struct {
	PrivateKeyFile string
	Vault          vaultData
}

type vaultData struct {
	Address      string
	CABundleFile string
	Mount        string
	Role         string
	Auth         authData
}

type authData struct {
	Token string
	GCP   gcpData
}

type gcpData struct {
	Role                string
	Type                string
	Mount               string
	ServiceAccountEmail string
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

func vaultAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.Address:      types.StringType,
		attr.CABundleFile: types.StringType,
		attr.Mount:        types.StringType,
		attr.Role:         types.StringType,
		attr.Auth:         types.ObjectType{AttrTypes: authAttrTypes()},
	}
}

func gcpAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.Role:                types.StringType,
		attr.Type:                types.StringType,
		attr.Mount:               types.StringType,
		attr.ServiceAccountEmail: types.StringType,
	}
}

func authAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.Token: types.StringType,
		attr.GCP:   types.ObjectType{AttrTypes: gcpAttrTypes()},
	}
}

func sshCAAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.PrivateKeyFile: types.StringType,
		attr.Vault:          types.ObjectType{AttrTypes: vaultAttrTypes()},
	}
}

func sshResourceElemType() fwattr.Type {
	return types.ObjectType{
		AttrTypes: map[string]fwattr.Type{
			attr.Name:     types.StringType,
			attr.Address:  types.StringType,
			attr.Username: types.StringType,
		},
	}
}

func kubernetesResourceElemType() fwattr.Type {
	return types.ObjectType{
		AttrTypes: map[string]fwattr.Type{
			attr.Name:      types.StringType,
			attr.Address:   types.StringType,
			attr.InCluster: types.BoolType,
		},
	}
}

func kubernetesAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.Resources: types.ListType{ElemType: kubernetesResourceElemType()},
	}
}

func sshAttrTypes() map[string]fwattr.Type {
	return map[string]fwattr.Type{
		attr.Gateway:   types.ObjectType{AttrTypes: sshGatewayAttrTypes()},
		attr.CA:        types.ObjectType{AttrTypes: sshCAAttrTypes()},
		attr.Resources: types.ListType{ElemType: sshResourceElemType()},
	}
}

func defaultGatewayObject() basetypes.ObjectValue {
	return types.ObjectValueMust(sshGatewayAttrTypes(), map[string]fwattr.Value{
		attr.Username:    types.StringValue(defaultSSHGatewayUsername),
		attr.KeyType:     types.StringValue(defaultSSHGatewayKeyType),
		attr.HostCertTTL: types.StringValue(defaultSSHGatewayHostCertTTL),
		attr.UserCertTTL: types.StringValue(defaultSSHGatewayUserCertTTL),
	})
}

func defaultGCPObject() basetypes.ObjectValue {
	return types.ObjectValueMust(gcpAttrTypes(), map[string]fwattr.Value{
		attr.Role:                types.StringNull(),
		attr.Type:                types.StringNull(),
		attr.Mount:               types.StringValue(defaultGCPMount),
		attr.ServiceAccountEmail: types.StringNull(),
	})
}

func defaultAuthObject() basetypes.ObjectValue {
	return types.ObjectValueMust(authAttrTypes(), map[string]fwattr.Value{
		attr.Token: types.StringNull(),
		attr.GCP:   defaultGCPObject(),
	})
}

func defaultVaultObject() basetypes.ObjectValue {
	return types.ObjectValueMust(vaultAttrTypes(), map[string]fwattr.Value{
		attr.Address:      types.StringNull(),
		attr.CABundleFile: types.StringValue(defaultVaultCABundleFile),
		attr.Mount:        types.StringValue(defaultVaultMount),
		attr.Role:         types.StringValue(defaultVaultRole),
		attr.Auth:         defaultAuthObject(),
	})
}

func defaultCAObject() basetypes.ObjectValue {
	return types.ObjectValueMust(sshCAAttrTypes(), map[string]fwattr.Value{
		attr.PrivateKeyFile: types.StringNull(),
		attr.Vault:          defaultVaultObject(),
	})
}

func defaultTLSObject() basetypes.ObjectValue {
	return types.ObjectValueMust(tlsAttrTypes(), map[string]fwattr.Value{
		attr.CertificateFile: types.StringValue(defaultTLSCertificateFile),
		attr.PrivateKeyFile:  types.StringValue(defaultTLSPrivateKeyFile),
	})
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
				Default:     objectdefault.StaticValue(defaultTLSObject()),
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
			attr.SSH: schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SSH configuration block containing gateway, CA, and resource settings.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(sshAttrTypes(), map[string]fwattr.Value{
					attr.Gateway:   defaultGatewayObject(),
					attr.CA:        defaultCAObject(),
					attr.Resources: types.ListValueMust(sshResourceElemType(), []fwattr.Value{}),
				})),
				Attributes: map[string]schema.Attribute{
					attr.Gateway: schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "SSH gateway settings. All fields are optional and fall back to built-in defaults.",
						Default:     objectdefault.StaticValue(defaultGatewayObject()),
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
					attr.CA: schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "SSH CA configuration. Specify either vault.address or private_key_file, not both.",
						Default:     objectdefault.StaticValue(defaultCAObject()),
						Attributes: map[string]schema.Attribute{
							attr.PrivateKeyFile: schema.StringAttribute{
								Optional:    true,
								Description: "Path to the SSH CA private key file. Can't be used together with vault.address.",
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(attr.Vault).AtName(attr.Address)),
								},
							},
							attr.Vault: schema.SingleNestedAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Vault SSH CA configuration.",
								Default:     objectdefault.StaticValue(defaultVaultObject()),
								Attributes: map[string]schema.Attribute{
									attr.Address: schema.StringAttribute{
										Optional:    true,
										Description: "Vault server address. Can't be used together with ca.private_key_file.",
										Validators: []validator.String{
											stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtParent().AtName(attr.PrivateKeyFile)),
										},
									},
									attr.CABundleFile: schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: fmt.Sprintf("Path to the Vault CA bundle file. Default: %q.", defaultVaultCABundleFile),
										Default:     stringdefault.StaticString(defaultVaultCABundleFile),
									},
									attr.Mount: schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: fmt.Sprintf("Vault SSH secrets engine mount path. Default: %q.", defaultVaultMount),
										Default:     stringdefault.StaticString(defaultVaultMount),
									},
									attr.Role: schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: fmt.Sprintf("Vault role for signing certificates. Default: %q.", defaultVaultRole),
										Default:     stringdefault.StaticString(defaultVaultRole),
									},
									attr.Auth: schema.SingleNestedAttribute{
										Optional:    true,
										Computed:    true,
										Description: "Vault authentication configuration.",
										Default:     objectdefault.StaticValue(defaultAuthObject()),
										Attributes: map[string]schema.Attribute{
											attr.Token: schema.StringAttribute{
												Optional:    true,
												Sensitive:   true,
												Description: "Vault token used for authentication. Can't be used together with gcp.",
												Validators: []validator.String{
													stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(attr.GCP)),
												},
											},
											attr.GCP: schema.SingleNestedAttribute{
												Optional:    true,
												Computed:    true,
												Description: "GCP authentication for Vault. Can't be used together with token.",
												Default:     objectdefault.StaticValue(defaultGCPObject()),
												Attributes: map[string]schema.Attribute{
													attr.Role: schema.StringAttribute{
														Optional:    true,
														Description: "GCP IAM role for Vault GCP authentication.",
														Validators: []validator.String{
															stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName(attr.Type)),
														},
													},
													attr.Type: schema.StringAttribute{
														Optional:    true,
														Description: `GCP authentication type for Vault (e.g. "iam" or "gce"). When set to "iam", service_account_email is required.`,
														Validators: []validator.String{
															stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName(attr.Role)),
															customvalidator.AlsoRequiresWhenValueIs(path.MatchRelative().AtParent().AtName(attr.ServiceAccountEmail), "iam"),
														},
													},
													attr.Mount: schema.StringAttribute{
														Optional:    true,
														Computed:    true,
														Description: fmt.Sprintf("Vault GCP auth mount path. Default: %q.", defaultGCPMount),
														Default:     stringdefault.StaticString(defaultGCPMount),
													},
													attr.ServiceAccountEmail: schema.StringAttribute{
														Optional:    true,
														Description: `Service account email. Required when type is "iam".`,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					attr.Resources: schema.ListAttribute{
						Optional:    true,
						Description: "List of SSH resources. Accepts full twingate_ssh_resource references.",
						ElementType: sshResourceElemType(),
					},
				},
			},
			attr.Kubernetes: schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Kubernetes configuration block containing resource settings.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(kubernetesAttrTypes(), map[string]fwattr.Value{
					attr.Resources: types.ListValueMust(kubernetesResourceElemType(), []fwattr.Value{}),
				})),
				Attributes: map[string]schema.Attribute{
					attr.Resources: schema.ListAttribute{
						Optional:    true,
						Description: "List of Kubernetes resources. Accepts full twingate_kubernetes_resource references.",
						ElementType: kubernetesResourceElemType(),
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
	var sshConf sshModel
	if diags := gateway.SSH.As(ctx, &sshConf, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractSSH
	}

	var sshRefs []sshResourceRef
	if diags := sshConf.Resources.ElementsAs(ctx, &sshRefs, false); diags.HasError() {
		return "", ErrExtractSSHResources
	}

	var k8sConf kubernetesModel
	if diags := gateway.Kubernetes.As(ctx, &k8sConf, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractKubernetes
	}

	var k8sRefs []kubernetesResourceRef
	if diags := k8sConf.Resources.ElementsAs(ctx, &k8sRefs, false); diags.HasError() {
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
	if diags := sshConf.Gateway.As(ctx, &sshGW, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractSSHGateway
	}

	var sshCA sshCAModel
	if diags := sshConf.CA.As(ctx, &sshCA, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractSSHCA
	}

	var vaultConf vaultModel
	if diags := sshCA.Vault.As(ctx, &vaultConf, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractVault
	}

	var authConf authModel
	if diags := vaultConf.Auth.As(ctx, &authConf, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractAuth
	}

	var gcpConf gcpModel
	if diags := authConf.GCP.As(ctx, &gcpConf, basetypes.ObjectAsOptions{}); diags.HasError() {
		return "", ErrExtractGCP
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
		SSH: sshData{
			Gateway: sshGatewayData{
				Username:    sshGW.Username.ValueString(),
				KeyType:     sshGW.KeyType.ValueString(),
				HostCertTTL: sshGW.HostCertTTL.ValueString(),
				UserCertTTL: sshGW.UserCertTTL.ValueString(),
			},
			CA: sshCAData{
				PrivateKeyFile: sshCA.PrivateKeyFile.ValueString(),
				Vault: vaultData{
					Address:      vaultConf.Address.ValueString(),
					CABundleFile: vaultConf.CABundleFile.ValueString(),
					Mount:        vaultConf.Mount.ValueString(),
					Role:         vaultConf.Role.ValueString(),
					Auth: authData{
						Token: authConf.Token.ValueString(),
						GCP: gcpData{
							Role:                gcpConf.Role.ValueString(),
							Type:                gcpConf.Type.ValueString(),
							Mount:               gcpConf.Mount.ValueString(),
							ServiceAccountEmail: gcpConf.ServiceAccountEmail.ValueString(),
						},
					},
				},
			},
			Resources: sshItems,
		},
		Kubernetes: kubernetesData{
			Resources: k8sItems,
		},
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

func (r *gatewayConfig) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var (
		sshConf sshModel
		k8sConf kubernetesModel
		cfg     gatewayConfigModel
	)

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !cfg.SSH.IsNull() && !cfg.SSH.IsUnknown() {
		if diags := cfg.SSH.As(ctx, &sshConf, basetypes.ObjectAsOptions{}); diags.HasError() {
			resp.Diagnostics.Append(diags...)

			return
		}
	}

	if !cfg.Kubernetes.IsNull() && !cfg.Kubernetes.IsUnknown() {
		if diags := cfg.Kubernetes.As(ctx, &k8sConf, basetypes.ObjectAsOptions{}); diags.HasError() {
			resp.Diagnostics.Append(diags...)

			return
		}
	}

	if sshConf.IsEmptyResources() && k8sConf.IsEmptyResources() {
		resp.Diagnostics.AddError(
			"Invalid configuration",
			`At least one of "ssh.resources" or "kubernetes.resources" must contain one or more items.`,
		)
	}

	if !sshConf.CA.IsNull() && !sshConf.CA.IsUnknown() {
		var caConf sshCAModel
		if diags := sshConf.CA.As(ctx, &caConf, basetypes.ObjectAsOptions{}); diags.HasError() {
			resp.Diagnostics.Append(diags...)

			return
		}

		if err := caConf.Validate(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Invalid configuration",
				err.Error(),
			)
		}
	}
}

type Getter interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}

type Setter interface {
	Set(ctx context.Context, val any) diag.Diagnostics
}
