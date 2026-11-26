package resource

import (
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAlias = "app.int"

func testBaseResourceModel(t *testing.T) baseResourceModel {
	t.Helper()

	return baseResourceModel{
		ID:               types.StringValue(testResourceID),
		Name:             types.StringValue(testResourceName),
		Address:          types.StringValue(testResourceAddress),
		RemoteNetworkID:  types.StringValue(testRemoteNetworkID),
		IsVisible:        types.BoolValue(true),
		Alias:            types.StringNull(),
		SecurityPolicyID: types.StringNull(),
		Tags:             types.MapNull(types.StringType),
		TagsAll:          types.MapNull(types.StringType),
		accessModel:      makeTestAccessModel(t.Context(), types.BoolValue(true), nil, nil),
	}
}

func testBaseResource() model.BaseResource {
	return model.BaseResource{
		ID:              testResourceID,
		Name:            testResourceName,
		Address:         testResourceAddress,
		RemoteNetworkID: testRemoteNetworkID,
		IsVisible:       boolPtr(true),
	}
}

func TestConvertBaseResource(t *testing.T) {
	group, serviceAccount := testGroupID("1"), testServiceAccountID("1")

	plan := testBaseResourceModel(t)
	plan.Alias = types.StringValue(testAlias)
	plan.IsVisible = types.BoolValue(false)
	plan.SecurityPolicyID = types.StringValue("policy-id")
	plan.Tags = stringMap(map[string]string{"env": "prod"})
	plan.TagsAll = stringMap(map[string]string{"env": "prod", "owner": "infra"})
	plan.accessModel = makeTestAccessModel(t.Context(), types.BoolValue(true), []string{group}, []string{serviceAccount})

	base, err := convertBaseResource(&plan)

	require.NoError(t, err)
	assert.Equal(t, model.BaseResource{
		Name:             testResourceName,
		Address:          testResourceAddress,
		RemoteNetworkID:  testRemoteNetworkID,
		IsVisible:        boolPtr(false),
		Alias:            stringPtr(testAlias),
		SecurityPolicyID: stringPtr("policy-id"),
		// the API receives the merged tags, not only the user-declared ones
		Tags:            map[string]string{"env": "prod", "owner": "infra"},
		GroupsAccess:    testGroups(group),
		ServiceAccounts: []string{serviceAccount},
	}, base)
}

func TestConvertPAMResource(t *testing.T) {
	plan := pamResourceModel{baseResourceModel: testBaseResourceModel(t), GatewayID: types.StringValue(testGatewayID)}

	pam, err := convertPAMResource(&plan)

	require.NoError(t, err)
	assert.Equal(t, testGatewayID, pam.GatewayID)
	assert.Equal(t, testResourceName, pam.Name)

	plan.accessModel = makeTestAccessModel(t.Context(), types.BoolValue(true), []string{"not-a-global-id"}, nil)

	_, err = convertPAMResource(&plan)
	require.ErrorIs(t, err, ErrWrongGlobalID)
}

func TestSecurityPolicyState(t *testing.T) {
	cases := []struct {
		name      string
		remote    *string
		reference types.String
		expected  types.String
	}{
		{name: "remote policy wins", remote: stringPtr("policy"), reference: types.StringValue(""), expected: types.StringValue("policy")},
		{name: "configured empty policy is kept", remote: nil, reference: types.StringValue(""), expected: types.StringValue("")},
		{name: "no policy anywhere", remote: nil, reference: types.StringNull(), expected: types.StringNull()},
		{name: "unknown reference", remote: nil, reference: types.StringUnknown(), expected: types.StringNull()},
		{name: "policy removed remotely", remote: nil, reference: types.StringValue("policy"), expected: types.StringNull()},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, securityPolicyState(c.remote, c.reference))
		})
	}
}

