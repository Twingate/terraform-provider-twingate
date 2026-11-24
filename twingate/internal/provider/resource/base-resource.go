package resource

import (
	"context"
	"maps"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/customplanmodifier"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/provider/providerdata"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// baseResourceModel holds the attributes every resource kind has in common:
// `twingate_resource` and the PAM resources (Kubernetes, SSH and Web App).
// `twingate_resource` is the reference for how these attributes behave.
type baseResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Address          types.String `tfsdk:"address"`
	RemoteNetworkID  types.String `tfsdk:"remote_network_id"`
	IsVisible        types.Bool   `tfsdk:"is_visible"`
	Alias            types.String `tfsdk:"alias"`
	SecurityPolicyID types.String `tfsdk:"security_policy_id"`
	Tags             types.Map    `tfsdk:"tags"`
	TagsAll          types.Map    `tfsdk:"tags_all"`

	accessModel
}

// pamResourceModel holds the attributes the PAM resources have in common:
// they are all reached through a Gateway.
type pamResourceModel struct {
	baseResourceModel

	GatewayID types.String `tfsdk:"gateway_id"`
}

// withBaseAttributes adds the attributes that are defined identically for every
// resource kind to the attributes of a specific one.
func withBaseAttributes(attributes map[string]schema.Attribute) map[string]schema.Attribute {
	maps.Copy(attributes, map[string]schema.Attribute{
		attr.IsVisible: schema.BoolAttribute{
			Optional:      true,
			Computed:      true,
			Description:   "Controls whether this Resource will be visible in the main Resource list in the Twingate Client. Default is `true`.",
			Default:       booldefault.StaticBool(true),
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		},
		attr.Alias: schema.StringAttribute{
			Optional:      true,
			Description:   "Set a DNS alias address for the Resource. Must be a DNS-valid name string.",
			PlanModifiers: []planmodifier.String{customplanmodifier.CaseInsensitiveDiff()},
		},
		attr.SecurityPolicyID: schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The ID of a `twingate_security_policy` to set as this Resource's Security Policy. Default is 'Null' which points to `Default Policy` on Admin console.",
			PlanModifiers: []planmodifier.String{
				customplanmodifier.UseNullStringWhenValueOmitted(),
			},
		},
		attr.Tags: schema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			Description: "A map of key-value pair tags to set on this resource.",
			Default:     mapdefault.StaticValue(types.MapNull(types.StringType)),
		},
		attr.TagsAll:         tagsAllAttribute(),
		attr.IsAuthoritative: isAuthoritativeAttribute(),
	})

	return attributes
}

// baseResource holds what the implementations of all resource kinds share:
// the provider data and the operations that do not depend on the resource kind.
type baseResource struct {
	client      *client.Client
	defaultTags map[string]string
}

func (r *baseResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerdata.ProviderData)
	if !ok {
		return
	}

	r.client = providerData.Client
	r.defaultTags = providerData.DefaultTags
}

// ModifyPlan merges provider default tags with the user-declared tags into
// `tags_all` and suppresses `access_policy` drift when the config omits the block
// and state holds only the API default values. ModifyPlan runs after all
// attribute-level plan modifiers, so this override is final.
func (r *baseResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	planTagsAll(ctx, req, resp, r.defaultTags)
	suppressAccessPolicyDefaultDrift(ctx, req, resp)
}

// deleteResource deletes the resource in state with the delete call of its kind.
func deleteResource(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, resourceName string, deleteByID func(ctx context.Context, resourceID string) error) {
	var resourceID types.String

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(attr.ID), &resourceID)...)

	if resp.Diagnostics.HasError() {
		return
	}

	addErr(&resp.Diagnostics, deleteByID(ctx, resourceID.ValueString()), operationDelete, resourceName)
}

// isBaseResourceChanged reports whether an attribute every resource kind sends
// with its update mutation differs between plan and state. Group and service
// account access is not part of it: access is updated separately.
func isBaseResourceChanged(plan, state *baseResourceModel) bool {
	return !plan.Name.Equal(state.Name) ||
		!plan.Address.Equal(state.Address) ||
		!plan.RemoteNetworkID.Equal(state.RemoteNetworkID) ||
		!plan.IsVisible.Equal(state.IsVisible) ||
		!plan.Alias.Equal(state.Alias) ||
		!plan.SecurityPolicyID.Equal(state.SecurityPolicyID) ||
		!plan.Tags.Equal(state.Tags) || !plan.TagsAll.Equal(state.TagsAll) ||
		!plan.AccessPolicy.Equal(state.AccessPolicy)
}

// isPAMResourceChanged reports whether an attribute every PAM resource sends
// with its update mutation differs between plan and state.
func isPAMResourceChanged(plan, state *pamResourceModel) bool {
	return isBaseResourceChanged(&plan.baseResourceModel, &state.baseResourceModel) ||
		!plan.GatewayID.Equal(state.GatewayID)
}

