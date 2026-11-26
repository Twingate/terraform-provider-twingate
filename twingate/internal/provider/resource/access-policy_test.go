package resource

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testAccessPolicySet builds an `access_policy` block value; a nil field is null.
func testAccessPolicySet(t *testing.T, mode, duration, approvalMode *string) types.Set {
	t.Helper()

	obj := types.ObjectValueMust(accessPolicyAttributeTypes(), map[string]tfattr.Value{
		attr.Mode:         types.StringPointerValue(mode),
		attr.Duration:     types.StringPointerValue(duration),
		attr.ApprovalMode: types.StringPointerValue(approvalMode),
	})

	set, diags := makeObjectsSet(t.Context(), obj)
	require.False(t, diags.HasError(), diags)

	return set
}

// accessPolicyOf returns the single `access_policy` block of a set.
func accessPolicyOf(t *testing.T, set types.Set) map[string]tfattr.Value {
	t.Helper()

	require.Len(t, set.Elements(), 1)

	obj, ok := set.Elements()[0].(types.Object)
	require.True(t, ok)

	return obj.Attributes()
}

// The access policy read from the API is written to state the way the plan
// spelled it, so equivalent values do not show up as a diff.
func TestConvertAccessPolicyToTerraform(t *testing.T) {
	autoLock, manual := model.AccessPolicyModeAutoLock, model.AccessPolicyModeManual

	t.Run("no remote policy or no planned policy - null", func(t *testing.T) {
		remote := &model.AccessPolicy{Mode: &autoLock}

		for _, policies := range [][2]*model.AccessPolicy{{nil, remote}, {remote, nil}} {
			set, diags := convertAccessPolicyToTerraform(t.Context(), policies[0], policies[1])

			require.False(t, diags.HasError(), diags)
			assert.True(t, set.IsNull())
		}
	})

	cases := []struct {
		name      string
		remote    model.AccessPolicy
		reference model.AccessPolicy
		expected  map[string]tfattr.Value
	}{
		{
			name:      "remote values are used",
			remote:    model.AccessPolicy{Mode: &autoLock, Duration: stringPtr("72h"), ApprovalMode: stringPtr(model.ApprovalModeAutomatic)},
			reference: model.AccessPolicy{Mode: &autoLock, Duration: stringPtr("2d"), ApprovalMode: stringPtr(model.ApprovalModeManual)},
			expected: map[string]tfattr.Value{
				attr.Mode:         types.StringValue(autoLock),
				attr.Duration:     types.StringValue("72h"),
				attr.ApprovalMode: types.StringValue(model.ApprovalModeAutomatic),
			},
		},
		{
			name:      "equal duration keeps the planned spelling",
			remote:    model.AccessPolicy{Mode: &autoLock, Duration: stringPtr("48h"), ApprovalMode: stringPtr(model.ApprovalModeManual)},
			reference: model.AccessPolicy{Mode: &autoLock, Duration: stringPtr("2d"), ApprovalMode: stringPtr(model.ApprovalModeManual)},
			expected: map[string]tfattr.Value{
				attr.Mode:         types.StringValue(autoLock),
				attr.Duration:     types.StringValue("2d"),
				attr.ApprovalMode: types.StringValue(model.ApprovalModeManual),
			},
		},
		{
			name:      "attributes the plan omits stay null",
			remote:    model.AccessPolicy{Mode: &autoLock, Duration: stringPtr("48h"), ApprovalMode: stringPtr(model.ApprovalModeManual)},
			reference: model.AccessPolicy{Mode: &autoLock, Duration: stringPtr("")},
			expected: map[string]tfattr.Value{
				attr.Mode:         types.StringValue(autoLock),
				attr.Duration:     types.StringNull(),
				attr.ApprovalMode: types.StringNull(),
			},
		},
		{
			name:      "manual mode restores the planned duration and approval mode",
			remote:    model.AccessPolicy{Mode: &manual},
			reference: model.AccessPolicy{Mode: &manual, Duration: stringPtr("1h"), ApprovalMode: stringPtr(model.ApprovalModeManual)},
			expected: map[string]tfattr.Value{
				attr.Mode:         types.StringValue(manual),
				attr.Duration:     types.StringValue("1h"),
				attr.ApprovalMode: types.StringValue(model.ApprovalModeManual),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			set, diags := convertAccessPolicyToTerraform(t.Context(), &c.remote, &c.reference)

			require.False(t, diags.HasError(), diags)
			assert.Equal(t, c.expected, accessPolicyOf(t, set))
		})
	}
}

