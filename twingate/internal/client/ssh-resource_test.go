package client

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

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
			name:        "graphql error - error propagated",
			responseBody: `{"errors":[{"message":"server error","locations":[{"line":1,"column":1}]}]}`,
			expectedErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := newTestClient()
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