package client

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
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
