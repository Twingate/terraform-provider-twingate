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
var _ datasource.DataSource = &gateway{}

func NewGatewayDatasource() datasource.DataSource {
	return &gateway{}
}

type gateway struct {
	client *client.Client
}

type gatewayDatasourceModel struct {
	ID              types.String `tfsdk:"id"`
	RemoteNetworkID types.String `tfsdk:"remote_network_id"`
	Address         types.String `tfsdk:"address"`
	X509CAID        types.String `tfsdk:"x509_ca_id"`
	SSHCAID         types.String `tfsdk:"ssh_ca_id"`
}

func (d *gateway) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateGateway
}

func (d *gateway) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gateway) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gateways are the Twingate components that route traffic to remote networks.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Gateway.",
			},
			attr.RemoteNetworkID: schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Remote Network the Gateway belongs to.",
			},
			attr.Address: schema.StringAttribute{
				Computed:    true,
				Description: "The address of the Gateway.",
			},
			attr.X509CAID: schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the X.509 Certificate Authority used for TLS.",
			},
			attr.SSHCAID: schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the SSH Certificate Authority used for SSH access.",
			},
		},
	}
}

func (d *gateway) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data gatewayDatasourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gateway, err := d.client.ReadGateway(ctx, data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateGateway)

		return
	}

	if gateway == nil {
		resp.Diagnostics.AddError(
			"Gateway not found",
			fmt.Sprintf("Gateway with ID %s not found", data.ID.ValueString()),
		)

		return
	}

	data.RemoteNetworkID = types.StringValue(gateway.RemoteNetworkID)
	data.Address = types.StringValue(gateway.Address)
	data.X509CAID = types.StringValue(gateway.X509CAID)

	if gateway.SSHCAID != "" {
		data.SSHCAID = types.StringValue(gateway.SSHCAID)
	} else {
		data.SSHCAID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
