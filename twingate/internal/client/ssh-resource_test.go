package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sshEntity is the mutation payload the API returns for an SSH resource.
const sshEntity = `{
	"id": "ssh-1",
	"name": "ssh-resource",
	"address": {"value": "10.0.0.1"},
	"remoteNetwork": {"id": "rn-1"},
	"gateway": {"id": "gw-1"},
	"isVisible": true,
	"alias": "",
	"securityPolicy": null,
	"tags": [{"key": "env", "value": "prod"}],
	"approvalMode": "",
	"accessPolicy": null,
	"downstream": {"port": 2222},
	"upstream": {"port": 22}
}`

func expectedSSHResource() *model.SSHResource {
	isVisible := true

	return &model.SSHResource{
		ID:              "ssh-1",
		Name:            "ssh-resource",
		Address:         "10.0.0.1",
		GatewayID:       "gw-1",
		RemoteNetworkID: "rn-1",
		IsVisible:       &isVisible,
		Tags:            map[string]string{"env": "prod"},
		Downstream:      &model.SSHDownstream{Port: 2222},
		Upstream:        &model.SSHUpstream{Port: 22},
	}
}

// capturingResponder answers with body and records the GraphQL variables of the
// last request.
func capturingResponder(t *testing.T, body string, variables *map[string]any) httpmock.Responder {
	t.Helper()

	return func(req *http.Request) (*http.Response, error) {
		var request struct {
			Variables map[string]any `json:"variables"`
		}

		require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
		*variables = request.Variables

		return httpmock.NewStringResponse(200, body), nil
	}
}

// An unset port block is sent as an explicit null so the API applies its default.
func TestSSHResourcePortVariables(t *testing.T) {
	cases := []struct {
		name               string
		input              *model.SSHResource
		expectedDownstream any
		expectedUpstream   any
	}{
		{
			name:               "ports unset - null is sent",
			input:              &model.SSHResource{ID: "ssh-1", Name: "ssh-resource", Address: "10.0.0.1", GatewayID: "gw-1", RemoteNetworkID: "rn-1"},
			expectedDownstream: nil,
			expectedUpstream:   nil,
		},
		{
			name: "ports set - port inputs are sent",
			input: &model.SSHResource{
				ID: "ssh-1", Name: "ssh-resource", Address: "10.0.0.1", GatewayID: "gw-1", RemoteNetworkID: "rn-1",
				Downstream: &model.SSHDownstream{Port: 2222},
				Upstream:   &model.SSHUpstream{Port: 22},
			},
			expectedDownstream: map[string]any{"port": float64(2222)},
			expectedUpstream:   map[string]any{"port": float64(22)},
		},
	}

	for _, c := range cases {
		t.Run("create: "+c.name, func(t *testing.T) {
			client := newTestClient(t.Context())
			httpmock.ActivateNonDefault(client.HTTPClient)
			defer httpmock.DeactivateAndReset()

			var variables map[string]any

			httpmock.RegisterResponder("POST", client.GraphqlServerURL,
				capturingResponder(t, `{"data":{"sshResourceCreate":{"ok":true,"error":null,"entity":`+sshEntity+`}}}`, &variables))

			resource, err := client.CreateSSHResource(context.Background(), c.input)

			require.NoError(t, err)
			assert.Equal(t, expectedSSHResource(), resource)
			assert.Equal(t, c.expectedDownstream, variables["downstream"])
			assert.Equal(t, c.expectedUpstream, variables["upstream"])
		})

		t.Run("update: "+c.name, func(t *testing.T) {
			client := newTestClient(t.Context())
			httpmock.ActivateNonDefault(client.HTTPClient)
			defer httpmock.DeactivateAndReset()

			var variables map[string]any

			httpmock.RegisterResponder("POST", client.GraphqlServerURL,
				capturingResponder(t, `{"data":{"sshResourceUpdate":{"ok":true,"error":null,"entity":`+sshEntity+`}}}`, &variables))

			resource, err := client.UpdateSSHResource(context.Background(), c.input)

			require.NoError(t, err)
			assert.Equal(t, expectedSSHResource(), resource)
			assert.Equal(t, c.expectedDownstream, variables["downstream"])
			assert.Equal(t, c.expectedUpstream, variables["upstream"])
		})
	}
}

func TestReadSSHResource(t *testing.T) {
	const resourceNode = `{
		"id": "ssh-1",
		"name": "ssh-resource",
		"address": {"value": "10.0.0.1"},
		"remoteNetwork": {"id": "rn-1"},
		"gateway": {"id": "gw-1"},
		"isVisible": true,
		"alias": "",
		"securityPolicy": null,
		"tags": [{"key": "env", "value": "prod"}],
		"approvalMode": "",
		"accessPolicy": null,
		"access": {"pageInfo": {"endCursor": "", "hasNextPage": false}, "edges": []},
		"downstream": {"port": 2222},
		"upstream": {"port": 22}
	}`

	client := newTestClient(t.Context())
	httpmock.ActivateNonDefault(client.HTTPClient)
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", client.GraphqlServerURL,
		httpmock.NewStringResponder(200, `{"data":{"resource":`+resourceNode+`}}`))

	resource, err := client.ReadSSHResource(context.Background(), "ssh-1")

	require.NoError(t, err)
	assert.Equal(t, expectedSSHResource(), resource)
}

func TestReadSSHResources(t *testing.T) {
	cases := []struct {
		name         string
		responseBody string
		expected     []*model.SSHResource
		expectedErr  bool
	}{
		{
			name:         "empty edges - returns empty, no error",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[]}}}`,
			expected:     []*model.SSHResource{},
		},
		{
			name: "only SSH resources - all returned",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[
				{"node":{"__typename":"SSHResource","id":"ssh-1","name":"ssh-resource-1"}},
				{"node":{"__typename":"SSHResource","id":"ssh-2","name":"ssh-resource-2"}}
			]}}}`,
			expected: []*model.SSHResource{
				{ID: "ssh-1", Name: "ssh-resource-1"},
				{ID: "ssh-2", Name: "ssh-resource-2"},
			},
		},
		{
			name: "mixed types - only SSH resources returned",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[
				{"node":{"__typename":"SSHResource","id":"ssh-1","name":"ssh-resource-1"}},
				{"node":{"__typename":"KubernetesResource","id":"k8s-1","name":"k8s-resource-1"}},
				{"node":{"__typename":"NetworkResource","id":"net-1","name":"network-resource-1"}}
			]}}}`,
			expected: []*model.SSHResource{
				{ID: "ssh-1", Name: "ssh-resource-1"},
			},
		},
		{
			name:         "graphql error - error propagated",
			responseBody: `{"errors":[{"message":"server error","locations":[{"line":1,"column":1}]}]}`,
			expectedErr:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := newTestClient(t.Context())
			httpmock.ActivateNonDefault(client.HTTPClient)
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder("POST", client.GraphqlServerURL,
				httpmock.NewStringResponder(200, c.responseBody))

			resources, err := client.ReadSSHResources(context.Background())

			if c.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, resources)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, resources)
			}
		})
	}
}
