package resource

import (
	"context"
	"fmt"
	"testing"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func emptyStringMap() types.Map {
	return types.MapValueMust(types.StringType, map[string]tfattr.Value{})
}

func stringMap(pairs map[string]string) types.Map {
	elements := make(map[string]tfattr.Value, len(pairs))
	for key, value := range pairs {
		elements[key] = types.StringValue(value)
	}

	return types.MapValueMust(types.StringType, elements)
}

// getKeyValueMap feeds the client, which sends an empty list for a nil map.
// Both null and empty must reach the API as "clear the stored rewrites".
func TestGetKeyValueMap(t *testing.T) {
	cases := []struct {
		name     string
		input    types.Map
		expected map[string]string
	}{
		{
			name:     "null map - nil",
			input:    types.MapNull(types.StringType),
			expected: nil,
		},
		{
			name:     "unknown map - nil",
			input:    types.MapUnknown(types.StringType),
			expected: nil,
		},
		{
			name:     "empty map - nil",
			input:    emptyStringMap(),
			expected: nil,
		},
		{
			name:     "populated map - converted",
			input:    stringMap(map[string]string{"x-a": "1", "x-b": "2"}),
			expected: map[string]string{"x-a": "1", "x-b": "2"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, getKeyValueMap(c.input))
		})
	}
}

func TestWebAppUpstreamRoundTrip(t *testing.T) {
	ctx := context.Background()

	obj, diags := webAppUpstreamObject(ctx, 8080)
	require.False(t, diags.HasError())

	upstream, diags := webAppUpstreamValue(ctx, obj)
	require.False(t, diags.HasError())

	assert.Equal(t, int64(8080), upstream.Port.ValueInt64())
}

func TestWebAppDownstreamRoundTrip(t *testing.T) {
	ctx := context.Background()

	obj, diags := webAppDownstreamObject(ctx, 80)
	require.False(t, diags.HasError())

	downstream, diags := webAppDownstreamValue(ctx, obj)
	require.False(t, diags.HasError())

	assert.Equal(t, int64(80), downstream.Port.ValueInt64())
}

func testWebAppModel(t *testing.T, tags, headerRewrites types.Map) webAppResourceModel {
	t.Helper()

	upstream, diags := webAppUpstreamObject(t.Context(), 8080)
	require.False(t, diags.HasError(), diags)

	downstream, diags := webAppDownstreamObject(t.Context(), 80)
	require.False(t, diags.HasError(), diags)

	return webAppResourceModel{
		ID:                    types.StringUnknown(),
		Name:                  types.StringValue(testResourceName),
		Address:               types.StringValue(testResourceAddress),
		GatewayID:             types.StringValue(testGatewayID),
		RemoteNetworkID:       types.StringValue(testRemoteNetworkID),
		IsVisible:             types.BoolValue(true),
		Alias:                 types.StringNull(),
		SecurityPolicyID:      types.StringNull(),
		Tags:                  tags,
		TagsAll:               plannedTagsAll(tags),
		Upstream:              upstream,
		Downstream:            downstream,
		RequestHeaderRewrites: headerRewrites,
		AccessPolicy:          makeObjectsSetNull(t.Context(), accessPolicyAttributeTypes()),
		GroupAccess:           makeObjectsSetNull(t.Context(), accessGroupAttributeTypes()),
	}
}

// webAppEntityResponse is a successful create or update mutation payload whose
// entity echoes the given tags and header rewrites.
func webAppEntityResponse(t *testing.T, mutation string, tags, headerRewrites map[string]string) string {
	t.Helper()

	return fmt.Sprintf(`{"data":{%q:{"ok":true,"error":null,"entity":{"id":%q,"name":%q,"address":{"value":%q},`+
		`"remoteNetwork":{"id":%q},"gateway":{"id":%q},"isVisible":true,"alias":"","securityPolicy":null,`+
		`"tags":%s,"approvalMode":"","accessPolicy":null,"upstream":{"port":8080},"downstream":{"port":80},`+
		`"requestHeaderRewrites":%s}}}}`,
		mutation, testResourceID, testResourceName, testResourceAddress, testRemoteNetworkID, testGatewayID,
		keyValueJSON(t, tags), keyValueJSON(t, headerRewrites))
}

