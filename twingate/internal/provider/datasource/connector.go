package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &connector{}

func NewConnectorDatasource() datasource.DataSource {
	return &connector{}
}

type connector struct {
	client *client.Client
}

type connectorModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	RemoteNetworkID      types.String `tfsdk:"remote_network_id"`
	StatusUpdatesEnabled types.Bool   `tfsdk:"status_updates_enabled"`
	State                types.String `tfsdk:"state"`
	Hostname             types.String `tfsdk:"hostname"`
	Version              types.String `tfsdk:"version"`
	PublicIP             types.String `tfsdk:"public_ip"`
	PrivateIPs           types.Set    `tfsdk:"private_ips"`
}

func (d *connector) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateConnector
}

func (d *connector) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *connector) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Connectors provide connectivity to Remote Networks. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Connector. The ID for the Connector can be obtained from the Admin API or the URL string in the Admin Console.",
			},
			// computed
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
			attr.State: schema.StringAttribute{
				Computed:    true,
				Description: "The Connector's state. One of `ALIVE`, `DEAD_NO_HEARTBEAT`, `DEAD_HEARTBEAT_TOO_OLD` or `DEAD_NO_RELAYS`.",
			},
			attr.Hostname: schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the machine hosting the Connector.",
			},
			attr.Version: schema.StringAttribute{
				Computed:    true,
				Description: "The Connector's version.",
			},
			attr.PublicIP: schema.StringAttribute{
				Computed:    true,
				Description: "The Connector's public IP address.",
			},
			attr.PrivateIPs: schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The Connector's private IP addresses.",
			},
		},
	}
}

func (d *connector) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectorModel

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
	data.State = types.StringValue(connector.State)
	data.Version = types.StringValue(connector.Version)
	data.Hostname = types.StringValue(connector.Hostname)
	data.PublicIP = types.StringValue(connector.PublicIP)
	data.PrivateIPs = utils.MakeStringSet(connector.PrivateIPs)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
