package resource

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testGroupID                 = "R3JvdXA6MQ=="
	testOtherGroupID            = "R3JvdXA6Mg=="
	testSecurityPolicyID        = "U2VjdXJpdHlQb2xpY3k6MQ=="
	testDefaultSecurityPolicyID = "U2VjdXJpdHlQb2xpY3k6MA=="
	testResourceAccessID        = testResourceID + resourceAccessIDSeparator + testGroupID
)

var resourceAccessMutationResponses = map[string]string{
	"resourceAccessAdd":    `{"data":{"resourceAccessAdd":{"ok":true,"error":null}}}`,
	"resourceAccessRemove": `{"data":{"resourceAccessRemove":{"ok":true,"error":null}}}`,
}

func newMockedResourceAccess(t *testing.T, readResponse string) (*resourceAccess, *[]string) {
	t.Helper()

	apiClient, requests := newRoutedMockedClient(t, readResponse, resourceAccessMutationResponses)

	return &resourceAccess{client: apiClient}, requests
}

func accessReadResponse(edges ...string) string {
	return fmt.Sprintf(`{"data":{"resource":{"id":%q,"access":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[%s]}}}}`,
		testResourceID, strings.Join(edges, ","))
}

func groupEdge(groupID, securityPolicyID, accessPolicy string) string {
	if accessPolicy == "" {
		accessPolicy = `null`
	}

	return fmt.Sprintf(`{"node":{"__typename":"Group","id":%q},"securityPolicy":{"id":%q,"name":"policy"},"approvalMode":"AUTOMATIC","accessPolicy":%s}`,
		groupID, securityPolicyID, accessPolicy)
}

func accessPolicySet(t *testing.T, mode, approvalMode, duration string) types.Set {
	t.Helper()

	value := func(val string) types.String {
		if val == "" {
			return types.StringNull()
		}

		return types.StringValue(val)
	}

	obj, diags := types.ObjectValue(accessPolicyAttributeTypes(), map[string]attr.Value{
		"mode":          value(mode),
		"approval_mode": value(approvalMode),
		"duration":      value(duration),
	})
	require.False(t, diags.HasError(), diags)

	set, diags := makeObjectsSet(t.Context(), obj)
	require.False(t, diags.HasError(), diags)

	return set
}

func nullAccessPolicySet(t *testing.T) types.Set {
	t.Helper()

	return makeObjectsSetNull(t.Context(), accessPolicyAttributeTypes())
}

func testResourceAccessModel(t *testing.T, securityPolicyID types.String, accessPolicy types.Set) resourceAccessModel {
	t.Helper()

	return resourceAccessModel{
		ID:               types.StringUnknown(),
		ResourceID:       types.StringValue(testResourceID),
		GroupID:          types.StringValue(testGroupID),
		SecurityPolicyID: securityPolicyID,
		AccessPolicy:     accessPolicy,
	}
}

func accessInput(t *testing.T, requestBody string) map[string]any {
	t.Helper()

	access, ok := requestVariable(t, requestBody, "access").([]any)
	require.True(t, ok, requestBody)
	require.Len(t, access, 1)

	input, ok := access[0].(map[string]any)
	require.True(t, ok, requestBody)

	return input
}

