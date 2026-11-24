package resource

import (
	"encoding/json"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	okResponse     = `{"data":{"resourceUpdate":{"ok":true,"error":null}}}`
	deleteResponse = `{"data":{"resourceDelete":{"ok":true,"error":null}}}`
)

// resourceEntity is the `twingate_resource` node the API returns, with the
// given principals assigned.
func resourceEntity(t *testing.T, groupIDs, serviceAccountIDs []string) string {
	t.Helper()

	edges := make([]any, 0, len(groupIDs)+len(serviceAccountIDs))

	for _, id := range groupIDs {
		edges = append(edges, map[string]any{"node": map[string]any{"__typename": "Group", "id": id}})
	}

	for _, id := range serviceAccountIDs {
		edges = append(edges, map[string]any{"node": map[string]any{"__typename": "ServiceAccount", "id": id}})
	}

	raw, err := json.Marshal(map[string]any{
		"id":                       testResourceID,
		"name":                     testResourceName,
		"address":                  map[string]any{"value": testResourceAddress},
		"remoteNetwork":            map[string]any{"id": testRemoteNetworkID},
		"isVisible":                true,
		"alias":                    "",
		"securityPolicy":           nil,
		"tags":                     []any{},
		"approvalMode":             model.ApprovalModeManual,
		"accessPolicy":             map[string]any{"mode": model.AccessPolicyModeManual, "durationSeconds": nil},
		"protocols":                nil,
		"isActive":                 true,
		"isBrowserShortcutEnabled": false,
		"routingMode":              "",
		"access": map[string]any{
			"pageInfo": map[string]any{"endCursor": "", "hasNextPage": false},
			"edges":    edges,
		},
	})
	require.NoError(t, err)

	return string(raw)
}

func resourceMutationResponse(t *testing.T, mutation string, groupIDs, serviceAccountIDs []string) string {
	t.Helper()

	return `{"data":{"` + mutation + `":{"ok":true,"error":null,"entity":` + resourceEntity(t, groupIDs, serviceAccountIDs) + `}}}`
}

func resourceReadResponse(t *testing.T, groupIDs, serviceAccountIDs []string) string {
	t.Helper()

	return `{"data":{"resource":` + resourceEntity(t, groupIDs, serviceAccountIDs) + `}}`
}

// testResourceModelWithAccess is a `twingate_resource` model holding the given access.
func testResourceModelWithAccess(t *testing.T, id types.String, isAuthoritative bool, groupIDs, serviceAccountIDs []string) resourceModel {
	t.Helper()

	res := testResourceModel(t)
	res.ID = id
	res.accessModel = makeTestAccessModel(t.Context(), types.BoolValue(isAuthoritative), groupIDs, serviceAccountIDs)

	return res
}

