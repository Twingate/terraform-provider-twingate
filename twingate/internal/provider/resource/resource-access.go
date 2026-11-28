package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/customplanmodifier"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/provider/providerdata"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const resourceAccessIDSeparator = ":"

var (
	ErrInvalidResourceAccessImportID = errors.New("expected import ID in the format <resource_id>" + resourceAccessIDSeparator + "<group_id>")
	ErrResourceAccessNotFound        = errors.New("group does not have access to the resource")
	ErrResourceAccessNotGroupID      = errors.New("group_id must be the ID of a Group")
)

var _ resource.Resource = &resourceAccess{}

func NewResourceAccessResource() resource.Resource {
	return &resourceAccess{}
}

type resourceAccess struct {
	client *client.Client
}

type resourceAccessModel struct {
	ID               types.String `tfsdk:"id"`
	ResourceID       types.String `tfsdk:"resource_id"`
	GroupID          types.String `tfsdk:"group_id"`
	SecurityPolicyID types.String `tfsdk:"security_policy_id"`
	AccessPolicy     types.Set    `tfsdk:"access_policy"`
}

func (r *resourceAccess) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateResourceAccess
}

func (r *resourceAccess) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerdata.ProviderData)
	if !ok {
		return
	}

	r.client = providerData.Client
}

func (r *resourceAccess) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Grants a Group access to a Twingate Resource. Only `twingate_resource` is supported. The Resource should set `is_authoritative = false` and must not also declare the same Group in its own `access_group` block.",
		Attributes: map[string]schema.Attribute{
			attr.ResourceID: schema.StringAttribute{
				Required:      true,
				Description:   "The ID of the Resource to grant access to.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			attr.GroupID: schema.StringAttribute{
				Required:      true,
				Description:   "The ID of the Group that will have access to the Resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			attr.SecurityPolicyID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of a `twingate_security_policy` to apply to this access. Defaults to 'Null', which points to the `Default Policy` in the Admin console.",
				PlanModifiers: []planmodifier.String{
					customplanmodifier.UseNullStringWhenValueOmitted(),
				},
			},
			attr.ID: schema.StringAttribute{
				Computed:      true,
				Description:   "The ID of this access, in the format `<resource_id>" + resourceAccessIDSeparator + "<group_id>`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
		Blocks: map[string]schema.Block{
			attr.AccessPolicy: accessPolicyBlock(),
		},
	}
}

func (r *resourceAccess) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceID, groupID, found := strings.Cut(req.ID, resourceAccessIDSeparator)
	if !found || resourceID == "" || groupID == "" {
		resp.Diagnostics.AddError("failed to import state", ErrInvalidResourceAccessImportID.Error())

		return
	}

	res, err := r.client.ReadResource(ctx, resourceID)
	if err != nil {
		resp.Diagnostics.AddError("failed to import state", err.Error())

		return
	}

	group := findAccessGroup(res.GroupsAccess, groupID)
	if group == nil {
		resp.Diagnostics.AddError("failed to import state", ErrResourceAccessNotFound.Error())

		return
	}

	accessPolicy, diags := convertAccessPolicyToTerraformForImport(ctx, group.AccessPolicy)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &resourceAccessModel{
		ID:               types.StringValue(req.ID),
		ResourceID:       types.StringValue(resourceID),
		GroupID:          types.StringValue(groupID),
		SecurityPolicyID: types.StringPointerValue(group.SecurityPolicyID),
		AccessPolicy:     accessPolicy,
	})...)
}

func (r *resourceAccess) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceAccessModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(plan.ResourceID.ValueString() + resourceAccessIDSeparator + plan.GroupID.ValueString())

	r.upsert(ctx, &plan, &resp.State, &resp.Diagnostics, operationCreate)
}

func (r *resourceAccess) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceAccessModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.ReadResource(ctx, state.ResourceID.ValueString())

	r.helper(ctx, res, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

func (r *resourceAccess) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceAccessModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.upsert(ctx, &plan, &resp.State, &resp.Diagnostics, operationUpdate)
}

func (r *resourceAccess) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceAccessModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveResourceAccess(ctx, state.ResourceID.ValueString(), []string{state.GroupID.ValueString()})
	addErr(&resp.Diagnostics, err, operationDelete, TwingateResourceAccess)
}

func (r *resourceAccess) upsert(ctx context.Context, plan *resourceAccessModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, operation string) {
	input, err := convertResourceAccessInput(plan)
	if err != nil {
		addErr(diagnostics, err, operation, TwingateResourceAccess)

		return
	}

	if err := r.client.AddResourceAccess(ctx, plan.ResourceID.ValueString(), input); err != nil {
		addErr(diagnostics, err, operation, TwingateResourceAccess)

		return
	}

	res, err := r.client.ReadResource(ctx, plan.ResourceID.ValueString())

	r.helper(ctx, res, plan, respState, diagnostics, err, operation)
}

func convertResourceAccessInput(plan *resourceAccessModel) ([]client.AccessInput, error) {
	if name, err := globalIDType(plan.GroupID.ValueString()); err != nil || name != "Group" {
		return nil, ErrResourceAccessNotGroupID
	}

	accessPolicy, err := getAccessPolicyAttribute(plan.AccessPolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access_policy: %w", err)
	}

	return convertResourceAccess(nil, []model.AccessGroup{{
		GroupID:          plan.GroupID.ValueString(),
		SecurityPolicyID: getOptionalString(plan.SecurityPolicyID),
		AccessPolicy:     accessPolicy,
	}}), nil
}

func (r *resourceAccess) helper(ctx context.Context, res *model.Resource, state *resourceAccessModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) {
	var group *model.AccessGroup

	if err == nil {
		group = findAccessGroup(res.GroupsAccess, state.GroupID.ValueString())
		if group == nil {
			err = ErrResourceAccessNotFound
		}
	}

	if err != nil {
		if operation == operationRead && (errors.Is(err, client.ErrGraphqlResultIsEmpty) || errors.Is(err, ErrResourceAccessNotFound)) {
			respState.RemoveResource(ctx)

			return
		}

		addErr(diagnostics, err, operation, TwingateResourceAccess)

		return
	}

	referenceAccessPolicy, err := getAccessPolicyAttribute(state.AccessPolicy)
	if err != nil {
		diagnostics.AddAttributeError(path.Root(attr.AccessPolicy), "failed to parse access_policy attribute", err.Error())

		return
	}

	if state.SecurityPolicyID.ValueString() != "" {
		state.SecurityPolicyID = types.StringPointerValue(group.SecurityPolicyID)
	}

	accessPolicy, diags := convertAccessPolicyToTerraform(ctx, group.AccessPolicy, referenceAccessPolicy)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return
	}

	state.AccessPolicy = accessPolicy

	diagnostics.Append(respState.Set(ctx, state)...)
}

func findAccessGroup(groups []model.AccessGroup, groupID string) *model.AccessGroup {
	for i := range groups {
		if groups[i].GroupID == groupID {
			return &groups[i]
		}
	}

	return nil
}
