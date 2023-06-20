package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &connectors{}

func NewConnectorsDatasource() datasource.DataSource {
	return &connectors{}
}

type connectors struct {
	client *client.Client
}

type connectorsModel struct {
	ID         types.String     `tfsdk:"id"`
	Connectors []connectorModel `tfsdk:"connectors"`
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

			// computed
			attr.Connectors: schema.ListNestedAttribute{
				Computed:    true,
				Optional:    true,
				Description: "List of Connectors",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Connector resource.",
						},
						attr.Name: schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Connector.",
						},
						attr.RemoteNetworkID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Remote Network the Connector is attached to.",
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
	connectors, err := d.client.ReadConnectors(ctx)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateConnectors)

		return
	}

	data := connectorsModel{
		ID:         types.StringValue("all-connectors"),
		Connectors: convertConnectorsToTerraform(connectors),
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func convertConnectorsToTerraform(connectors []*model.Connector) []connectorModel {
	return utils.Map(connectors, func(connector *model.Connector) connectorModel {
		return connectorModel{
			Name:                 types.StringValue(connector.Name),
			RemoteNetworkID:      types.StringValue(connector.NetworkID),
			StatusUpdatesEnabled: types.BoolPointerValue(connector.StatusUpdatesEnabled),
		}
	})
}