func TestResourceAccessCreate(t *testing.T) {
	cases := []struct {
		name                     string
		plan                     resourceAccessModel
		apiEdge                  string
		expectedSecurityPolicyID any
		expectedApprovalMode     any
		expectedAccessPolicy     map[string]any
		expectedState            resourceAccessModel
	}{
		{
			name:                     "security policy and access policy are sent",
			plan:                     testResourceAccessModel(t, types.StringValue(testSecurityPolicyID), accessPolicySet(t, "AUTO_LOCK", "AUTOMATIC", "30d")),
			apiEdge:                  groupEdge(testGroupID, testSecurityPolicyID, `{"mode":"AUTO_LOCK","durationSeconds":2592000}`),
			expectedSecurityPolicyID: testSecurityPolicyID,
			expectedApprovalMode:     "AUTOMATIC",
			expectedAccessPolicy:     map[string]any{"mode": "AUTO_LOCK", "durationSeconds": float64(2592000)},
			expectedState:            testResourceAccessModel(t, types.StringValue(testSecurityPolicyID), accessPolicySet(t, "AUTO_LOCK", "AUTOMATIC", "30d")),
		},
		{
			name:                     "omitted security policy stays null in state",
			plan:                     testResourceAccessModel(t, types.StringNull(), nullAccessPolicySet(t)),
			apiEdge:                  groupEdge(testGroupID, testDefaultSecurityPolicyID, `{"mode":"MANUAL","durationSeconds":null}`),
			expectedSecurityPolicyID: nil,
			expectedApprovalMode:     nil,
			expectedAccessPolicy:     map[string]any{"mode": "MANUAL", "durationSeconds": nil},
			expectedState:            testResourceAccessModel(t, types.StringNull(), nullAccessPolicySet(t)),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, requests := newMockedResourceAccess(t, accessReadResponse(
				c.apiEdge,
				groupEdge(testOtherGroupID, testDefaultSecurityPolicyID, ""),
			))

			resp := &resource.CreateResponse{State: nullStateOf(t, res)}
			res.Create(t.Context(), resource.CreateRequest{Plan: planOf(t, res, c.plan)}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			assert.Equal(t, map[string]any{
				"principalId":      testGroupID,
				"securityPolicyId": c.expectedSecurityPolicyID,
				"approvalMode":     c.expectedApprovalMode,
				"accessPolicy":     c.expectedAccessPolicy,
			}, accessInput(t, (*requests)[0]))

			var state resourceAccessModel

			readStateInto(t, resp.State, &state)

			c.expectedState.ID = types.StringValue(testResourceAccessID)
			assert.Equal(t, c.expectedState, state)
		})
	}
}

func TestResourceAccessCreateRejectsInvalidGroupID(t *testing.T) {
	cases := []struct {
		name    string
		groupID string
	}{
		{
			name:    "not a global id",
			groupID: "not-a-global-id",
		},
		{
			name:    "service account id",
			groupID: "U2VydmljZUFjY291bnQ6MQ==",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, requests := newMockedResourceAccess(t, accessReadResponse())

			plan := testResourceAccessModel(t, types.StringNull(), nullAccessPolicySet(t))
			plan.GroupID = types.StringValue(c.groupID)

			resp := &resource.CreateResponse{State: nullStateOf(t, res)}
			res.Create(t.Context(), resource.CreateRequest{Plan: planOf(t, res, plan)}, resp)

			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), ErrResourceAccessNotGroupID.Error())
			assert.Empty(t, *requests)
		})
	}
}

func TestResourceAccessCreateFailsWhenGroupMissingAfterWrite(t *testing.T) {
	res, requests := newMockedResourceAccess(t, accessReadResponse(
		groupEdge(testOtherGroupID, testDefaultSecurityPolicyID, ""),
	))

	resp := &resource.CreateResponse{State: nullStateOf(t, res)}
	res.Create(t.Context(), resource.CreateRequest{
		Plan: planOf(t, res, testResourceAccessModel(t, types.StringNull(), nullAccessPolicySet(t))),
	}, resp)

	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), ErrResourceAccessNotFound.Error())
	assert.Len(t, *requests, 2)
}

