package resource

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/customplanmodifier"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/customvalidator"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// This file holds the access handling shared by `twingate_resource` and the PAM
// resources (`twingate_kubernetes_resource`, `twingate_ssh_resource`,
// `twingate_web_app_resource`): the `is_authoritative` attribute and the
// `access_policy`, `access_group` and `access_service` blocks.

// accessModel is the Terraform side of a resource's access configuration.
type accessModel struct {
	IsAuthoritative types.Bool
	AccessPolicy    types.Set
	GroupAccess     types.Set
	ServiceAccess   types.Set
}

// resourceAccess is the API side of a resource's access: who can reach it.
type resourceAccess struct {
	Groups          []model.AccessGroup
	ServiceAccounts []string
}

// accessBlocks returns the access blocks every resource kind exposes.
func accessBlocks() map[string]schema.Block {
	return map[string]schema.Block{
		attr.AccessGroup:   groupAccessBlock(),
		attr.AccessService: serviceAccessBlock(),
		attr.AccessPolicy:  accessPolicyBlock(),
	}
}

func isAuthoritativeAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		Optional:      true,
		Computed:      true,
		Description:   "Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to `false`, assignments made outside of Terraform will be ignored.",
		PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		Default:       booldefault.StaticBool(true),
	}
}

// isAccessAuthoritative reads the `is_authoritative` flag, falling back to the
// default (`true`) when it is not set: state written before the attribute
// existed holds null, and must keep behaving authoritatively.
func isAccessAuthoritative(val types.Bool) bool {
	if val.IsNull() || val.IsUnknown() {
		return true
	}

	return val.ValueBool()
}

// parseAccess converts and validates the access blocks of a plan.
func parseAccess(plan accessModel) (resourceAccess, *model.AccessPolicy, error) {
	accessGroups, err := getGroupAccessAttribute(plan.GroupAccess)
	if err != nil {
		return resourceAccess{}, nil, fmt.Errorf("failed to parse access_group: %w", err)
	}

	serviceAccountIDs := getServiceAccountAccessAttribute(plan.ServiceAccess)

	for _, access := range accessGroups {
		if access.SecurityPolicyID == nil &&
			(access.AccessPolicy == nil || access.AccessPolicy.Duration == nil) &&
			len(strings.TrimSpace(access.GroupID)) == 0 {
			return resourceAccess{}, nil, ErrInvalidAttributeCombination
		}

		if err := checkGlobalID(access.GroupID); err != nil {
			return resourceAccess{}, nil, err
		}
	}

	for _, id := range serviceAccountIDs {
		if err := checkGlobalID(id); err != nil {
			return resourceAccess{}, nil, err
		}
	}

	accessPolicy, err := getAccessPolicyAttribute(plan.AccessPolicy)
	if err != nil {
		return resourceAccess{}, nil, fmt.Errorf("failed to parse access_policy: %w", err)
	}

	return resourceAccess{Groups: accessGroups, ServiceAccounts: serviceAccountIDs}, accessPolicy, nil
}

// updateResourceAccess brings the remote access of a resource in line with the plan.
// It is a no-op when neither `access_group` nor `access_service` changed.
// readRemote returns the access currently assigned to the resource.
func updateResourceAccess(ctx context.Context, apiClient *client.Client, resourceID string, isAuthoritative bool, plan, state accessModel, desired resourceAccess, readRemote func(ctx context.Context) (resourceAccess, error)) error {
	if plan.GroupAccess.Equal(state.GroupAccess) && plan.ServiceAccess.Equal(state.ServiceAccess) {
		return nil
	}

	remote, err := readRemote(ctx)
	if err != nil {
		return fmt.Errorf("failed to update resource access: failed to get changedIDs: %w", err)
	}

	idsToDelete, toAdd, err := getChangedAccess(isAuthoritative, plan, state, desired, remote)
	if err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	if err := apiClient.RemoveResourceAccess(ctx, resourceID, idsToDelete); err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	if err := apiClient.AddResourceAccess(ctx, resourceID, client.NewAccessInputs(toAdd.ServiceAccounts, toAdd.Groups)); err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	return nil
}

