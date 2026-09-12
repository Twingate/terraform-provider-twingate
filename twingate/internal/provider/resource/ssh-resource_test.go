package resource

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSSHModel(t *testing.T, tags types.Map) sshResourceModel {
	t.Helper()

	return sshResourceModel{
		ID:               types.StringUnknown(),
		Name:             types.StringValue(testResourceName),
		Address:          types.StringValue(testResourceAddress),
		GatewayID:        types.StringValue(testGatewayID),
		RemoteNetworkID:  types.StringValue(testRemoteNetworkID),
		IsVisible:        types.BoolValue(true),
		Alias:            types.StringNull(),
		SecurityPolicyID: types.StringNull(),
		Tags:             tags,
		AccessPolicy:     makeObjectsSetNull(t.Context(), accessPolicyAttributeTypes()),
		GroupAccess:      makeObjectsSetNull(t.Context(), accessGroupAttributeTypes()),
	}
}

// sshEntityResponse is a successful create or update mutation payload whose
// entity echoes the given tags.
func sshEntityResponse(t *testing.T, mutation string, tags map[string]string) string {
	t.Helper()

	return fmt.Sprintf(`{"data":{%q:{"ok":true,"error":null,"entity":{"id":%q,"name":%q,"address":{"value":%q},`+
		`"remoteNetwork":{"id":%q},"gateway":{"id":%q},"isVisible":true,"alias":"","securityPolicy":null,`+
		`"tags":%s,"approvalMode":"","accessPolicy":null}}}}`,
		mutation, testResourceID, testResourceName, testResourceAddress, testRemoteNetworkID, testGatewayID,
		keyValueJSON(t, tags))
}

// Tags reach the API as a list of key-value inputs on both create and update: an
// omitted map is sent as an empty list, which clears stored tags, and whatever
// the API returns lands in state.
func TestSSHResourceTags(t *testing.T) {
	cases := []struct {
		name          string
		planTags      types.Map
		apiTags       map[string]string
		expectedState types.Map
	}{
		{
			name:          "no tags - empty list is sent",
			planTags:      types.MapNull(types.StringType),
			apiTags:       nil,
			expectedState: types.MapNull(types.StringType),
		},
		{
			name:          "tags - key-value inputs are sent",
			planTags:      stringMap(map[string]string{"env": "prod", "team": "infra"}),
			apiTags:       map[string]string{"env": "prod", "team": "infra"},
			expectedState: stringMap(map[string]string{"env": "prod", "team": "infra"}),
		},
	}

	for _, c := range cases {
		t.Run("create: "+c.name, func(t *testing.T) {
			apiClient, requestBody := newMockedClient(t, sshEntityResponse(t, "sshResourceCreate", c.apiTags))
			sshRes := &sshResource{client: apiClient}

			resp := &resource.CreateResponse{State: nullStateOf(t, sshRes)}
			sshRes.Create(t.Context(), resource.CreateRequest{
				Plan: planOf(t, sshRes, testSSHModel(t, c.planTags)),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			assert.ElementsMatch(t, keyValueInputs(c.apiTags), requestVariable(t, *requestBody, "tags"))

			var state sshResourceModel

			readStateInto(t, resp.State, &state)
			assert.Equal(t, types.StringValue(testResourceID), state.ID)
			assert.Equal(t, c.expectedState, state.Tags)
		})

		t.Run("update: "+c.name, func(t *testing.T) {
			apiClient, requestBody := newMockedClient(t, sshEntityResponse(t, "sshResourceUpdate", c.apiTags))
			sshRes := &sshResource{client: apiClient}

			stateModel := testSSHModel(t, stringMap(map[string]string{"stale": "tag"}))
			stateModel.ID = types.StringValue(testResourceID)

			planModel := testSSHModel(t, c.planTags)
			planModel.ID = types.StringValue(testResourceID)

			resp := &resource.UpdateResponse{State: stateOf(t, sshRes, stateModel)}
			sshRes.Update(t.Context(), resource.UpdateRequest{
				Plan:  planOf(t, sshRes, planModel),
				State: stateOf(t, sshRes, stateModel),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			assert.Equal(t, testResourceID, requestVariable(t, *requestBody, "id"))
			assert.ElementsMatch(t, keyValueInputs(c.apiTags), requestVariable(t, *requestBody, "tags"))

			var state sshResourceModel

			readStateInto(t, resp.State, &state)
			assert.Equal(t, c.expectedState, state.Tags)
		})
	}
}
