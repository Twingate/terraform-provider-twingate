package resource

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccessResourceID    = "UmVzb3VyY2U6MQ=="
	testAccessGroupID       = "R3JvdXA6MQ=="
	testAccessOtherGroupID  = "R3JvdXA6Mg=="
	testAccessPolicyID      = "U2VjdXJpdHlQb2xpY3k6MQ=="
	testAccessDefaultPolicy = "U2VjdXJpdHlQb2xpY3k6MA=="
)

type resourceAccessAPI struct {
	readResponse string
	mutations    []string
}

func newResourceAccessMockedClient(t *testing.T, readResponse string) (*client.Client, *resourceAccessAPI) {
	t.Helper()

	apiClient := client.NewClient(t.Context(), "https://test.twindev.com", "xxxx",
		time.Second, 0, client.DefaultAgent, "test", client.CacheOptions{})

	httpmock.ActivateNonDefault(apiClient.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)

	api := &resourceAccessAPI{readResponse: readResponse}

	httpmock.RegisterResponder(http.MethodPost, apiClient.GraphqlServerURL, func(req *http.Request) (*http.Response, error) {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		body := string(raw)

		switch {
		case strings.Contains(body, "resourceAccessAdd"):
			api.mutations = append(api.mutations, body)

			return httpmock.NewStringResponse(http.StatusOK, `{"data":{"resourceAccessAdd":{"ok":true,"error":null}}}`), nil
		case strings.Contains(body, "resourceAccessRemove"):
			api.mutations = append(api.mutations, body)

			return httpmock.NewStringResponse(http.StatusOK, `{"data":{"resourceAccessRemove":{"ok":true,"error":null}}}`), nil
		default:
			return httpmock.NewStringResponse(http.StatusOK, api.readResponse), nil
		}
	})

	return apiClient, api
}

func accessReadResponse(edges ...string) string {
	return fmt.Sprintf(`{"data":{"resource":{"id":%q,"access":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[%s]}}}}`,
		testAccessResourceID, strings.Join(edges, ","))
}

func groupEdge(groupID, policyID string, accessPolicy string) string {
	if accessPolicy == "" {
		accessPolicy = `null`
	}

	return fmt.Sprintf(`{"node":{"__typename":"Group","id":%q},"securityPolicy":{"id":%q,"name":"policy"},"approvalMode":"AUTOMATIC","accessPolicy":%s}`,
		groupID, policyID, accessPolicy)
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

func groupAccessModel(t *testing.T, securityPolicyID types.String, accessPolicy types.Set) resourceAccessModel {
	t.Helper()

	return resourceAccessModel{
		ID:               types.StringUnknown(),
		ResourceID:       types.StringValue(testAccessResourceID),
		GroupID:          types.StringValue(testAccessGroupID),
		SecurityPolicyID: securityPolicyID,
		AccessPolicy:     accessPolicy,
	}
}

func createResourceAccess(t *testing.T, apiClient *client.Client, plan resourceAccessModel) (resourceAccessModel, *resource.CreateResponse) {
	t.Helper()

	res := &resourceAccess{client: apiClient}
	req := resource.CreateRequest{Plan: planOf(t, res, &plan)}
	resp := &resource.CreateResponse{State: nullStateOf(t, res)}

	res.Create(t.Context(), req, resp)

	var state resourceAccessModel
	if !resp.Diagnostics.HasError() && !resp.State.Raw.IsNull() {
		readStateInto(t, resp.State, &state)
	}

	return state, resp
}

func TestResourceAccessCreateGroup(t *testing.T) {
	apiClient, api := newResourceAccessMockedClient(t, accessReadResponse(
		groupEdge(testAccessGroupID, testAccessPolicyID, `{"mode":"AUTO_LOCK","durationSeconds":2592000}`),
		groupEdge(testAccessOtherGroupID, testAccessDefaultPolicy, ""),
	))

	plan := groupAccessModel(t, types.StringValue(testAccessPolicyID), accessPolicySet(t, "AUTO_LOCK", "AUTOMATIC", "30d"))

	state, resp := createResourceAccess(t, apiClient, plan)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	require.Len(t, api.mutations, 1)
	access := requestVariable(t, api.mutations[0], "access").([]any)
	require.Len(t, access, 1)
	assert.Equal(t, map[string]any{
		"principalId":      testAccessGroupID,
		"securityPolicyId": testAccessPolicyID,
		"approvalMode":     "AUTOMATIC",
		"accessPolicy":     map[string]any{"mode": "AUTO_LOCK", "durationSeconds": float64(2592000)},
	}, access[0])

	assert.Equal(t, testAccessResourceID+":"+testAccessGroupID, state.ID.ValueString())
	assert.Equal(t, testAccessPolicyID, state.SecurityPolicyID.ValueString())
	assert.True(t, state.AccessPolicy.Equal(accessPolicySet(t, "AUTO_LOCK", "AUTOMATIC", "30d")), state.AccessPolicy.String())
}

func TestResourceAccessCreateGroupWithoutSecurityPolicy(t *testing.T) {
	apiClient, api := newResourceAccessMockedClient(t, accessReadResponse(
		groupEdge(testAccessGroupID, testAccessDefaultPolicy, `{"mode":"MANUAL","durationSeconds":null}`),
	))

	state, resp := createResourceAccess(t, apiClient, groupAccessModel(t, types.StringNull(), nullAccessPolicySet(t)))
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	access := requestVariable(t, api.mutations[0], "access").([]any)
	assert.Nil(t, access[0].(map[string]any)["securityPolicyId"])

	assert.True(t, state.SecurityPolicyID.IsNull())
	assert.True(t, state.AccessPolicy.IsNull())
}

func TestResourceAccessCreateRejectsInvalidGroupID(t *testing.T) {
	cases := map[string]string{
		"not a global id":    "not-a-global-id",
		"service account id": "U2VydmljZUFjY291bnQ6MQ==",
	}

	for name, groupID := range cases {
		t.Run(name, func(t *testing.T) {
			apiClient, api := newResourceAccessMockedClient(t, accessReadResponse())

			plan := groupAccessModel(t, types.StringNull(), nullAccessPolicySet(t))
			plan.GroupID = types.StringValue(groupID)

			_, resp := createResourceAccess(t, apiClient, plan)
			require.True(t, resp.Diagnostics.HasError())
			assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), ErrResourceAccessNotGroupID.Error())
			assert.Empty(t, api.mutations)
		})
	}
}

