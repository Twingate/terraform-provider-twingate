package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &x509CertificateAuthority{}

func NewX509CertificateAuthorityDatasource() datasource.DataSource {
	return &x509CertificateAuthority{}
}

type x509CertificateAuthority struct {
	client *client.Client
}

type x509CertificateAuthorityDatasourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

func (d *x509CertificateAuthority) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateX509CertificateAuthority
}

func (d *x509CertificateAuthority) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	httpClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = httpClient
}

func (d *x509CertificateAuthority) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "X509 Certificate Authorities allow Twingate to verify certificates presented by resources during TLS connections.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the X509 Certificate Authority.",
			},
			attr.Name: schema.StringAttribute{
				Computed:    true,
				Description: "The name of the X509 Certificate Authority.",
			},
			attr.Fingerprint: schema.StringAttribute{
				Computed:    true,
				Description: "The SHA-256 fingerprint of the X509 certificate.",
			},
		},
	}
}

func (d *x509CertificateAuthority) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data x509CertificateAuthorityDatasourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	certificateAuthority, err := d.client.ReadX509CertificateAuthority(ctx, data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateX509CertificateAuthority)

		return
	}

	if certificateAuthority == nil {
		resp.Diagnostics.AddError(
			"Certificate Authority not found",
			fmt.Sprintf("Certificate Authority with ID %s not found", data.ID.ValueString()),
		)

		return
	}

	data.Name = types.StringValue(certificateAuthority.Name)
	data.Fingerprint = types.StringValue(certificateAuthority.Fingerprint)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
