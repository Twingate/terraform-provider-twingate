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

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &exitNetworks{}

func NewExitNetworksDatasource() datasource.DataSource {
	return &exitNetworks{}
}

type exitNetworks struct {
	client   *client.Client
	exitNode bool
}

type exitNetworksModel struct {
	ID           types.String       `tfsdk:"id"`
	Name         types.String       `tfsdk:"name"`
	NameRegexp   types.String       `tfsdk:"name_regexp"`
	NameContains types.String       `tfsdk:"name_contains"`
	NameExclude  types.String       `tfsdk:"name_exclude"`
	NamePrefix   types.String       `tfsdk:"name_prefix"`
	NameSuffix   types.String       `tfsdk:"name_suffix"`
	ExitNetworks []exitNetworkModel `tfsdk:"exit_networks"`
}

func (d *exitNetworks) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateExitNetworks
}

func (d *exitNetworks) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.exitNode = true
}

//nolint:dupl
func (d *exitNetworks) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "TODO: Exit Networks behave similarly to Remote Networks. For more information, see Twingate's [documentation](https://www.twingate.com/docs/exit-networks).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: computedDatasourceIDDescription,
			},

			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only exit networks that exactly match this name. If no options are passed it will return all exit networks. Only one option can be used at a time.",
			},
			attr.Name + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the name of the exit network.",
			},
			attr.Name + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exist in the name of the exit network.",
			},
			attr.Name + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the exact value does not exist in the name of the exit network.",
			},
			attr.Name + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the exit network must start with the value.",
			},
			attr.Name + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the exit network must end with the value.",
			},

			attr.ExitNetworks: schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Exit Networks",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Exit Network.",
						},
						attr.Name: schema.StringAttribute{
							Optional:    true,
							Description: "The name of the Exit Network.",
						},
						attr.Location: schema.StringAttribute{
							Computed:    true,
							Description: fmt.Sprintf("The location of the Exit Network. Must be one of the following: %s.", strings.Join(model.Locations, ", ")),
						},
					},
				},
			},
		},
	}
}

func (d *exitNetworks) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data exitNetworksModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name, filter := getNameFilter(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix)

	if countOptionalAttributes(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrRemoteNetworksDatasourceShouldSetOneOptionalNameAttribute, TwingateExitNetworks)

		return
	}

	networks, err := d.client.ReadRemoteNetworks(ctx, name, filter, d.exitNode)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateExitNetworks)

		return
	}

	data.ID = types.StringValue("all-exit-networks")
	data.ExitNetworks = convertExitNetworksToTerraform(networks)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