func TestSetBaseResourceState(t *testing.T) {
	ctx := t.Context()

	t.Run("API values are written, tags stay as declared", func(t *testing.T) {
		var diags diag.Diagnostics

		state := testBaseResourceModel(t)
		state.ID = types.StringUnknown()
		state.Tags = stringMap(map[string]string{"env": "prod"})

		remote := testBaseResource()
		remote.Alias = stringPtr(testAlias)
		remote.IsVisible = boolPtr(false)
		remote.Tags = map[string]string{"env": "prod", "owner": "infra"}

		setBaseResourceState(ctx, &state, &state, remote, true, &diags)

		require.False(t, diags.HasError(), diags)
		assert.Equal(t, types.StringValue(testResourceID), state.ID)
		// the alias is always taken from the API, even when it is not configured
		assert.Equal(t, types.StringValue(testAlias), state.Alias)
		assert.Equal(t, types.BoolValue(false), state.IsVisible)
		assert.Equal(t, stringMap(map[string]string{"env": "prod"}), state.Tags)
		assert.Equal(t, stringMap(map[string]string{"env": "prod", "owner": "infra"}), state.TagsAll)
	})

	t.Run("separate reference decides what is kept", func(t *testing.T) {
		var diags diag.Diagnostics

		state := testBaseResourceModel(t)
		state.IsVisible = types.BoolNull()

		reference := testBaseResourceModel(t)
		reference.IsVisible = types.BoolUnknown()
		reference.SecurityPolicyID = types.StringValue("")
		reference.Tags = stringMap(map[string]string{"env": "prod"})

		setBaseResourceState(ctx, &state, &reference, testBaseResource(), true, &diags)

		require.False(t, diags.HasError(), diags)
		assert.Equal(t, types.BoolNull(), state.IsVisible)
		assert.Equal(t, types.StringValue(""), state.SecurityPolicyID)
		assert.Equal(t, reference.Tags, state.Tags)
	})

	t.Run("non-authoritative access ignores external assignments", func(t *testing.T) {
		var diags diag.Diagnostics

		state := testBaseResourceModel(t)
		state.accessModel = makeTestAccessModel(ctx, types.BoolValue(false), []string{"g1"}, []string{"s1"})
		expected := state.accessModel

		remote := testBaseResource()
		remote.GroupsAccess = testGroups("g1", "g-external")
		remote.ServiceAccounts = []string{"s1", "s-external"}

		setBaseResourceState(ctx, &state, &state, remote, false, &diags)

		require.False(t, diags.HasError(), diags)
		assert.Equal(t, expected, state.accessModel)
	})

	t.Run("authoritative access shows external assignments", func(t *testing.T) {
		var diags diag.Diagnostics

		state := testBaseResourceModel(t)
		state.accessModel = makeTestAccessModel(ctx, types.BoolValue(true), []string{"g1"}, nil)

		remote := testBaseResource()
		remote.GroupsAccess = testGroups("g1", "g-external")
		remote.ServiceAccounts = []string{"s-external"}

		setBaseResourceState(ctx, &state, &state, remote, true, &diags)

		require.False(t, diags.HasError(), diags)
		assertAccessState(t, state.accessModel, true, []string{"g1", "g-external"}, []string{"s-external"})
	})
}

func TestSetPAMResourceState(t *testing.T) {
	var diags diag.Diagnostics

	// State written before `is_authoritative` existed holds null: it defaults to authoritative.
	state := pamResourceModel{baseResourceModel: testBaseResourceModel(t), GatewayID: types.StringUnknown()}
	state.IsAuthoritative = types.BoolNull()

	remote := model.PAMResource{BaseResource: testBaseResource(), GatewayID: testGatewayID}
	remote.ServiceAccounts = []string{"s-external"}

	setPAMResourceState(t.Context(), &state, remote, &diags)

	require.False(t, diags.HasError(), diags)
	assert.Equal(t, types.StringValue(testGatewayID), state.GatewayID)
	assertAccessState(t, state.accessModel, true, nil, []string{"s-external"})
}