// The planned access is assigned right after `twingate_resource` is created
// and lands in state.
func TestResourceCreateAssignsAccess(t *testing.T) {
	group, serviceAccount := testGroupID("1"), testServiceAccountID("1")

	apiClient, requests := newSequencedMockedClient(t, resourceMutationResponse(t, "resourceCreate", nil, nil), accessAddResponse)
	res := &twingateResource{baseResource: baseResource{client: apiClient}}

	plan := testResourceModelWithAccess(t, types.StringUnknown(), true, []string{group}, []string{serviceAccount})

	resp := &resource.CreateResponse{State: nullStateOf(t, res)}
	res.Create(t.Context(), resource.CreateRequest{Plan: planOf(t, res, plan)}, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Len(t, *requests, 2)
	assert.Equal(t, []string{serviceAccount, group}, accessPrincipalIDs(t, (*requests)[1]))

	var state resourceModel

	readStateInto(t, resp.State, &state)
	assert.Equal(t, types.StringValue(testResourceID), state.ID)
	assert.Equal(t, types.BoolValue(true), state.IsActive)
	assertAccessState(t, state.accessModel, true, []string{group}, []string{serviceAccount})
}

// A resource planned as inactive is created and then deactivated.
func TestResourceCreateInactive(t *testing.T) {
	apiClient, requests := newSequencedMockedClient(t, resourceMutationResponse(t, "resourceCreate", nil, nil), okResponse)
	res := &twingateResource{baseResource: baseResource{client: apiClient}}

	plan := testResourceModel(t)
	plan.IsActive = types.BoolValue(false)

	resp := &resource.CreateResponse{State: nullStateOf(t, res)}
	res.Create(t.Context(), resource.CreateRequest{Plan: planOf(t, res, plan)}, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Len(t, *requests, 2)
	assert.Equal(t, false, requestVariable(t, (*requests)[1], "isActive"))

	var state resourceModel

	readStateInto(t, resp.State, &state)
	assert.Equal(t, types.BoolValue(false), state.IsActive)
}

func TestResourceCreateErrors(t *testing.T) {
	group := testGroupID("1")

	cases := []struct {
		name      string
		plan      func(t *testing.T) resourceModel
		responses []string
	}{
		{
			name: "invalid access ID - nothing is sent",
			plan: func(t *testing.T) resourceModel {
				t.Helper()

				return testResourceModelWithAccess(t, types.StringUnknown(), true, nil, []string{"not-a-global-id"})
			},
		},
		{
			name: "create fails",
			plan: func(t *testing.T) resourceModel {
				t.Helper()

				return testResourceModel(t)
			},
			responses: []string{graphqlErrorResponse},
		},
		{
			name: "assigning access fails",
			plan: func(t *testing.T) resourceModel {
				t.Helper()

				return testResourceModelWithAccess(t, types.StringUnknown(), true, []string{group}, nil)
			},
			responses: []string{resourceMutationResponse(t, "resourceCreate", nil, nil), graphqlErrorResponse},
		},
		{
			name: "deactivating fails",
			plan: func(t *testing.T) resourceModel {
				t.Helper()

				plan := testResourceModel(t)
				plan.IsActive = types.BoolValue(false)

				return plan
			},
			responses: []string{resourceMutationResponse(t, "resourceCreate", nil, nil), graphqlErrorResponse},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			apiClient, requests := newSequencedMockedClient(t, c.responses...)
			res := &twingateResource{baseResource: baseResource{client: apiClient}}

			resp := &resource.CreateResponse{State: nullStateOf(t, res)}
			res.Create(t.Context(), resource.CreateRequest{Plan: planOf(t, res, c.plan(t))}, resp)

			require.True(t, resp.Diagnostics.HasError())
			assert.Len(t, *requests, len(c.responses))
		})
	}
}

// An access-only change removes and adds principals and reads the resource
// back; the update mutation is not sent.
func TestResourceUpdateAccess(t *testing.T) {
	group, externalGroup := testGroupID("1"), testGroupID("external")
	oldService, newService := testServiceAccountID("old"), testServiceAccountID("new")

	cases := []struct {
		name            string
		isAuthoritative bool
		expectedRemoved []any
	}{
		{name: "authoritative - external assignments are removed", isAuthoritative: true, expectedRemoved: []any{oldService, externalGroup}},
		{name: "non-authoritative - external assignments are kept", isAuthoritative: false, expectedRemoved: []any{oldService}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			apiClient, requests := newSequencedMockedClient(t,
				resourceReadResponse(t, []string{group, externalGroup}, []string{oldService}),
				accessRemoveResponse,
				accessAddResponse,
				resourceReadResponse(t, []string{group}, []string{newService}),
			)
			res := &twingateResource{baseResource: baseResource{client: apiClient}}

			stateModel := testResourceModelWithAccess(t, types.StringValue(testResourceID), c.isAuthoritative, []string{group}, []string{oldService})
			planModel := testResourceModelWithAccess(t, types.StringValue(testResourceID), c.isAuthoritative, []string{group}, []string{newService})

			resp := &resource.UpdateResponse{State: stateOf(t, res, stateModel)}
			res.Update(t.Context(), resource.UpdateRequest{
				Plan:  planOf(t, res, planModel),
				State: stateOf(t, res, stateModel),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Len(t, *requests, 4)
			assert.NotContains(t, (*requests)[3], "mutation")
			assert.ElementsMatch(t, c.expectedRemoved, requestVariable(t, (*requests)[1], "principalIds"))
			assert.Equal(t, []string{newService}, accessPrincipalIDs(t, (*requests)[2]))

			var state resourceModel

			readStateInto(t, resp.State, &state)
			assertAccessState(t, state.accessModel, c.isAuthoritative, []string{group}, []string{newService})
		})
	}
}

// A change to the resource itself sends the update mutation.
func TestResourceUpdateSendsMutationWhenChanged(t *testing.T) {
	group := testGroupID("1")

	apiClient, requests := newSequencedMockedClient(t, resourceMutationResponse(t, "resourceUpdate", []string{group}, nil))
	res := &twingateResource{baseResource: baseResource{client: apiClient}}

	stateModel := testResourceModelWithAccess(t, types.StringValue(testResourceID), true, []string{group}, nil)
	stateModel.Name = types.StringValue("old-" + testResourceName)

	planModel := testResourceModelWithAccess(t, types.StringValue(testResourceID), true, []string{group}, nil)

	resp := &resource.UpdateResponse{State: stateOf(t, res, stateModel)}
	res.Update(t.Context(), resource.UpdateRequest{
		Plan:  planOf(t, res, planModel),
		State: stateOf(t, res, stateModel),
	}, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Len(t, *requests, 1)
	assert.Equal(t, testResourceName, requestVariable(t, (*requests)[0], "name"))

	var state resourceModel

	readStateInto(t, resp.State, &state)
	assert.Equal(t, types.StringValue(testResourceName), state.Name)
	assertAccessState(t, state.accessModel, true, []string{group}, nil)
}

// A failure while updating access stops the update before the resource is touched.
func TestResourceUpdateAccessFailure(t *testing.T) {
	apiClient, requests := newSequencedMockedClient(t, graphqlErrorResponse)
	res := &twingateResource{baseResource: baseResource{client: apiClient}}

	stateModel := testResourceModelWithAccess(t, types.StringValue(testResourceID), true, nil, []string{testServiceAccountID("old")})
	planModel := testResourceModelWithAccess(t, types.StringValue(testResourceID), true, nil, []string{testServiceAccountID("new")})

	resp := &resource.UpdateResponse{State: stateOf(t, res, stateModel)}
	res.Update(t.Context(), resource.UpdateRequest{
		Plan:  planOf(t, res, planModel),
		State: stateOf(t, res, stateModel),
	}, resp)

	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "failed to update resource access")
	assert.Len(t, *requests, 1)
}

func TestResourceDelete(t *testing.T) {
	t.Run("deletes by the ID in state", func(t *testing.T) {
		apiClient, requests := newSequencedMockedClient(t, deleteResponse)
		res := &twingateResource{baseResource: baseResource{client: apiClient}}

		stateModel := testResourceModelWithAccess(t, types.StringValue(testResourceID), true, nil, nil)

		resp := &resource.DeleteResponse{State: stateOf(t, res, stateModel)}
		res.Delete(t.Context(), resource.DeleteRequest{State: stateOf(t, res, stateModel)}, resp)

		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
		require.Len(t, *requests, 1)
		assert.Equal(t, testResourceID, requestVariable(t, (*requests)[0], "id"))
	})

	t.Run("API failure is reported", func(t *testing.T) {
		apiClient, _ := newSequencedMockedClient(t, graphqlErrorResponse)
		res := &twingateResource{baseResource: baseResource{client: apiClient}}

		stateModel := testResourceModelWithAccess(t, types.StringValue(testResourceID), true, nil, nil)

		resp := &resource.DeleteResponse{State: stateOf(t, res, stateModel)}
		res.Delete(t.Context(), resource.DeleteRequest{State: stateOf(t, res, stateModel)}, resp)

		require.True(t, resp.Diagnostics.HasError())
	})
}

func TestPAMResourceDelete(t *testing.T) {
	for _, c := range pamResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			apiClient, requests := newSequencedMockedClient(t, deleteResponse)
			res := c.newResource(apiClient)

			stateModel := c.model(t, types.StringValue(testResourceID), types.BoolValue(true), nil, nil)

			resp := &resource.DeleteResponse{State: stateOf(t, res, stateModel)}
			res.Delete(t.Context(), resource.DeleteRequest{State: stateOf(t, res, stateModel)}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Len(t, *requests, 1)
			assert.Equal(t, testResourceID, requestVariable(t, (*requests)[0], "id"))
		})
	}
}

