package datasource

import (
	"context"
	"fmt"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &remoteNetwork{}

func NewRemoteNetworkDatasource() datasource.DataSource {
	return &remoteNetwork{}
}

type remoteNetwork struct {
	client *client.Client
}

type remoteNetworkModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Location types.String `tfsdk:"location"`
	Type     types.String `tfsdk:"type"`
}

func (d *remoteNetwork) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateRemoteNetwork
}

func (d *remoteNetwork) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *remoteNetwork) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the Remote Network",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot(attr.Name),
					}...),
				},
			},
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the Remote Network",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot(attr.ID),
					}...),
				},
			},
			attr.Location: schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("The location of the Remote Network. Must be one of the following: %s.", strings.Join(model.Locations, ", ")),
			},
			attr.Type: schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("The type of the Remote Network. Must be one of the following: %s.", strings.Join([]string{model.NetworkTypeRegular, model.NetworkTypeExit}, ", ")),
			},
		},
	}
}

func (d *remoteNetwork) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data remoteNetworkModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	network, err := d.client.ReadRemoteNetwork(ctx, data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateRemoteNetwork)

		return
	}

	data.ID = types.StringValue(network.ID)
	data.Name = types.StringValue(network.Name)
	data.Location = types.StringValue(network.Location)
	data.Type = types.StringValue(network.Type)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