// getChangedAccess returns the principal IDs to remove and the access to add.
// In authoritative mode everything remote that is not desired is removed;
// otherwise only what Terraform itself assigned earlier (the state) is.
func getChangedAccess(isAuthoritative bool, plan, state accessModel, desired, remote resourceAccess) ([]string, resourceAccess, error) {
	old := remote

	if !isAuthoritative {
		var err error

		old, err = getOldNonAuthoritativeAccess(plan, state)
		if err != nil {
			return nil, resourceAccess{}, fmt.Errorf("failed to parse access_group: %w", err)
		}
	}

	// ids to delete
	groupsToDelete := setDifferenceGroups(old.Groups, desired.Groups)
	serviceAccountsToDelete := setDifference(old.ServiceAccounts, desired.ServiceAccounts)

	// access to add
	toAdd := resourceAccess{
		Groups:          setDifferenceGroupAccess(desired.Groups, remote.Groups),
		ServiceAccounts: setDifference(desired.ServiceAccounts, remote.ServiceAccounts),
	}

	return append(groupsToDelete, serviceAccountsToDelete...), toAdd, nil
}

func getOldNonAuthoritativeAccess(plan, state accessModel) (resourceAccess, error) {
	var old resourceAccess

	if !plan.GroupAccess.Equal(state.GroupAccess) {
		groups, err := getGroupAccessAttribute(state.GroupAccess)
		if err != nil {
			return resourceAccess{}, err
		}

		old.Groups = groups
	}

	if !plan.ServiceAccess.Equal(state.ServiceAccess) {
		old.ServiceAccounts = getServiceAccountAccessAttribute(state.ServiceAccess)
	}

	return old, nil
}

// filterNonAuthoritativeAccess keeps only the remote access that the reference
// (plan or state) knows about, so assignments made outside of Terraform are ignored.
func filterNonAuthoritativeAccess(reference accessModel, remote resourceAccess) (resourceAccess, error) {
	refGroupAccess, err := getGroupAccessAttribute(reference.GroupAccess)
	if err != nil {
		return resourceAccess{}, fmt.Errorf("failed to parse access_group: %w", err)
	}

	return resourceAccess{
		Groups:          setIntersectionGroupAccess(refGroupAccess, remote.Groups),
		ServiceAccounts: setIntersection(getServiceAccountAccessAttribute(reference.ServiceAccess), remote.ServiceAccounts),
	}, nil
}

// convertAccessToTerraform converts remote access into its Terraform representation.
// The reference (plan or state) decides which optional values are kept.
// The result must be ignored when diagnostics holds an error.
func convertAccessToTerraform(ctx context.Context, isAuthoritative bool, reference accessModel, remoteAccessPolicy *model.AccessPolicy, remote resourceAccess, diagnostics *diag.Diagnostics) accessModel {
	referenceAccessPolicy, err := getAccessPolicyAttribute(reference.AccessPolicy)
	if err != nil {
		diagnostics.AddAttributeError(
			path.Root(attr.AccessPolicy),
			"failed to parse access_policy attribute",
			err.Error(),
		)

		return reference
	}

	accessPolicy, diags := convertAccessPolicyToTerraform(ctx, remoteAccessPolicy, referenceAccessPolicy)
	diagnostics.Append(diags...)

	groupAccess, diags := convertGroupsAccessToTerraform(ctx, remote.Groups, reference.GroupAccess)
	diagnostics.Append(diags...)

	serviceAccess, diags := convertServiceAccessToTerraform(ctx, remote.ServiceAccounts)
	diagnostics.Append(diags...)

	return accessModel{
		IsAuthoritative: types.BoolValue(isAuthoritative),
		AccessPolicy:    accessPolicy,
		GroupAccess:     groupAccess,
		ServiceAccess:   serviceAccess,
	}
}

