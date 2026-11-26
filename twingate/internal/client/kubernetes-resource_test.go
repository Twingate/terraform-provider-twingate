package client

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestReadKubernetesResources(t *testing.T) {
	cases := []struct {
		name         string
		responseBody string
		expected     []*model.KubernetesResource
		expectedErr  bool
	}{
		{
			name:         "empty edges - returns empty, no error",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[]}}}`,
			expected:     []*model.KubernetesResource{},
		},
		{
			name: "only Kubernetes resources - all returned",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[
				{"node":{"__typename":"KubernetesResource","id":"k8s-1","name":"k8s-resource-1"}},
				{"node":{"__typename":"KubernetesResource","id":"k8s-2","name":"k8s-resource-2"}}
			]}}}`,
			expected: []*model.KubernetesResource{
				{
					PAMResource: model.PAMResource{
						BaseResource: model.BaseResource{
							ID:   "k8s-1",
							Name: "k8s-resource-1",
						},
					},
				},
				{
					PAMResource: model.PAMResource{
						BaseResource: model.BaseResource{
							ID:   "k8s-2",
							Name: "k8s-resource-2",
						},
					},
				},
			},
		},
		{
			name: "mixed types - only Kubernetes resources returned",
			responseBody: `{"data":{"resources":{"pageInfo":{"endCursor":"","hasNextPage":false},"edges":[
				{"node":{"__typename":"SSHResource","id":"ssh-1","name":"ssh-resource-1"}},
				{"node":{"__typename":"KubernetesResource","id":"k8s-1","name":"k8s-resource-1"}},
				{"node":{"__typename":"NetworkResource","id":"net-1","name":"network-resource-1"}}
			]}}}`,
			expected: []*model.KubernetesResource{
				{
					PAMResource: model.PAMResource{
						BaseResource: model.BaseResource{
							ID:   "k8s-1",
							Name: "k8s-resource-1",
						},
					},
				},
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

			resources, err := client.ReadKubernetesResources(context.Background())

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