// A failure while updating access stops the update before the resource is touched.
func TestPAMResourceUpdateAccessFailure(t *testing.T) {
	for _, c := range pamResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			apiClient, requests := newSequencedMockedClient(t, graphqlErrorResponse)
			res := c.newResource(apiClient)

			stateModel := c.model(t, types.StringValue(testResourceID), types.BoolValue(true), nil, []string{testServiceAccountID("old")})
			planModel := c.model(t, types.StringValue(testResourceID), types.BoolValue(true), nil, []string{testServiceAccountID("new")})

			resp := &resource.UpdateResponse{State: stateOf(t, res, stateModel)}
			res.Update(t.Context(), resource.UpdateRequest{
				Plan:  planOf(t, res, planModel),
				State: stateOf(t, res, stateModel),
			}, resp)

			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "failed to update resource access")
			assert.Len(t, *requests, 1)
		})
	}
}

// An invalid access ID in the plan fails the update before anything is sent.
func TestPAMResourceUpdateRejectsInvalidAccessIDs(t *testing.T) {
	for _, c := range pamResourceCases() {
		t.Run(c.name, func(t *testing.T) {
			apiClient, requests := newSequencedMockedClient(t)
			res := c.newResource(apiClient)

			stateModel := c.model(t, types.StringValue(testResourceID), types.BoolValue(true), nil, nil)
			planModel := c.model(t, types.StringValue(testResourceID), types.BoolValue(true), nil, []string{"not-a-global-id"})

			resp := &resource.UpdateResponse{State: stateOf(t, res, stateModel)}
			res.Update(t.Context(), resource.UpdateRequest{
				Plan:  planOf(t, res, planModel),
				State: stateOf(t, res, stateModel),
			}, resp)

			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), ErrWrongGlobalID.Error())
			assert.Empty(t, *requests)
		})
	}
}
