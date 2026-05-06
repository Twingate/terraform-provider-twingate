package datasource

import (
	"context"
	"fmt"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/providerdata"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	oidcAPI     = "/oidc/v2"
	httpsPrefix = "https://"
	defaultType = model.SyncToS3TypeOIDC
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &syncToS3{}

func NewSyncToS3Datasource() datasource.DataSource {
	return &syncToS3{}
}

type syncToS3 struct {
	client      *client.Client
	regionalURL string
}

type syncToS3Model struct {
	ID         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	OidcURL    types.String `tfsdk:"oidc_url"`
	OidcPrefix types.String `tfsdk:"oidc_prefix"`
}

func (d *syncToS3) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateSyncToS3
}

func (d *syncToS3) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = providerData.Client
	d.regionalURL = providerData.Config.RegionalURL
}

func (d *syncToS3) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: userDescription,
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: computedDatasourceIDDescription,
			},
			attr.Type: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf(`The type of the resource. One of: %s. Defaults to "oidc".`, utils.DocList(model.SyncToS3Types)),
				Validators: []validator.String{
					stringvalidator.OneOf(append(model.SyncToS3Types, "")...),
				},
			},
			attr.OidcURL: schema.StringAttribute{
				Computed:    true,
				Description: `The IAM Identity Provider URL (return only if type "oidc"). Example: https://tenant.twingate.com` + oidcAPI,
			},
			attr.OidcPrefix: schema.StringAttribute{
				Computed:    true,
				Description: `The IAM Identity Provider prefix (return only if type "oidc"). Example: tenant.twingate.com` + oidcAPI,
			},
		},
	}
}

func (d *syncToS3) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data syncToS3Model

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.IsNull() {
		data.Type = types.StringValue(defaultType)
	}

	data.ID = types.StringValue(terraformSyncToS3DatasourceID(data.Type.ValueString()))
	data.OidcURL = types.StringNull()
	data.OidcPrefix = types.StringNull()

	if data.Type.ValueString() == model.SyncToS3TypeOIDC {
		url := d.regionalURL + oidcAPI
		data.OidcURL = types.StringValue(url)
		data.OidcPrefix = types.StringValue(strings.TrimPrefix(url, httpsPrefix))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func terraformSyncToS3DatasourceID(syncType string) string {
	return "sync-to-s3-" + syncType
}