// refreshAccess is the read path of the PAM resources: it applies the
// `is_authoritative` flag of the reference to the remote access and converts
// the result to Terraform. The result must be ignored when diagnostics holds an error.
func refreshAccess(ctx context.Context, reference accessModel, remoteAccessPolicy *model.AccessPolicy, remote resourceAccess, diagnostics *diag.Diagnostics) accessModel {
	isAuthoritative := isAccessAuthoritative(reference.IsAuthoritative)

	if !isAuthoritative {
		filtered, err := filterNonAuthoritativeAccess(reference, remote)
		if err != nil {
			diagnostics.AddAttributeError(path.Root(attr.AccessGroup), "failed to parse access_group attribute", err.Error())

			return reference
		}

		remote = filtered
	}

	return convertAccessToTerraform(ctx, isAuthoritative, reference, remoteAccessPolicy, remote, diagnostics)
}

// setImportedAccess writes the access of an imported resource to state.
// Imported resources are authoritative: everything assigned remotely is imported.
func setImportedAccess(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, remoteAccessPolicy *model.AccessPolicy, remote resourceAccess) {
	diagnostics.Append(state.SetAttribute(ctx, path.Root(attr.IsAuthoritative), types.BoolValue(true))...)

	accessPolicy, diags := convertAccessPolicyToTerraformForImport(ctx, remoteAccessPolicy)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(state.SetAttribute(ctx, path.Root(attr.AccessPolicy), accessPolicy)...)

	if len(remote.Groups) > 0 {
		accessGroup, diags := convertGroupsAccessToTerraformForImport(ctx, remote.Groups)
		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		diagnostics.Append(state.SetAttribute(ctx, path.Root(attr.AccessGroup), accessGroup)...)
	}

	if len(remote.ServiceAccounts) > 0 {
		accessService, diags := convertServiceAccessToTerraform(ctx, remote.ServiceAccounts)
		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		diagnostics.Append(state.SetAttribute(ctx, path.Root(attr.AccessService), accessService)...)
	}
}

// suppressAccessPolicyDefaultDrift prevents spurious drift after import: Twingate always
// returns a default access_policy (mode=MANUAL, approval_mode=MANUAL), but if the user's
// config has no access_policy block we should treat those defaults as "unmanaged" and keep
// the plan equal to state so no change is shown or applied.
func suppressAccessPolicyDefaultDrift(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on create (no state yet) or destroy.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Only act when the user has not configured access_policy.
	var configAccessPolicy types.Set
	req.Config.GetAttribute(ctx, path.Root(attr.AccessPolicy), &configAccessPolicy)

	if !configAccessPolicy.IsNull() && len(configAccessPolicy.Elements()) > 0 {
		return
	}

	// Get the state's access_policy value.
	var stateAccessPolicy types.Set
	req.State.GetAttribute(ctx, path.Root(attr.AccessPolicy), &stateAccessPolicy)

	if stateAccessPolicy.IsNull() || len(stateAccessPolicy.Elements()) == 0 {
		return
	}

	// Check that every element in state is a default value.
	for _, elem := range stateAccessPolicy.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			return
		}

		attrs := obj.Attributes()

		mode, _ := attrs[attr.Mode].(types.String)
		approvalMode, _ := attrs[attr.ApprovalMode].(types.String)
		duration, _ := attrs[attr.Duration].(types.String)

		modeIsDefault := mode.IsNull() || mode.ValueString() == model.AccessPolicyModeManual
		approvalModeIsDefault := approvalMode.IsNull() || approvalMode.ValueString() == model.ApprovalModeManual
		durationIsDefault := duration.IsNull() || duration.ValueString() == ""

		if !modeIsDefault || !approvalModeIsDefault || !durationIsDefault {
			return
		}
	}

	// All state values are defaults and config has no access_policy — keep plan = state
	// so Terraform sees no diff and applies no change.
	resp.Plan.SetAttribute(ctx, path.Root(attr.AccessPolicy), stateAccessPolicy)
}

func groupAccessBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
		Description: "Restrict access to certain group",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attr.GroupID: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Group ID that will have permission to access the Resource.",
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`\w+`), "Group ID can't be empty"),
					},
				},
				attr.SecurityPolicyID: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The ID of a `twingate_security_policy` to use as the access policy for the group IDs in the access block. Default is 'Null' which points to `Default Policy` on Admin console.",
					Validators: []validator.String{
						stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName(attr.GroupID)),
					},
					PlanModifiers: []planmodifier.String{
						customplanmodifier.UseNullStringWhenValueOmitted(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				attr.AccessPolicy: accessPolicyBlock(),
			},
		},
	}
}

func serviceAccessBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
		Description: "Restrict access to certain service account",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attr.ServiceAccountID: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The ID of the service account that should have access to this Resource.",
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`\w+`), "ServiceAccount ID can't be empty"),
					},
				},
			},
		},
	}
}

func accessPolicyBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Validators: []validator.Set{
			setvalidator.SizeAtMost(1),
		},
		Description: "Restrict access according to JIT access policy",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attr.Mode: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: fmt.Sprintf("This will set the access_policy mode for the policy. The valid values are `%s`, `%s` and `%s`.", model.AccessPolicyModeManual, model.AccessPolicyModeAutoLock, model.AccessPolicyModeAccessRequest),
					PlanModifiers: []planmodifier.String{
						customplanmodifier.UseNullStringWhenValueOmitted(),
					},
					Validators: []validator.String{
						stringvalidator.OneOf(model.AccessPolicyModeManual, model.AccessPolicyModeAutoLock, model.AccessPolicyModeAccessRequest),
					},
				},

				attr.Duration: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "This will set the access duration for the policy. Duration must be between 1 hour and 365 days. Examples of valid values include `1h` and `2d`.",
					PlanModifiers: []planmodifier.String{
						customplanmodifier.UseNullStringWhenValueOmitted(),
						customplanmodifier.Duration(),
					},
					Validators: []validator.String{
						customvalidator.Duration(),
					},
				},

				attr.ApprovalMode: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: fmt.Sprintf("This will set the approval model for the policy. The valid values are `%s` and `%s`.", model.ApprovalModeAutomatic, model.ApprovalModeManual),
					PlanModifiers: []planmodifier.String{
						customplanmodifier.UseNullStringWhenValueOmitted(),
					},
					Validators: []validator.String{
						stringvalidator.OneOf(model.ApprovalModeAutomatic, model.ApprovalModeManual),
					},
				},
			},
		},
	}
}

func getGroupAccessAttribute(list types.Set) ([]model.AccessGroup, error) {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil, nil
	}

	access := make([]model.AccessGroup, 0, len(list.Elements()))

	for _, item := range list.Elements() {
		obj := item.(types.Object)
		if obj.IsNull() || obj.IsUnknown() {
			continue
		}

		groupVal := obj.Attributes()[attr.GroupID]
		accessGroup := model.AccessGroup{
			GroupID: groupVal.(types.String).ValueString(),
		}

		securityPolicyVal := obj.Attributes()[attr.SecurityPolicyID]
		if securityPolicyVal != nil && !securityPolicyVal.IsNull() && !securityPolicyVal.IsUnknown() {
			accessGroup.SecurityPolicyID = securityPolicyVal.(types.String).ValueStringPointer()
		}

		var (
			err          error
			accessPolicy *model.AccessPolicy
		)

		accessPolicyVal := obj.Attributes()[attr.AccessPolicy]
		if accessPolicyVal != nil && !accessPolicyVal.IsNull() && !accessPolicyVal.IsUnknown() {
			accessPolicyRaw, ok := accessPolicyVal.(types.Set)
			if ok {
				accessPolicy, err = getAccessPolicyAttribute(accessPolicyRaw)
				if err != nil {
					return nil, fmt.Errorf("error parsing access_policy: %w", err)
				}
			}
		}

		accessGroup.AccessPolicy = accessPolicy

		access = append(access, accessGroup)
	}

	return access, nil
}

