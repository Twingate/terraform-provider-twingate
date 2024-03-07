package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrConnectorsDatasourceShouldSetOneOptionalNameAttribute = errors.New("Only one of name, name_regex, name_contains, name_exclude, name_prefix or name_suffix must be set.")

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &connectors{}

func NewConnectorsDatasource() datasource.DataSource {
	return &connectors{}
}

type connectors struct {
	client *client.Client
}

type connectorsModel struct {
	ID           types.String     `tfsdk:"id"`
	Name         types.String     `tfsdk:"name"`
	NameRegexp   types.String     `tfsdk:"name_regexp"`
	NameContains types.String     `tfsdk:"name_contains"`
	NameExclude  types.String     `tfsdk:"name_exclude"`
	NamePrefix   types.String     `tfsdk:"name_prefix"`
	NameSuffix   types.String     `tfsdk:"name_suffix"`
	Connectors   []connectorModel `tfsdk:"connectors"`
}

func (d *connectors) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateConnectors
}

func (d *connectors) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *connectors) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Connectors provide connectivity to Remote Networks. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: computedDatasourceIDDescription,
			},

			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only connectors that exactly match this name. If no options are passed it will return all connectors. Only one option can be used at a time.",
			},
			attr.Name + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the name of the connector.",
			},
			attr.Name + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exist in the name of the connector.",
			},
			attr.Name + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the exact value does not exist in the name of the connector.",
			},
			attr.Name + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the connector must start with the value.",
			},
			attr.Name + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the connector must end with the value.",
			},

			// computed
			attr.Connectors: schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Connectors",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Connector.",
						},
						attr.Name: schema.StringAttribute{
							Computed:    true,
							Description: "The Name of the Connector.",
						},
						attr.RemoteNetworkID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Remote Network attached to the Connector.",
						},
						attr.StatusUpdatesEnabled: schema.BoolAttribute{
							Computed:    true,
							Description: "Determines whether status notifications are enabled for the Connector.",
						},
					},
				},
			},
		},
	}
}

func (d *connectors) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectorsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name, filter := getNameFilter(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix)

	if countOptionalAttributes(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrConnectorsDatasourceShouldSetOneOptionalNameAttribute, TwingateResources)

		return
	}

	connectors, err := d.client.ReadConnectors(ctx, name, filter)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateConnectors)

		return
	}

	data.ID = types.StringValue("all-connectors")
	data.Connectors = convertConnectorsToTerraform(connectors)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
