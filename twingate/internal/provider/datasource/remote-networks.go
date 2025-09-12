package datasource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrRemoteNetworksDatasourceShouldSetOneOptionalNameAttribute = errors.New("Only one of name, name_regex, name_contains, name_exclude, name_prefix or name_suffix must be set.")

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
	Name           types.String         `tfsdk:"name"`
	NameRegexp     types.String         `tfsdk:"name_regexp"`
	NameContains   types.String         `tfsdk:"name_contains"`
	NameExclude    types.String         `tfsdk:"name_exclude"`
	NamePrefix     types.String         `tfsdk:"name_prefix"`
	NameSuffix     types.String         `tfsdk:"name_suffix"`
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
				Description: computedDatasourceIDDescription,
			},

			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only remote networks that exactly match this name. If no options are passed it will return all remote networks. Only one option can be used at a time.",
			},
			attr.Name + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the name of the remote network.",
			},
			attr.Name + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exist in the name of the remote network.",
			},
			attr.Name + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the exact value does not exist in the name of the remote network.",
			},
			attr.Name + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the remote network must start with the value.",
			},
			attr.Name + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the remote network must end with the value.",
			},

			attr.RemoteNetworks: schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Remote Networks",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Remote Network.",
						},
						attr.Name: schema.StringAttribute{
							Optional:    true,
							Description: "The name of the Remote Network.",
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

	name, filter := GetNameFilter(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix)

	if CountOptionalAttributes(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrRemoteNetworksDatasourceShouldSetOneOptionalNameAttribute, TwingateRemoteNetworks)

		return
	}

	networks, err := d.client.ReadRemoteNetworks(ctx, name, filter)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateRemoteNetworks)

		return
	}

	data.ID = types.StringValue("all-remote-networks")
	data.RemoteNetworks = convertRemoteNetworksToTerraform(networks)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
