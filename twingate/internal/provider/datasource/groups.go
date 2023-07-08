package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &groups{}

func NewGroupsDatasource() datasource.DataSource {
	return &groups{}
}

type groups struct {
	client *client.Client
}

type groupsModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	IsActive types.Bool   `tfsdk:"is_active"`
	Groups   []groupModel `tfsdk:"groups"`
}

func (d *groups) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateGroups
}

func (d *groups) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groups) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: computedDatasourceIDDescription,
			},
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only Groups that exactly match this name.",
			},
			attr.IsActive: schema.BoolAttribute{
				Optional:    true,
				Description: "Returns only Groups matching the specified state.",
			},
			attr.Type: schema.StringAttribute{
				Optional:    true,
				Description: fmt.Sprintf("Returns only Groups of the specified type (valid options: %s, %s or %s).", model.GroupTypeManual, model.GroupTypeSynced, model.GroupTypeSystem),
				Validators: []validator.String{
					stringvalidator.OneOf(model.GroupTypeManual, model.GroupTypeSynced, model.GroupTypeSystem),
				},
			},

			attr.Groups: schema.ListNestedAttribute{
				Computed:    true,
				Optional:    true,
				Description: "List of Groups",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Group",
						},
						attr.Name: schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Group",
						},
						attr.IsActive: schema.BoolAttribute{
							Computed:    true,
							Description: "Indicates if the Group is active",
						},
						attr.Type: schema.StringAttribute{
							Computed:    true,
							Description: "The type of the Group",
						},
						attr.SecurityPolicyID: schema.StringAttribute{
							Computed:    true,
							Description: "The Security Policy assigned to the Group.",
						},
					},
				},
			},
		},
	}
}

func (d *groups) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data groupsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	filter := buildFilter(&data)

	groups, err := d.client.ReadGroups(ctx, filter)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateGroups)

		return
	}

	data.Groups = convertGroupsToTerraform(groups)

	id := "all-groups"
	if filter.HasName() {
		id = "groups-by-name-" + *filter.Name
	}

	data.ID = types.StringValue(id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildFilter(data *groupsModel) *model.GroupsFilter {
	filter := &model.GroupsFilter{
		Name:     data.Name.ValueStringPointer(),
		Type:     data.Type.ValueStringPointer(),
		IsActive: data.IsActive.ValueBoolPointer(),
	}

	if filter.Name == nil && filter.Type == nil && filter.IsActive == nil {
		return nil
	}

	return filter
}