// Without an `access_policy` block in the config, the API default policy held
// in state is kept in the plan, so no change is shown; anything else is left alone.
func TestSuppressAccessPolicyDefaultDrift(t *testing.T) {
	res := &sshResource{}
	nullPolicy := makeObjectsSetNull(t.Context(), accessPolicyAttributeTypes())
	defaultPolicy := testAccessPolicySet(t, stringPtr(model.AccessPolicyModeManual), nil, stringPtr(model.ApprovalModeManual))
	customPolicy := testAccessPolicySet(t, stringPtr(model.AccessPolicyModeAutoLock), stringPtr("2d"), stringPtr(model.ApprovalModeManual))

	cases := []struct {
		name     string
		config   types.Set
		state    *types.Set
		expected types.Set
	}{
		{name: "create - plan is left alone", config: nullPolicy, state: nil, expected: nullPolicy},
		{name: "config declares a policy - plan is left alone", config: customPolicy, state: &defaultPolicy, expected: customPolicy},
		{name: "no policy in state - plan is left alone", config: nullPolicy, state: &nullPolicy, expected: nullPolicy},
		{name: "custom policy in state - plan removes it", config: nullPolicy, state: &customPolicy, expected: nullPolicy},
		{name: "default policy in state - plan keeps it", config: nullPolicy, state: &defaultPolicy, expected: defaultPolicy},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			planModel := testSSHModel(t, types.MapNull(types.StringType))
			planModel.ID = types.StringValue(testResourceID)
			planModel.AccessPolicy = c.config

			plan := planOf(t, res, planModel)
			state := nullStateOf(t, res)

			if c.state != nil {
				stateModel := planModel
				stateModel.AccessPolicy = *c.state
				state = stateOf(t, res, stateModel)
			}

			resp := &resource.ModifyPlanResponse{Plan: plan}
			suppressAccessPolicyDefaultDrift(t.Context(), resource.ModifyPlanRequest{
				Config: tfsdk.Config{Schema: plan.Schema, Raw: plan.Raw},
				Plan:   plan,
				State:  state,
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			var planned types.Set

			diags := resp.Plan.GetAttribute(t.Context(), path.Root(attr.AccessPolicy), &planned)
			require.False(t, diags.HasError(), diags)
			assert.True(t, c.expected.Equal(planned), "expected %s, got %s", c.expected, planned)
		})
	}
}

// An `access_policy` nested in an `access_group` block is parsed and validated
// together with the group.
func TestGetGroupAccessAttributePolicy(t *testing.T) {
	groupAccess := func(policy types.Set) types.Set {
		obj := types.ObjectValueMust(accessGroupAttributeTypes(), map[string]tfattr.Value{
			attr.GroupID:          types.StringValue(testGroupID("1")),
			attr.SecurityPolicyID: types.StringValue("policy-id"),
			attr.AccessPolicy:     policy,
		})

		set, diags := makeObjectsSet(t.Context(), obj)
		require.False(t, diags.HasError(), diags)

		return set
	}

	t.Run("valid policy", func(t *testing.T) {
		policy := testAccessPolicySet(t, stringPtr(model.AccessPolicyModeAutoLock), stringPtr("2d"), stringPtr(model.ApprovalModeManual))

		groups, err := getGroupAccessAttribute(groupAccess(policy))

		require.NoError(t, err)
		require.Len(t, groups, 1)
		assert.Equal(t, testGroupID("1"), groups[0].GroupID)
		assert.Equal(t, stringPtr("policy-id"), groups[0].SecurityPolicyID)
		assert.Equal(t, &model.AccessPolicy{
			Mode:         stringPtr(model.AccessPolicyModeAutoLock),
			Duration:     stringPtr("2d"),
			ApprovalMode: stringPtr(model.ApprovalModeManual),
		}, groups[0].AccessPolicy)
	})

	t.Run("invalid policy", func(t *testing.T) {
		policy := testAccessPolicySet(t, stringPtr(model.AccessPolicyModeAutoLock), stringPtr("not-a-duration"), nil)

		_, err := getGroupAccessAttribute(groupAccess(policy))

		require.ErrorContains(t, err, "error parsing access_policy")
	})

	t.Run("invalid policy fails parseAccess", func(t *testing.T) {
		plan := makeTestAccessModel(t.Context(), types.BoolValue(true), nil, nil)
		plan.GroupAccess = groupAccess(testAccessPolicySet(t, stringPtr(model.AccessPolicyModeAutoLock), stringPtr("not-a-duration"), nil))

		_, _, err := parseAccess(plan)

		require.ErrorContains(t, err, "failed to parse access_group")
	})
}

