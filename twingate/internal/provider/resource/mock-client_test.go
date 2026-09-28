package resource

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

const (
	testResourceID      = "resource-id"
	testResourceName    = "resource-name"
	testResourceAddress = "resource.example.com"
	testGatewayID       = "gateway-id"
	testRemoteNetworkID = "remote-network-id"
)

// newMockedClient returns an API client served by httpmock: every request gets
// responseBody back, and the returned pointer captures the last request body.
func newMockedClient(t *testing.T, responseBody string) (*client.Client, *string) {
	t.Helper()

	apiClient := client.NewClient(t.Context(), "https://test.twindev.com", "xxxx",
		time.Second, 0, client.DefaultAgent, "test", client.CacheOptions{})

	httpmock.ActivateNonDefault(apiClient.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)

	var requestBody string

	httpmock.RegisterResponder(http.MethodPost, apiClient.GraphqlServerURL, func(req *http.Request) (*http.Response, error) {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		requestBody = string(raw)

		return httpmock.NewStringResponse(http.StatusOK, responseBody), nil
	})

	return apiClient, &requestBody
}

// requestVariable extracts one GraphQL variable from a captured request body.
func requestVariable(t *testing.T, requestBody, name string) any {
	t.Helper()

	var request struct {
		Variables map[string]any `json:"variables"`
	}

	require.NoError(t, json.Unmarshal([]byte(requestBody), &request), requestBody)

	value, ok := request.Variables[name]
	require.True(t, ok, "variable %q missing from request: %s", name, requestBody)

	return value
}

// keyValueInputs is the wire form of a string map: a list of {key, value}
// objects, empty rather than absent for a nil map.
func keyValueInputs(pairs map[string]string) []any {
	inputs := make([]any, 0, len(pairs))

	for key, value := range pairs {
		inputs = append(inputs, map[string]any{"key": key, "value": value})
	}

	return inputs
}

// keyValueJSON renders pairs the way the API returns them in a response body.
func keyValueJSON(t *testing.T, pairs map[string]string) string {
	t.Helper()

	raw, err := json.Marshal(keyValueInputs(pairs))
	require.NoError(t, err)

	return string(raw)
}

func resourceSchema(t *testing.T, res resource.Resource) schema.Schema {
	t.Helper()

	resp := &resource.SchemaResponse{}
	res.Schema(t.Context(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	return resp.Schema
}

func planOf(t *testing.T, res resource.Resource, model any) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: resourceSchema(t, res)}
	diags := plan.Set(t.Context(), model)
	require.False(t, diags.HasError(), diags)

	return plan
}

func stateOf(t *testing.T, res resource.Resource, model any) tfsdk.State {
	t.Helper()

	state := tfsdk.State{Schema: resourceSchema(t, res)}
	diags := state.Set(t.Context(), model)
	require.False(t, diags.HasError(), diags)

	return state
}

// nullStateOf is the empty state the framework hands to Create.
func nullStateOf(t *testing.T, res resource.Resource) tfsdk.State {
	t.Helper()

	resSchema := resourceSchema(t, res)

	return tfsdk.State{
		Schema: resSchema,
		Raw:    tftypes.NewValue(resSchema.Type().TerraformType(t.Context()), nil),
	}
}

func readStateInto(t *testing.T, state tfsdk.State, model any) {
	t.Helper()

	diags := state.Get(t.Context(), model)
	require.False(t, diags.HasError(), diags)
}
