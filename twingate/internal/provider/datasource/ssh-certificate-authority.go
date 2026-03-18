//nolint:dupl
package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/providerdata"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &sshCertificateAuthority{}

func NewSSHCertificateAuthorityDatasource() datasource.DataSource {
	return &sshCertificateAuthority{}
}

type sshCertificateAuthority struct {
	client *client.Client
}

type sshCertificateAuthorityDatasourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

func (d *sshCertificateAuthority) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateSSHCertificateAuthority
}

func (d *sshCertificateAuthority) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerdata.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *providerdata.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	httpClient := providerData.Client
	d.client = httpClient
}

func (d *sshCertificateAuthority) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SSH Certificate Authorities allow Twingate to sign SSH certificates for authenticating users to resources.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the SSH Certificate Authority.",
			},
			attr.Name: schema.StringAttribute{
				Computed:    true,
				Description: "The name of the SSH Certificate Authority.",
			},
			attr.Fingerprint: schema.StringAttribute{
				Computed:    true,
				Description: "The fingerprint of the SSH public key.",
			},
		},
	}
}

func (d *sshCertificateAuthority) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sshCertificateAuthorityDatasourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	certificateAuthority, err := d.client.ReadSSHCertificateAuthority(ctx, data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateSSHCertificateAuthority)

		return
	}

	if certificateAuthority == nil {
		resp.Diagnostics.AddError(
			"SSH Certificate Authority not found",
			fmt.Sprintf("SSH Certificate Authority with ID %s not found", data.ID.ValueString()),
		)

		return
	}

	data.Name = types.StringValue(certificateAuthority.Name)
	data.Fingerprint = types.StringValue(certificateAuthority.Fingerprint)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
