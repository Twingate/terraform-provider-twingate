package datasources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &remoteNetworks{}

func NewRemoteNetworksDatasource() datasource.DataSource {
	return &remoteNetworks{}
}

type remoteNetworks struct {
	client *client.Client
}

type remoteNetworksModel struct {
	ID             types.String         `tfsdk:"id"`
	RemoteNetworks []remoteNetworkModel `tfsdk:"remote_networks"`
}

func (d *remoteNetworks) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateRemoteNetworks
}

func (d *remoteNetworks) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *remoteNetworks) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Remote Networks datasource",
			},

			attr.RemoteNetworks: schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Remote Networks",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Remote Network",
						},
						attr.Name: schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Remote Network",
						},
						attr.Location: schema.StringAttribute{
							Computed:    true,
							Description: fmt.Sprintf("The location of the Remote Network. Must be one of the following: %s.", strings.Join(model.Locations, ", ")),
						},
					},
				},
			},
		},
	}
}

func (d *remoteNetworks) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data remoteNetworksModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	networks, err := d.client.ReadRemoteNetworks(ctx)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateRemoteNetworks)

		return
	}

	data.ID = types.StringValue("all-remote-networks")
	data.RemoteNetworks = convertRemoteNetworksToTerraform(networks)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func convertRemoteNetworksToTerraform(networks []*model.RemoteNetwork) []remoteNetworkModel {
	return utils.Map(networks, func(network *model.RemoteNetwork) remoteNetworkModel {
		return remoteNetworkModel{
			ID:       types.StringValue(network.ID),
			Name:     types.StringValue(network.Name),
			Location: types.StringValue(network.Location),
		}
	})
}