func getAccessPolicyAttribute(list types.Set) (*model.AccessPolicy, error) {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil, nil //nolint:nilnil
	}

	obj := list.Elements()[0].(types.Object)
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil //nolint:nilnil
	}

	access := &model.AccessPolicy{}

	modeVal := obj.Attributes()[attr.Mode]
	if modeVal != nil && !modeVal.IsNull() && !modeVal.IsUnknown() {
		access.Mode = modeVal.(types.String).ValueStringPointer()
	}

	durationVal := obj.Attributes()[attr.Duration]
	if durationVal != nil && !durationVal.IsNull() && !durationVal.IsUnknown() {
		access.Duration = durationVal.(types.String).ValueStringPointer()
	}

	approvalModeVal := obj.Attributes()[attr.ApprovalMode]
	if approvalModeVal != nil && !approvalModeVal.IsNull() && !approvalModeVal.IsUnknown() {
		access.ApprovalMode = approvalModeVal.(types.String).ValueStringPointer()
	}

	if err := access.Validate(); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return access, nil
}

func getServiceAccountAccessAttribute(list types.Set) []string {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil
	}

	serviceAccountIDs := make([]string, 0, len(list.Elements()))

	for _, item := range list.Elements() {
		obj := item.(types.Object)
		if obj.IsNull() || obj.IsUnknown() {
			continue
		}

		val := obj.Attributes()[attr.ServiceAccountID]
		if val == nil || val.IsNull() || val.IsUnknown() {
			continue
		}

		serviceAccountIDs = append(serviceAccountIDs, val.(types.String).ValueString())
	}

	return serviceAccountIDs
}

func checkGlobalID(val string) error {
	const expectedTokens = 2

	data, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return ErrWrongGlobalID
	}

	tokens := strings.Split(string(data), ":")
	if len(tokens) != expectedTokens {
		return ErrWrongGlobalID
	}

	name := tokens[0]

	if name != "Group" && name != "ServiceAccount" {
		return ErrWrongGlobalID
	}

	return nil
}

func convertAccessPolicyToTerraform(ctx context.Context, accessPolicy, referenceAccessPolicy *model.AccessPolicy) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if accessPolicy == nil || referenceAccessPolicy == nil {
		return makeObjectsSetNull(ctx, accessPolicyAttributeTypes()), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.Mode:         types.StringPointerValue(accessPolicy.Mode),
		attr.Duration:     types.StringPointerValue(accessPolicy.Duration),
		attr.ApprovalMode: types.StringPointerValue(accessPolicy.ApprovalMode),
	}

	if referenceAccessPolicy.ApprovalMode == nil || *referenceAccessPolicy.ApprovalMode == "" {
		attributes[attr.ApprovalMode] = types.StringNull()
	}

	if referenceAccessPolicy.Duration == nil || *referenceAccessPolicy.Duration == "" {
		attributes[attr.Duration] = types.StringNull()
	}

	if referenceAccessPolicy.Duration != nil && accessPolicy.Duration != nil {
		referenceDuration, _ := utils.ParseDurationWithDays(*referenceAccessPolicy.Duration)
		duration, _ := utils.ParseDurationWithDays(*accessPolicy.Duration)

		if duration == referenceDuration {
			attributes[attr.Duration] = types.StringPointerValue(referenceAccessPolicy.Duration)
		}
	}

	// restore duration and approvalMode when mode=MANUAL
	if accessPolicy.Mode != nil && *accessPolicy.Mode == model.AccessPolicyModeManual {
		attributes[attr.ApprovalMode] = types.StringPointerValue(referenceAccessPolicy.ApprovalMode)
		attributes[attr.Duration] = types.StringPointerValue(referenceAccessPolicy.Duration)
	}

	obj, diags := types.ObjectValue(accessPolicyAttributeTypes(), attributes)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessPolicyAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, obj)
}