// Tags and request header rewrites share one converter, so both must reach the
// API as key-value lists on create and update: an omitted map is sent as an
// empty list, which clears the stored entries, and a declared map round-trips
// back into state.
func TestWebAppResourceKeyValueMaps(t *testing.T) {
	cases := []struct {
		name                   string
		planTags               types.Map
		planHeaderRewrites     types.Map
		apiTags                map[string]string
		apiHeaderRewrites      map[string]string
		expectedTags           types.Map
		expectedHeaderRewrites types.Map
	}{
		{
			name:                   "no maps - empty lists are sent",
			planTags:               types.MapNull(types.StringType),
			planHeaderRewrites:     types.MapNull(types.StringType),
			expectedTags:           types.MapNull(types.StringType),
			expectedHeaderRewrites: types.MapNull(types.StringType),
		},
		{
			name:                   "maps - key-value inputs are sent",
			planTags:               stringMap(map[string]string{"env": "prod"}),
			planHeaderRewrites:     stringMap(map[string]string{"x-forwarded-host": "app.internal"}),
			apiTags:                map[string]string{"env": "prod"},
			apiHeaderRewrites:      map[string]string{"x-forwarded-host": "app.internal"},
			expectedTags:           stringMap(map[string]string{"env": "prod"}),
			expectedHeaderRewrites: stringMap(map[string]string{"x-forwarded-host": "app.internal"}),
		},
		{
			name:                   "empty maps - state keeps the empty maps from the plan",
			planTags:               emptyStringMap(),
			planHeaderRewrites:     emptyStringMap(),
			expectedTags:           emptyStringMap(),
			expectedHeaderRewrites: emptyStringMap(),
		},
	}

	for _, c := range cases {
		t.Run("create: "+c.name, func(t *testing.T) {
			apiClient, requestBody := newMockedClient(t, webAppEntityResponse(t, "webAppResourceCreate", c.apiTags, c.apiHeaderRewrites))
			webAppRes := &webAppResource{client: apiClient}

			resp := &resource.CreateResponse{State: nullStateOf(t, webAppRes)}
			webAppRes.Create(t.Context(), resource.CreateRequest{
				Plan: planOf(t, webAppRes, testWebAppModel(t, c.planTags, c.planHeaderRewrites)),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			assert.ElementsMatch(t, keyValueInputs(c.apiTags), requestVariable(t, *requestBody, "tags"))
			assert.ElementsMatch(t, keyValueInputs(c.apiHeaderRewrites), requestVariable(t, *requestBody, "requestHeaderRewrites"))

			var state webAppResourceModel

			readStateInto(t, resp.State, &state)
			assert.Equal(t, types.StringValue(testResourceID), state.ID)
			assert.Equal(t, c.expectedTags, state.Tags)
			assert.Equal(t, c.expectedHeaderRewrites, state.RequestHeaderRewrites)
		})

		t.Run("update: "+c.name, func(t *testing.T) {
			apiClient, requestBody := newMockedClient(t, webAppEntityResponse(t, "webAppResourceUpdate", c.apiTags, c.apiHeaderRewrites))
			webAppRes := &webAppResource{client: apiClient}

			stateModel := testWebAppModel(t, stringMap(map[string]string{"stale": "tag"}), stringMap(map[string]string{"x-stale": "1"}))
			stateModel.ID = types.StringValue(testResourceID)

			planModel := testWebAppModel(t, c.planTags, c.planHeaderRewrites)
			planModel.ID = types.StringValue(testResourceID)

			resp := &resource.UpdateResponse{State: stateOf(t, webAppRes, stateModel)}
			webAppRes.Update(t.Context(), resource.UpdateRequest{
				Plan:  planOf(t, webAppRes, planModel),
				State: stateOf(t, webAppRes, stateModel),
			}, resp)

			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			assert.Equal(t, testResourceID, requestVariable(t, *requestBody, "id"))
			assert.ElementsMatch(t, keyValueInputs(c.apiTags), requestVariable(t, *requestBody, "tags"))
			assert.ElementsMatch(t, keyValueInputs(c.apiHeaderRewrites), requestVariable(t, *requestBody, "requestHeaderRewrites"))

			var state webAppResourceModel

			readStateInto(t, resp.State, &state)
			assert.Equal(t, c.expectedTags, state.Tags)
			assert.Equal(t, c.expectedHeaderRewrites, state.RequestHeaderRewrites)
		})
	}
}