func TestResourceAccessCreateFailsWhenGroupMissingAfterWrite(t *testing.T) {
	apiClient, api := newResourceAccessMockedClient(t, accessReadResponse(
		groupEdge(testAccessOtherGroupID, testAccessDefaultPolicy, ""),
	))

	_, resp := createResourceAccess(t, apiClient, groupAccessModel(t, types.StringNull(), nullAccessPolicySet(t)))
	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), ErrResourceAccessNotFound.Error())
	assert.Len(t, api.mutations, 1)
}

func readResourceAccess(t *testing.T, apiClient *client.Client, current resourceAccessModel) tfsdk.State {
	t.Helper()

	res := &resourceAccess{client: apiClient}
	resp := &resource.ReadResponse{State: stateOf(t, res, &current)}

	res.Read(t.Context(), resource.ReadRequest{State: stateOf(t, res, &current)}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	return resp.State
}

func TestResourceAccessRead(t *testing.T) {
	current := groupAccessModel(t, types.StringValue(testAccessPolicyID), nullAccessPolicySet(t))
	current.ID = types.StringValue(testAccessResourceID + ":" + testAccessGroupID)

	cases := []struct {
		name          string
		response      string
		expectRemoved bool
		expectPolicy  string
	}{
		{
			name:         "security policy changed outside terraform",
			response:     accessReadResponse(groupEdge(testAccessGroupID, testAccessDefaultPolicy, "")),
			expectPolicy: testAccessDefaultPolicy,
		},
		{
			name:          "group no longer has access",
			response:      accessReadResponse(groupEdge(testAccessOtherGroupID, testAccessPolicyID, "")),
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
			apiClient, _ := newResourceAccessMockedClient(t, c.response)

			state := readResourceAccess(t, apiClient, current)

			if c.expectRemoved {
				assert.True(t, state.Raw.IsNull())

				return
			}

			var got resourceAccessModel
			readStateInto(t, state, &got)
			assert.Equal(t, c.expectPolicy, got.SecurityPolicyID.ValueString())
		})
	}
}

func TestResourceAccessDelete(t *testing.T) {
	apiClient, api := newResourceAccessMockedClient(t, accessReadResponse())

	current := groupAccessModel(t, types.StringNull(), nullAccessPolicySet(t))
	current.ID = types.StringValue(testAccessResourceID + ":" + testAccessGroupID)

	res := &resourceAccess{client: apiClient}
	resp := &resource.DeleteResponse{State: stateOf(t, res, &current)}
	res.Delete(t.Context(), resource.DeleteRequest{State: stateOf(t, res, &current)}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	require.Len(t, api.mutations, 1)
	assert.Equal(t, testAccessResourceID, requestVariable(t, api.mutations[0], "id"))
	assert.Equal(t, []any{testAccessGroupID}, requestVariable(t, api.mutations[0], "principalIds"))
}

func TestResourceAccessImport(t *testing.T) {
	cases := []struct {
		name        string
		id          string
		expectError bool
		check       func(t *testing.T, state resourceAccessModel)
	}{
		{
			name: "group",
			id:   testAccessResourceID + ":" + testAccessGroupID,
			check: func(t *testing.T, state resourceAccessModel) {
				assert.Equal(t, testAccessGroupID, state.GroupID.ValueString())
				assert.Equal(t, testAccessPolicyID, state.SecurityPolicyID.ValueString())
				assert.True(t, state.AccessPolicy.Equal(accessPolicySet(t, "AUTO_LOCK", "AUTOMATIC", "1h")), state.AccessPolicy.String())
			},
		},
		{name: "principal without access", id: testAccessResourceID + ":" + testAccessOtherGroupID, expectError: true},
		{name: "missing separator", id: testAccessResourceID, expectError: true},
		{name: "missing principal", id: testAccessResourceID + ":", expectError: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			apiClient, _ := newResourceAccessMockedClient(t, accessReadResponse(
				groupEdge(testAccessGroupID, testAccessPolicyID, `{"mode":"AUTO_LOCK","durationSeconds":3600}`),
			))

			res := &resourceAccess{client: apiClient}
			resp := &resource.ImportStateResponse{State: nullStateOf(t, res)}
			res.ImportState(t.Context(), resource.ImportStateRequest{ID: c.id}, resp)

			if c.expectError {
				assert.True(t, resp.Diagnostics.HasError())

				return
			}

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			var state resourceAccessModel
			readStateInto(t, resp.State, &state)
			assert.Equal(t, c.id, state.ID.ValueString())
			assert.Equal(t, testAccessResourceID, state.ResourceID.ValueString())
			c.check(t, state)
		})
	}
}