func convertAccessPolicyToTerraformForImport(ctx context.Context, accessPolicy *model.AccessPolicy) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if accessPolicy == nil {
		return makeObjectsSetNull(ctx, accessPolicyAttributeTypes()), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.Mode:         types.StringPointerValue(accessPolicy.Mode),
		attr.Duration:     types.StringPointerValue(accessPolicy.Duration),
		attr.ApprovalMode: types.StringPointerValue(accessPolicy.ApprovalMode),
	}

	obj, diags := types.ObjectValue(accessPolicyAttributeTypes(), attributes)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessPolicyAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, obj)
}

func convertServiceAccessToTerraform(ctx context.Context, serviceAccounts []string) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(serviceAccounts) == 0 {
		return makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()), diagnostics
	}

	objects := make([]types.Object, 0, len(serviceAccounts))

	for _, account := range serviceAccounts {
		attributes := map[string]tfattr.Value{
			attr.ServiceAccountID: types.StringValue(account),
		}

		obj, diags := types.ObjectValue(accessServiceAccountAttributeTypes(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}

func convertGroupsAccessToTerraform(ctx context.Context, groupAccess []model.AccessGroup, referenceGroupAccess types.Set) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	reference, err := getGroupAccessAttribute(referenceGroupAccess)
	if err != nil {
		diagnostics.AddAttributeError(
			path.Root(attr.AccessGroup),
			"failed to parse access_group attribute",
			err.Error(),
		)

		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	referenceLookup := make(map[string]model.AccessGroup)

	for _, access := range reference {
		referenceLookup[access.GroupID] = access
	}

	if len(groupAccess) == 0 {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	objects := make([]types.Object, 0, len(groupAccess))

	for _, access := range groupAccess {
		attributes := map[string]tfattr.Value{
			attr.GroupID:          types.StringValue(access.GroupID),
			attr.SecurityPolicyID: types.StringPointerValue(access.SecurityPolicyID),
		}

		var referenceAccessPolicy *model.AccessPolicy

		referenceGroup, exists := referenceLookup[access.GroupID]
		if exists {
			referenceAccessPolicy = referenceGroup.AccessPolicy

			// restore security_policy_id when it's not set in the reference, the API returns the effective policy
			if referenceGroup.SecurityPolicyID == nil || *referenceGroup.SecurityPolicyID == "" {
				attributes[attr.SecurityPolicyID] = types.StringPointerValue(referenceGroup.SecurityPolicyID)
			}
		}

		accessPolicy, diags := convertAccessPolicyToTerraform(ctx, access.AccessPolicy, referenceAccessPolicy)
		diagnostics.Append(diags...)

		attributes[attr.AccessPolicy] = accessPolicy

		obj, diags := types.ObjectValue(accessGroupAttributeTypes(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}

func convertGroupsAccessToTerraformForImport(ctx context.Context, groupAccess []model.AccessGroup) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(groupAccess) == 0 {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	objects := make([]types.Object, 0, len(groupAccess))

	for _, access := range groupAccess {
		attributes := map[string]tfattr.Value{
			attr.GroupID:          types.StringValue(access.GroupID),
			attr.SecurityPolicyID: types.StringPointerValue(access.SecurityPolicyID),
		}

		accessPolicy, diags := convertAccessPolicyToTerraformForImport(ctx, access.AccessPolicy)
		diagnostics.Append(diags...)

		attributes[attr.AccessPolicy] = accessPolicy

		obj, diags := types.ObjectValue(accessGroupAttributeTypes(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}

func accessGroupAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.GroupID:          types.StringType,
		attr.SecurityPolicyID: types.StringType,
		attr.AccessPolicy: types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: accessPolicyAttributeTypes(),
			},
		},
	}
}

func accessServiceAccountAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.ServiceAccountID: types.StringType,
	}
}

func accessPolicyAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.Mode:         types.StringType,
		attr.Duration:     types.StringType,
		attr.ApprovalMode: types.StringType,
	}
}
