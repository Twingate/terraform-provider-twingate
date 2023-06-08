package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ErrAllowedToChangeOnlyManualGroups(group *model.Group) error {
	return fmt.Errorf("Only groups of type %s may be modified. Group %s is a %s type group.", model.GroupTypeManual, group.Name, group.Type) //nolint
}

func NewGroupResource() resource.Resource {
	return &group{}
}

type group struct {
	client *client.Client
}

type groupModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	IsAuthoritative  types.Bool   `tfsdk:"is_authoritative"`
	UserIDs          types.Set    `tfsdk:"user_ids"`
	SecurityPolicyID types.String `tfsdk:"security_policy_id"`
}

func (r *group) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateGroup
}

func (r *group) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

func (r *group) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(attr.ID), req, resp)
}

func (r *group) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).",
		Attributes: map[string]schema.Attribute{
			attr.Name: schema.StringAttribute{
				Required:    true,
				Description: "The name of the group",
			},
			attr.IsAuthoritative: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Determines whether User assignments to this Group will override any existing assignments. Default is `true`. If set to `false`, assignments made outside of Terraform will be ignored.",
			},
			attr.UserIDs: schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of User IDs that have permission to access the Group.",
			},
			attr.SecurityPolicyID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Defines which Security Policy applies to this Group. The Security Policy ID can be obtained from the `twingate_security_policy` and `twingate_security_policies` data sources.",
			},
			// computed
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Group",
			},
		},
	}
}

func (r *group) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.CreateGroup(ctx, convertGroup(&plan))

	r.helper(ctx, group, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func (r *group) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.ReadGroup(ctx, state.ID.ValueString())
	if group != nil {
		group.IsAuthoritative = convertAuthoritativeFlag(state.IsAuthoritative)
	}

	r.helper(ctx, group, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

func (r *group) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state groupModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	group := convertGroup(&plan)
	remoteGroup, err := r.isAllowedToChangeGroup(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationUpdate, TwingateGroup)

	if resp.Diagnostics.HasError() {
		return
	}

	oldIDs := getOldGroupUserIDs(&state, group, remoteGroup)
	if err := r.client.DeleteGroupUsers(ctx, state.ID.ValueString(), setDifference(oldIDs, group.Users)); err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, TwingateGroup)

		return
	}

	group.ID = state.ID.ValueString()
	group, err = r.client.UpdateGroup(ctx, group)

	r.helper(ctx, group, &plan, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func getOldGroupUserIDs(state *groupModel, group, remoteGroup *model.Group) []string {
	if group.IsAuthoritative {
		return remoteGroup.Users
	}

	return convertUsers(state.UserIDs.Elements())
}

func (r *group) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.isAllowedToChangeGroup(ctx, state.ID.ValueString()); err != nil {
		addErr(&resp.Diagnostics, err, operationDelete, TwingateGroup)

		return
	}

	err := r.client.DeleteGroup(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationDelete, TwingateGroup)
}

func (r *group) isAllowedToChangeGroup(ctx context.Context, groupID string) (*model.Group, error) {
	group, err := r.client.ReadGroup(ctx, groupID)
	if err != nil {
		return nil, err //nolint
	}

	if group.Type != model.GroupTypeManual {
		return nil, ErrAllowedToChangeOnlyManualGroups(group)
	}

	return group, nil
}

func (r *group) helper(ctx context.Context, group *model.Group, state *groupModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			respState.RemoveResource(ctx)

			return
		}

		addErr(diagnostics, err, operation, TwingateGroup)

		return
	}

	if !group.IsAuthoritative {
		group.Users = setIntersection(convertUsers(state.UserIDs.Elements()), group.Users)
	}

	state.ID = types.StringValue(group.ID)
	state.Name = types.StringValue(group.Name)
	state.SecurityPolicyID = types.StringValue(group.SecurityPolicyID)
	state.IsAuthoritative = types.BoolValue(group.IsAuthoritative)

	if !state.UserIDs.IsNull() {
		userIDs, diags := types.SetValueFrom(ctx, types.StringType, group.Users)

		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		state.UserIDs = userIDs
	}

	// Set refreshed state
	diags := respState.Set(ctx, state)
	diagnostics.Append(diags...)
}

func convertGroup(data *groupModel) *model.Group {
	return &model.Group{
		ID:               data.ID.ValueString(),
		Name:             data.Name.ValueString(),
		Users:            convertUsers(data.UserIDs.Elements()),
		IsAuthoritative:  convertAuthoritativeFlag(data.IsAuthoritative),
		SecurityPolicyID: data.SecurityPolicyID.ValueString(),
	}
}

func convertUsers(userIDs []tfattr.Value) []string {
	return utils.Map(userIDs, func(item tfattr.Value) string {
		return item.(types.String).ValueString()
	})
}

func convertAuthoritativeFlag(val types.Bool) bool {
	if !val.IsUnknown() {
		return val.ValueBool()
	}

	// default value
	return true
}