package client

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestReadWebAppResources(t *testing.T) {
	cases := []struct {
		name         string
		responseBody string
		expected     []*model.WebAppResource
		expectedErr  bool
	}{
		{
			name:         "empty edges - returns empty, no error",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[]}}}`,
			expected:     []*model.WebAppResource{},
		},
		{
			name: "only web app resources - all returned",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[
				{"node":{"__typename":"WebAppResource","id":"web-1","name":"web-resource-1"}},
				{"node":{"__typename":"WebAppResource","id":"web-2","name":"web-resource-2"}}
			]}}}`,
			expected: []*model.WebAppResource{
				{ID: "web-1", Name: "web-resource-1"},
				{ID: "web-2", Name: "web-resource-2"},
			},
		},
		{
			name: "mixed types - only web app resources returned",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[
				{"node":{"__typename":"WebAppResource","id":"web-1","name":"web-resource-1"}},
				{"node":{"__typename":"SSHResource","id":"ssh-1","name":"ssh-resource-1"}},
				{"node":{"__typename":"NetworkResource","id":"net-1","name":"network-resource-1"}}
			]}}}`,
			expected: []*model.WebAppResource{
				{ID: "web-1", Name: "web-resource-1"},
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

			resources, err := client.ReadWebAppResources(context.Background())

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

func TestReadWebAppResource(t *testing.T) {
	const resourceNode = `{
		"id": "web-1",
		"name": "web-resource",
		"address": {"value": "internal.acme.com"},
		"remoteNetwork": {"id": "rn-1"},
		"gateway": {"id": "gw-1"},
		"isVisible": true,
		"alias": "app.int",
		"securityPolicy": {"id": "sp-1", "name": "policy"},
		"tags": [{"key": "env", "value": "prod"}],
		"approvalMode": "MANUAL",
		"accessPolicy": {"mode": "APPROVAL_REQUIRED", "durationSeconds": 3600},
		"access": {"pageInfo": {"endCursor": "", "hasNextPage": false}, "edges": []},
		"upstream": {"port": 8080, "tls": true},
		"downstream": {"port": 8443, "tls": true},
		"requestHeaderRewrites": [{"key": "x-user", "value": "{{username}}"}]
	}`

	cases := []struct {
		name         string
		resourceID   string
		responseBody string
		expected     *model.WebAppResource
		expectedErr  bool
	}{
		{
			name:         "resource found - mapped to model",
			resourceID:   "web-1",
			responseBody: `{"data":{"resource":` + resourceNode + `}}`,
			expected:     expectedWebAppResource(),
		},
		{
			name:        "empty id - error before request",
			resourceID:  "",
			expectedErr: true,
		},
		{
			name:         "resource missing - error",
			resourceID:   "web-1",
			responseBody: `{"data":{"resource":null}}`,
			expectedErr:  true,
		},
		{
			name:         "graphql error - error propagated",
			resourceID:   "web-1",
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

			resource, err := client.ReadWebAppResource(context.Background(), c.resourceID)

			if c.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, resource)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, resource)
			}
		})
	}
}

// webAppEntity is the mutation payload the API returns for a web app resource.
const webAppEntity = `{
	"id": "web-1",
	"name": "web-resource",
	"address": {"value": "internal.acme.com"},
	"remoteNetwork": {"id": "rn-1"},
	"gateway": {"id": "gw-1"},
	"isVisible": true,
	"alias": "app.int",
	"securityPolicy": {"id": "sp-1", "name": "policy"},
	"tags": [{"key": "env", "value": "prod"}],
	"approvalMode": "MANUAL",
	"accessPolicy": {"mode": "APPROVAL_REQUIRED", "durationSeconds": 3600},
	"upstream": {"port": 8080, "tls": true},
	"downstream": {"port": 8443, "tls": true},
	"requestHeaderRewrites": [{"key": "x-user", "value": "{{username}}"}]
}`

func expectedWebAppResource() *model.WebAppResource {
	isVisible := true
	alias := "app.int"
	securityPolicyID := "sp-1"
	approvalMode := "MANUAL"
	accessMode := "APPROVAL_REQUIRED"
	duration := "1h"

	return &model.WebAppResource{
		ID:                    "web-1",
		Name:                  "web-resource",
		Address:               "internal.acme.com",
		GatewayID:             "gw-1",
		RemoteNetworkID:       "rn-1",
		IsVisible:             &isVisible,
		Alias:                 &alias,
		SecurityPolicyID:      &securityPolicyID,
		Tags:                  map[string]string{"env": "prod"},
		Upstream:              model.WebAppUpstream{Port: 8080, TLS: true},
		Downstream:            model.WebAppDownstream{Port: 8443, TLS: true},
		RequestHeaderRewrites: map[string]string{"x-user": "{{username}}"},
		AccessPolicy: &model.AccessPolicy{
			Mode:         &accessMode,
			Duration:     &duration,
			ApprovalMode: &approvalMode,
		},
	}
}

func webAppInput() *model.WebAppResource {
	return &model.WebAppResource{
		Name:            "web-resource",
		Address:         "internal.acme.com",
		GatewayID:       "gw-1",
		RemoteNetworkID: "rn-1",
		Upstream:        model.WebAppUpstream{Port: 8080, TLS: true},
		Downstream:      model.WebAppDownstream{Port: 8443, TLS: true},
	}
}

func TestCreateWebAppResource(t *testing.T) {
	cases := []struct {
		name         string
		responseBody string
		expected     *model.WebAppResource
		expectedErr  bool
	}{
		{
			name:         "entity returned - mapped to model",
			responseBody: `{"data":{"webAppResourceCreate":{"ok":true,"error":null,"entity":` + webAppEntity + `}}}`,
			expected:     expectedWebAppResource(),
		},
		{
			name:         "mutation not ok - error",
			responseBody: `{"data":{"webAppResourceCreate":{"ok":false,"error":"name already taken","entity":null}}}`,
			expectedErr:  true,
		},
		{
			name:         "empty entity - error",
			responseBody: `{"data":{"webAppResourceCreate":{"ok":true,"error":null,"entity":null}}}`,
			expectedErr:  true,
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

			resource, err := client.CreateWebAppResource(context.Background(), webAppInput())

			if c.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, resource)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, resource)
			}
		})
	}
}

func TestUpdateWebAppResource(t *testing.T) {
	cases := []struct {
		name         string
		resourceID   string
		responseBody string
		expected     *model.WebAppResource
		expectedErr  bool
	}{
		{
			name:         "entity returned - mapped to model",
			resourceID:   "web-1",
			responseBody: `{"data":{"webAppResourceUpdate":{"ok":true,"error":null,"entity":` + webAppEntity + `}}}`,
			expected:     expectedWebAppResource(),
		},
		{
			name:        "empty id - error before request",
			resourceID:  "",
			expectedErr: true,
		},
		{
			name:         "mutation not ok - error",
			resourceID:   "web-1",
			responseBody: `{"data":{"webAppResourceUpdate":{"ok":false,"error":"gateway not found","entity":null}}}`,
			expectedErr:  true,
		},
		{
			name:         "empty entity - error",
			resourceID:   "web-1",
			responseBody: `{"data":{"webAppResourceUpdate":{"ok":true,"error":null,"entity":null}}}`,
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

			input := webAppInput()
			input.ID = c.resourceID

			resource, err := client.UpdateWebAppResource(context.Background(), input)

			if c.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, resource)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, resource)
			}
		})
	}
}

func TestDeleteWebAppResource(t *testing.T) {
	cases := []struct {
		name         string
		resourceID   string
		responseBody string
		expectedErr  bool
	}{
		{
			name:         "ok - no error",
			resourceID:   "web-1",
			responseBody: `{"data":{"resourceDelete":{"ok":true,"error":null}}}`,
		},
		{
			name:        "empty id - error before request",
			resourceID:  "",
			expectedErr: true,
		},
		{
			name:         "mutation not ok - error",
			resourceID:   "web-1",
			responseBody: `{"data":{"resourceDelete":{"ok":false,"error":"resource not found"}}}`,
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

			err := client.DeleteWebAppResource(context.Background(), c.resourceID)

			if c.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