func TestCheckGlobalID(t *testing.T) {
	encode := func(val string) string {
		return base64.StdEncoding.EncodeToString([]byte(val))
	}

	cases := []struct {
		name  string
		id    string
		valid bool
	}{
		{name: "group", id: encode("Group:1"), valid: true},
		{name: "service account", id: encode("ServiceAccount:1"), valid: true},
		{name: "not base64", id: "not-a-global-id"},
		{name: "no type separator", id: encode("Group")},
		{name: "too many separators", id: encode("Group:1:2")},
		{name: "other type", id: encode("User:1")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := checkGlobalID(c.id)

			if c.valid {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrWrongGlobalID)
			}
		})
	}
}

// invalidGroupAccess is an `access_group` block whose nested policy cannot be parsed.
func invalidGroupAccess(t *testing.T) types.Set {
	t.Helper()

	obj := types.ObjectValueMust(accessGroupAttributeTypes(), map[string]tfattr.Value{
		attr.GroupID:          types.StringValue(testGroupID("1")),
		attr.SecurityPolicyID: types.StringNull(),
		attr.AccessPolicy:     testAccessPolicySet(t, stringPtr(model.AccessPolicyModeAutoLock), stringPtr("not-a-duration"), nil),
	})

	set, diags := makeObjectsSet(t.Context(), obj)
	require.False(t, diags.HasError(), diags)

	return set
}

// An access value that cannot be parsed is reported instead of being dropped.
func TestAccessParseFailures(t *testing.T) {
	ctx := t.Context()

	t.Run("non-authoritative update with an invalid group in state", func(t *testing.T) {
		apiClient, requests := newSequencedMockedClient(t)
		plan := makeTestAccessModel(ctx, types.BoolValue(false), nil, nil)
		state := makeTestAccessModel(ctx, types.BoolValue(false), nil, nil)
		state.GroupAccess = invalidGroupAccess(t)

		err := updateResourceAccess(ctx, apiClient, testResourceID, false, plan, state, resourceAccess{}, func(context.Context) (resourceAccess, error) {
			return resourceAccess{}, nil
		})

		require.ErrorContains(t, err, "failed to parse access_group")
		assert.Empty(t, *requests)
	})

	t.Run("removing access fails", func(t *testing.T) {
		apiClient, requests := newSequencedMockedClient(t, graphqlErrorResponse)
		plan := makeTestAccessModel(ctx, types.BoolValue(true), nil, nil)
		state := makeTestAccessModel(ctx, types.BoolValue(true), nil, []string{"s1"})

		err := updateResourceAccess(ctx, apiClient, testResourceID, true, plan, state, resourceAccess{}, func(context.Context) (resourceAccess, error) {
			return resourceAccess{ServiceAccounts: []string{"s1"}}, nil
		})

		require.ErrorContains(t, err, "failed to update resource access")
		assert.Len(t, *requests, 1)
	})

	t.Run("non-authoritative read with an invalid reference group", func(t *testing.T) {
		reference := makeTestAccessModel(ctx, types.BoolValue(false), nil, nil)
		reference.GroupAccess = invalidGroupAccess(t)

		_, err := filterNonAuthoritativeAccess(reference, resourceAccess{})

		require.ErrorContains(t, err, "failed to parse access_group")
	})

	t.Run("state conversion with an invalid reference policy", func(t *testing.T) {
		var diags diag.Diagnostics

		reference := makeTestAccessModel(ctx, types.BoolValue(true), nil, nil)
		reference.AccessPolicy = testAccessPolicySet(t, stringPtr(model.AccessPolicyModeAutoLock), stringPtr("not-a-duration"), nil)

		convertAccessToTerraform(ctx, true, reference, nil, resourceAccess{}, &diags)

		require.True(t, diags.HasError())
		assert.Equal(t, "failed to parse access_policy attribute", diags.Errors()[0].Summary())
	})

	t.Run("state conversion with an invalid reference group", func(t *testing.T) {
		set, diags := convertGroupsAccessToTerraform(ctx, testGroups(testGroupID("1")), invalidGroupAccess(t))

		require.True(t, diags.HasError())
		assert.Equal(t, "failed to parse access_group attribute", diags.Errors()[0].Summary())
		assert.True(t, set.IsNull())
	})
}

// Outside of the cluster the planned address is sent as it is.
func TestKubernetesAddressOutOfCluster(t *testing.T) {
	plan := testKubernetesModel(t, types.MapNull(types.StringType))
	plan.InCluster = types.BoolValue(false)
	plan.BearerTokenFile = types.StringValue("/var/run/token")
	plan.CAFile = types.StringValue("/var/run/ca.crt")
	plan.Address = types.StringValue(testResourceAddress)

	address, err := kubernetesAddress(&plan)

	require.NoError(t, err)
	assert.Equal(t, testResourceAddress, address)
}
