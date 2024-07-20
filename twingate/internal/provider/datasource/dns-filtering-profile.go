package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &dnsFilteringProfile{}

func NewDNSFilteringProfileDatasource() datasource.DataSource {
	return &dnsFilteringProfile{}
}

type dnsFilteringProfile struct {
	client *client.Client
}

type dnsFilteringProfileModel struct {
	ID             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	Priority       types.Float64 `tfsdk:"priority"`
	FallbackMethod types.String  `tfsdk:"fallback_method"`
	Groups         types.Set     `tfsdk:"groups"`
	AllowedDomains types.Object  `tfsdk:"allowed_domains"`
	DeniedDomains  types.Object  `tfsdk:"denied_domains"`
	//ContentCategories  types.Object  `tfsdk:"content_categories"`
	//SecurityCategories types.Object  `tfsdk:"security_categories"`
	//PrivacyCategories  types.Object  `tfsdk:"privacy_categories"`
}

func (d *dnsFilteringProfile) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateDNSFilteringProfile
}

func (d *dnsFilteringProfile) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *dnsFilteringProfile) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "DNS filtering gives you the ability to control what websites your users can access. For more information, see Twingate's [documentation](https://www.twingate.com/docs/dns-filtering).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The DNS filtering profile's ID.",
			},
			// computed
			attr.Name: schema.StringAttribute{
				Computed:    true,
				Description: "The DNS filtering profile's name.",
			},
			attr.Priority: schema.Float64Attribute{
				Computed:    true,
				Description: "A floating point number representing the profile's priority.",
			},
			attr.FallbackMethod: schema.StringAttribute{
				Computed:    true,
				Description: "The DNS filtering profile's fallback method. One of AUTOMATIC or STRICT.",
			},
			attr.Groups: schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A set of group IDs that have this as their DNS filtering profile. Defaults to an empty set.",
			},
		},

		Blocks: map[string]schema.Block{
			attr.AllowedDomains: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.Domains: schema.SetAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "A set of allowed domains.",
					},
				},
			},
			attr.DeniedDomains: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.Domains: schema.SetAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "A set of denied domains.",
					},
				},
			},
		},
	}
}

func (d *dnsFilteringProfile) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dnsFilteringProfileModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	connector, err := d.client.ReadConnector(ctx, data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateConnector)

		return
	}

	data.Name = types.StringValue(connector.Name)
	data.RemoteNetworkID = types.StringValue(connector.NetworkID)
	data.StatusUpdatesEnabled = types.BoolPointerValue(connector.StatusUpdatesEnabled)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
