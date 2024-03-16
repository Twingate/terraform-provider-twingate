package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrGroupsDatasourceShouldSetOneOptionalNameAttribute = errors.New("Only one of name, name_regex, name_contains, name_exclude, name_prefix or name_suffix must be set.")

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &groups{}

func NewGroupsDatasource() datasource.DataSource {
	return &groups{}
}

type groups struct {
	client *client.Client
}

type groupsModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	NameRegexp   types.String `tfsdk:"name_regexp"`
	NameContains types.String `tfsdk:"name_contains"`
	NameExclude  types.String `tfsdk:"name_exclude"`
	NamePrefix   types.String `tfsdk:"name_prefix"`
	NameSuffix   types.String `tfsdk:"name_suffix"`
	Types        types.Set    `tfsdk:"types"`
	IsActive     types.Bool   `tfsdk:"is_active"`
	Groups       []groupModel `tfsdk:"groups"`
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

//nolint:funlen
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
				Description: "Returns only groups that exactly match this name. If no options are passed it will return all resources. Only one option can be used at a time.",
			},
			attr.Name + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the name of the group.",
			},
			attr.Name + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exist in the name of the group.",
			},
			attr.Name + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the exact value does not exist in the name of the group.",
			},
			attr.Name + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the group must start with the value.",
			},
			attr.Name + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the group must end with the value.",
			},
			attr.IsActive: schema.BoolAttribute{
				Optional:    true,
				Description: "Returns only Groups matching the specified state.",
			},
			attr.Types: schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: fmt.Sprintf("Returns groups that match a list of types. valid types: `%s`, `%s`, `%s`.", model.GroupTypeManual, model.GroupTypeSynced, model.GroupTypeSystem),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf(model.GroupTypeManual, model.GroupTypeSynced, model.GroupTypeSystem)),
				},
			},

			attr.Groups: schema.ListNestedAttribute{
				Computed:    true,
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

	if countOptionalAttributes(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrGroupsDatasourceShouldSetOneOptionalNameAttribute, TwingateGroups)

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

//nolint:cyclop
func buildFilter(data *groupsModel) *model.GroupsFilter {
	var name, filter string

	if data.Name.ValueString() != "" {
		name = data.Name.ValueString()
	}

	if data.NameRegexp.ValueString() != "" {
		name = data.NameRegexp.ValueString()
		filter = attr.FilterByRegexp
	}

	if data.NameContains.ValueString() != "" {
		name = data.NameContains.ValueString()
		filter = attr.FilterByContains
	}

	if data.NameExclude.ValueString() != "" {
		name = data.NameExclude.ValueString()
		filter = attr.FilterByExclude
	}

	if data.NamePrefix.ValueString() != "" {
		name = data.NamePrefix.ValueString()
		filter = attr.FilterByPrefix
	}

	if data.NameSuffix.ValueString() != "" {
		name = data.NameSuffix.ValueString()
		filter = attr.FilterBySuffix
	}

	groupFilter := &model.GroupsFilter{
		Name:       &name,
		NameFilter: filter,
		IsActive:   data.IsActive.ValueBoolPointer(),
	}

	if len(data.Types.Elements()) > 0 {
		groupFilter.Types = utils.Map(data.Types.Elements(), func(item tfattr.Value) string {
			return item.(types.String).ValueString()
		})
	}

	if groupFilter.Name == nil && len(groupFilter.Types) == 0 && groupFilter.IsActive == nil {
		return nil
	}

	return groupFilter
}