// Every resource kind exposes the same base attributes and access blocks.
func TestResourcesShareBaseSchema(t *testing.T) {
	resources := map[string]resource.Resource{
		TwingateResource:           &twingateResource{},
		TwingateKubernetesResource: &kubernetesResource{},
		TwingateSSHResource:        &sshResource{},
		TwingateWebAppResource:     &webAppResource{},
	}

	reference := resourceSchema(t, resources[TwingateResource])

	for name, res := range resources {
		t.Run(name, func(t *testing.T) {
			resSchema := resourceSchema(t, res)

			for _, attribute := range []string{attr.IsVisible, attr.Alias, attr.SecurityPolicyID, attr.Tags, attr.TagsAll, attr.IsAuthoritative} {
				assert.Equal(t, reference.Attributes[attribute], resSchema.Attributes[attribute], attribute)
			}

			for _, attribute := range []string{attr.ID, attr.Name, attr.Address, attr.RemoteNetworkID} {
				assert.Contains(t, resSchema.Attributes, attribute)
			}

			assert.Equal(t, reference.Blocks, resSchema.Blocks)
		})
	}
}

// The PAM resources follow the Network resource on read: the alias always comes
// from the API and an explicitly configured empty security policy is kept.
func TestPAMResourceReadFollowsReferenceBehaviour(t *testing.T) {
	for _, c := range pamResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			response := strings.Replace(c.readResponse(t, nil, nil), `"alias":""`, `"alias":"`+testAlias+`"`, 1)
			require.Contains(t, response, testAlias)

			apiClient, _ := newMockedClient(t, response)
			res := c.newResource(apiClient)

			stateModel := c.model(t, types.StringValue(testResourceID), types.BoolValue(true), nil, nil)
			pamModelOf(t, stateModel).SecurityPolicyID = types.StringValue("")

			resp := &resource.ReadResponse{State: stateOf(t, res, stateModel)}
			res.Read(t.Context(), resource.ReadRequest{State: stateOf(t, res, stateModel)}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			state := c.emptyModel()
			readStateInto(t, resp.State, state)

			pam := pamModelOf(t, state)
			assert.Equal(t, types.StringValue(testAlias), pam.Alias)
			assert.Equal(t, types.StringValue(""), pam.SecurityPolicyID)
			assert.Equal(t, types.StringValue(testGatewayID), pam.GatewayID)
		})
	}
}

func TestIsBaseResourceChanged(t *testing.T) {
	ctx := t.Context()

	cases := []struct {
		name    string
		change  func(plan *baseResourceModel)
		changed bool
	}{
		{name: "nothing", change: func(*baseResourceModel) {}},
		{name: "name", change: func(plan *baseResourceModel) { plan.Name = types.StringValue("new") }, changed: true},
		{name: "address", change: func(plan *baseResourceModel) { plan.Address = types.StringValue("new.example.com") }, changed: true},
		{name: "remote network", change: func(plan *baseResourceModel) { plan.RemoteNetworkID = types.StringValue("new") }, changed: true},
		{name: "visibility", change: func(plan *baseResourceModel) { plan.IsVisible = types.BoolValue(false) }, changed: true},
		{name: "alias", change: func(plan *baseResourceModel) { plan.Alias = types.StringValue(testAlias) }, changed: true},
		{name: "security policy", change: func(plan *baseResourceModel) { plan.SecurityPolicyID = types.StringValue("policy") }, changed: true},
		{name: "tags", change: func(plan *baseResourceModel) { plan.Tags = stringMap(map[string]string{"env": "prod"}) }, changed: true},
		{name: "default tags", change: func(plan *baseResourceModel) { plan.TagsAll = stringMap(map[string]string{"env": "prod"}) }, changed: true},
		{
			name: "access policy",
			change: func(plan *baseResourceModel) {
				plan.AccessPolicy = types.SetValueMust(types.ObjectType{AttrTypes: accessPolicyAttributeTypes()}, nil)
			},
			changed: true,
		},
		// group and service account access is updated separately
		{name: "access groups", change: func(plan *baseResourceModel) { plan.GroupAccess = makeTestGroupAccess(ctx, "g1") }},
		{name: "access services", change: func(plan *baseResourceModel) { plan.ServiceAccess = makeTestServiceAccess(ctx, "s1") }},
		{name: "authoritative flag", change: func(plan *baseResourceModel) { plan.IsAuthoritative = types.BoolValue(false) }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := testBaseResourceModel(t)
			plan := testBaseResourceModel(t)
			c.change(&plan)

			assert.Equal(t, c.changed, isBaseResourceChanged(&plan, &state))

			// the PAM check includes the base one
			pamState := pamResourceModel{baseResourceModel: state, GatewayID: types.StringValue(testGatewayID)}
			pamPlan := pamResourceModel{baseResourceModel: plan, GatewayID: types.StringValue(testGatewayID)}

			assert.Equal(t, c.changed, isPAMResourceChanged(&pamPlan, &pamState))
		})
	}
}