// keepPlannedAccess records the planned access on a PAM resource read right
// after its access was updated. The update mutation path does the same: its
// payload has no access. A read that soon may still return the previous
// assignments, which Terraform would reject as an inconsistent result; real
// drift shows up on the next refresh.
func keepPlannedAccess(res *model.PAMResource, planned model.PAMResource) {
	res.GroupsAccess = planned.GroupsAccess
	res.ServiceAccounts = planned.ServiceAccounts
}

// convertBaseResource converts and validates the base attributes of a plan.
func convertBaseResource(plan *baseResourceModel) (model.BaseResource, error) {
	access, accessPolicy, err := parseAccess(plan.accessModel)
	if err != nil {
		return model.BaseResource{}, err
	}

	return model.BaseResource{
		Name:             plan.Name.ValueString(),
		Address:          plan.Address.ValueString(),
		RemoteNetworkID:  plan.RemoteNetworkID.ValueString(),
		IsVisible:        getOptionalBool(plan.IsVisible),
		Alias:            getOptionalString(plan.Alias),
		SecurityPolicyID: plan.SecurityPolicyID.ValueStringPointer(),
		Tags:             getKeyValueMap(plan.TagsAll),
		AccessPolicy:     accessPolicy,
		GroupsAccess:     access.Groups,
		ServiceAccounts:  access.ServiceAccounts,
	}, nil
}

// convertPAMResource converts and validates the attributes of a plan that all
// PAM resources share.
func convertPAMResource(plan *pamResourceModel) (model.PAMResource, error) {
	base, err := convertBaseResource(&plan.baseResourceModel)
	if err != nil {
		return model.PAMResource{}, err
	}

	return model.PAMResource{
		BaseResource: base,
		GatewayID:    plan.GatewayID.ValueString(),
	}, nil
}

func baseResourceAccess(res model.BaseResource) resourceAccess {
	return resourceAccess{Groups: res.GroupsAccess, ServiceAccounts: res.ServiceAccounts}
}

// setBaseResourceState writes the base attributes of a resource to state.
// The reference (plan or state) decides which optional values are kept; state and
// reference may be the same model. When the resource is not authoritative, access
// assigned outside of Terraform is ignored.
func setBaseResourceState(ctx context.Context, state, reference *baseResourceModel, res model.BaseResource, isAuthoritative bool, diagnostics *diag.Diagnostics) {
	// reference fields are read up front: state and reference may be the same model
	referenceAccess := reference.accessModel
	referenceTags := reference.Tags
	access := baseResourceAccess(res)

	if !isAuthoritative {
		filtered, err := filterNonAuthoritativeAccess(referenceAccess, access)
		if err != nil {
			diagnostics.AddAttributeError(path.Root(attr.AccessGroup), "failed to parse access_group attribute", err.Error())

			return
		}

		access = filtered
	}

	state.ID = types.StringValue(res.ID)
	state.Name = types.StringValue(res.Name)
	state.Address = types.StringValue(res.Address)
	state.RemoteNetworkID = types.StringValue(res.RemoteNetworkID)
	state.Alias = types.StringPointerValue(res.Alias)
	state.SecurityPolicyID = securityPolicyState(res.SecurityPolicyID, reference.SecurityPolicyID)

	if !state.IsVisible.IsNull() || !reference.IsVisible.IsUnknown() {
		state.IsVisible = types.BoolPointerValue(res.IsVisible)
	}

	accessState := convertAccessToTerraform(ctx, isAuthoritative, referenceAccess, res.AccessPolicy, access, diagnostics)
	if diagnostics.HasError() {
		return
	}

	state.accessModel = accessState
	// `tags` keeps the user-declared value; `tags_all` mirrors the API so drift shows there.
	state.TagsAll = utils.ConvertMapValue(res.Tags)
	state.Tags = referenceTags
}

// securityPolicyState keeps an explicitly configured empty `security_policy_id`:
// the API answers with no policy for it, which would otherwise show up as drift.
func securityPolicyState(remote *string, reference types.String) types.String {
	if remote == nil && !reference.IsNull() && !reference.IsUnknown() && reference.ValueString() == "" {
		return reference
	}

	return types.StringPointerValue(remote)
}

// setPAMResourceState writes the attributes all PAM resources share to state.
// PAM resources refresh a single model, which is both state and reference.
func setPAMResourceState(ctx context.Context, state *pamResourceModel, res model.PAMResource, diagnostics *diag.Diagnostics) {
	setBaseResourceState(ctx, &state.baseResourceModel, &state.baseResourceModel, res.BaseResource, convertAuthoritativeFlag(state.IsAuthoritative), diagnostics)

	state.GatewayID = types.StringValue(res.GatewayID)
}

// setImportedBaseResource writes the base attributes of an imported resource to
// state; the following read fills in the rest.
func setImportedBaseResource(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, res model.BaseResource, defaultTags map[string]string) {
	diagnostics.Append(state.SetAttribute(ctx, path.Root(attr.SecurityPolicyID), types.StringPointerValue(res.SecurityPolicyID))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root(attr.Alias), types.StringPointerValue(res.Alias))...)

	setImportedAccess(ctx, state, diagnostics, res.AccessPolicy, baseResourceAccess(res))

	if diagnostics.HasError() {
		return
	}

	setImportedTags(ctx, state, res.Tags, defaultTags)
}