func TestResourceAccessUpdate(t *testing.T) {
	cases := []struct {
		name                     string
		state                    resourceAccessModel
		plan                     resourceAccessModel
		apiEdge                  string
		expectedSecurityPolicyID any
		expectedState            resourceAccessModel
	}{
		{
			name:                     "security policy and access policy change in place",
			state:                    testResourceAccessModel(t, types.StringValue(testDefaultSecurityPolicyID), nullAccessPolicySet(t)),
			plan:                     testResourceAccessModel(t, types.StringValue(testSecurityPolicyID), accessPolicySet(t, "AUTO_LOCK", "AUTOMATIC", "2d")),
			apiEdge:                  groupEdge(testGroupID, testSecurityPolicyID, `{"mode":"AUTO_LOCK","durationSeconds":172800}`),
			expectedSecurityPolicyID: testSecurityPolicyID,
			expectedState:            testResourceAccessModel(t, types.StringValue(testSecurityPolicyID), accessPolicySet(t, "AUTO_LOCK", "AUTOMATIC", "2d")),
		},
		{
			name:                     "removed security policy is sent as null",
			state:                    testResourceAccessModel(t, types.StringValue(testSecurityPolicyID), nullAccessPolicySet(t)),
			plan:                     testResourceAccessModel(t, types.StringNull(), nullAccessPolicySet(t)),
			apiEdge:                  groupEdge(testGroupID, testDefaultSecurityPolicyID, ""),
			expectedSecurityPolicyID: nil,
			expectedState:            testResourceAccessModel(t, types.StringNull(), nullAccessPolicySet(t)),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, requests := newMockedResourceAccess(t, accessReadResponse(c.apiEdge))

			c.state.ID = types.StringValue(testResourceAccessID)
			c.plan.ID = types.StringValue(testResourceAccessID)

			resp := &resource.UpdateResponse{State: stateOf(t, res, c.state)}
			res.Update(t.Context(), resource.UpdateRequest{
				Plan:  planOf(t, res, c.plan),
				State: stateOf(t, res, c.state),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			assert.Equal(t, testResourceID, requestVariable(t, (*requests)[0], "id"))
			assert.Equal(t, c.expectedSecurityPolicyID, accessInput(t, (*requests)[0])["securityPolicyId"])

			var state resourceAccessModel

			readStateInto(t, resp.State, &state)

			c.expectedState.ID = types.StringValue(testResourceAccessID)
			assert.Equal(t, c.expectedState, state)
		})
	}
}

func TestResourceAccessRead(t *testing.T) {
	cases := []struct {
		name                     string
		response                 string
		expectRemoved            bool
		expectedSecurityPolicyID string
	}{
		{
			name:                     "security policy changed outside terraform",
			response:                 accessReadResponse(groupEdge(testGroupID, testDefaultSecurityPolicyID, "")),
			expectedSecurityPolicyID: testDefaultSecurityPolicyID,
		},
		{
			name:          "group no longer has access",
			response:      accessReadResponse(groupEdge(testOtherGroupID, testSecurityPolicyID, "")),
			expectRemoved: true,
		},
		{
			name:          "resource deleted",
			response:      `{"data":{"resource":null}}`,
			expectRemoved: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, _ := newMockedResourceAccess(t, c.response)

			current := testResourceAccessModel(t, types.StringValue(testSecurityPolicyID), nullAccessPolicySet(t))
			current.ID = types.StringValue(testResourceAccessID)

			resp := &resource.ReadResponse{State: stateOf(t, res, current)}
			res.Read(t.Context(), resource.ReadRequest{State: stateOf(t, res, current)}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			if c.expectRemoved {
				assert.True(t, resp.State.Raw.IsNull())

				return
			}

			var state resourceAccessModel

			readStateInto(t, resp.State, &state)
			assert.Equal(t, types.StringValue(c.expectedSecurityPolicyID), state.SecurityPolicyID)
		})
	}
}

func TestResourceAccessDelete(t *testing.T) {
	res, requests := newMockedResourceAccess(t, accessReadResponse())

	current := testResourceAccessModel(t, types.StringNull(), nullAccessPolicySet(t))
	current.ID = types.StringValue(testResourceAccessID)

	resp := &resource.DeleteResponse{State: stateOf(t, res, current)}
	res.Delete(t.Context(), resource.DeleteRequest{State: stateOf(t, res, current)}, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Len(t, *requests, 1)
	assert.Equal(t, testResourceID, requestVariable(t, (*requests)[0], "id"))
	assert.Equal(t, []any{testGroupID}, requestVariable(t, (*requests)[0], "principalIds"))
}

func TestResourceAccessImportState(t *testing.T) {
	cases := []struct {
		name          string
		id            string
		expectError   bool
		expectedState resourceAccessModel
	}{
		{
			name: "group with access",
			id:   testResourceAccessID,
			expectedState: resourceAccessModel{
				ID:               types.StringValue(testResourceAccessID),
				ResourceID:       types.StringValue(testResourceID),
				GroupID:          types.StringValue(testGroupID),
				SecurityPolicyID: types.StringValue(testSecurityPolicyID),
				AccessPolicy:     accessPolicySet(t, "AUTO_LOCK", "AUTOMATIC", "1h"),
			},
		},
		{
			name:        "group without access",
			id:          testResourceID + resourceAccessIDSeparator + testOtherGroupID,
			expectError: true,
		},
		{
			name:        "missing separator",
			id:          testResourceID,
			expectError: true,
		},
		{
			name:        "missing group id",
			id:          testResourceID + resourceAccessIDSeparator,
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, _ := newMockedResourceAccess(t, accessReadResponse(
				groupEdge(testGroupID, testSecurityPolicyID, `{"mode":"AUTO_LOCK","durationSeconds":3600}`),
			))

			resp := &resource.ImportStateResponse{State: nullStateOf(t, res)}
			res.ImportState(t.Context(), resource.ImportStateRequest{ID: c.id}, resp)

			if c.expectError {
				assert.True(t, resp.Diagnostics.HasError())

				return
			}

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			var state resourceAccessModel

			readStateInto(t, resp.State, &state)
			assert.Equal(t, c.expectedState, state)
		})
	}
}