func TestIsPAMResourceChangedGateway(t *testing.T) {
	state := pamResourceModel{baseResourceModel: testBaseResourceModel(t), GatewayID: types.StringValue(testGatewayID)}
	plan := pamResourceModel{baseResourceModel: testBaseResourceModel(t), GatewayID: types.StringValue("new-gateway")}

	assert.True(t, isPAMResourceChanged(&plan, &state))
}

func TestIsSpecificResourceChanged(t *testing.T) {
	noTags := types.MapNull(types.StringType)

	t.Run(TwingateResource, func(t *testing.T) {
		state, plan := testResourceModel(t), testResourceModel(t)
		assert.False(t, isResourceChanged(&plan, &state))

		plan.RoutingMode = types.StringValue(model.RoutingModeBypassTwingate)
		assert.True(t, isResourceChanged(&plan, &state))

		plan = testResourceModel(t)
		plan.IsActive = types.BoolValue(!state.IsActive.ValueBool())
		assert.True(t, isResourceChanged(&plan, &state))

		plan = testResourceModel(t)
		plan.Name = types.StringValue("new")
		assert.True(t, isResourceChanged(&plan, &state))
	})

	t.Run(TwingateKubernetesResource, func(t *testing.T) {
		state, plan := testKubernetesModel(t, noTags), testKubernetesModel(t, noTags)
		state.Address = types.StringValue(defaultKubernetesAddress)

		assert.False(t, isKubernetesResourceChanged(&plan, &state, defaultKubernetesAddress))
		// the address is compared as it is sent, not as it is planned
		assert.True(t, isKubernetesResourceChanged(&plan, &state, testK8sCustomAddress))

		// credential files are not sent to the API
		plan.BearerTokenFile = types.StringValue(testK8sBearerToken)
		plan.CAFile = types.StringValue(testK8sCAFile)
		assert.False(t, isKubernetesResourceChanged(&plan, &state, defaultKubernetesAddress))

		plan.GatewayID = types.StringValue("new-gateway")
		assert.True(t, isKubernetesResourceChanged(&plan, &state, defaultKubernetesAddress))
	})

	t.Run(TwingateSSHResource, func(t *testing.T) {
		state, plan := testSSHModel(t, noTags), testSSHModel(t, noTags)
		assert.False(t, isSSHResourceChanged(&plan, &state))

		plan.Upstream = sshPortObjectOf(t, 2222)
		assert.True(t, isSSHResourceChanged(&plan, &state))

		plan = testSSHModel(t, noTags)
		plan.Downstream = sshPortObjectOf(t, 2222)
		assert.True(t, isSSHResourceChanged(&plan, &state))
	})

	t.Run(TwingateWebAppResource, func(t *testing.T) {
		state, plan := testWebAppModel(t, noTags, noTags), testWebAppModel(t, noTags, noTags)
		assert.False(t, isWebAppResourceChanged(&plan, &state))

		plan.RequestHeaderRewrites = stringMap(map[string]string{"x-forwarded-host": "app.internal"})
		assert.True(t, isWebAppResourceChanged(&plan, &state))

		plan = testWebAppModel(t, noTags, noTags)
		plan.Upstream = types.ObjectNull(webAppUpstreamAttributeTypes)
		assert.True(t, isWebAppResourceChanged(&plan, &state))

		plan = testWebAppModel(t, noTags, noTags)
		plan.Downstream = types.ObjectNull(webAppDownstreamAttributeTypes)
		assert.True(t, isWebAppResourceChanged(&plan, &state))
	})
}
