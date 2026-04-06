package client

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
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
				{ID: "k8s-1", Name: "k8s-resource-1"},
				{ID: "k8s-2", Name: "k8s-resource-2"},
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
				{ID: "k8s-1", Name: "k8s-resource-1"},
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
			client := newTestClient()
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

func TestConvertGroupsToAccessInput(t *testing.T) {
	cases := []struct {
		name     string
		groups   []model.AccessGroup
		expected []AccessInput
	}{
		{
			name:     "nil groups - returns empty slice",
			groups:   nil,
			expected: []AccessInput{},
		},
		{
			name:     "empty groups - returns empty slice",
			groups:   []model.AccessGroup{},
			expected: []AccessInput{},
		},
		{
			name: "group without security policy or access policy",
			groups: []model.AccessGroup{
				{GroupID: "group-1"},
			},
			expected: []AccessInput{
				{PrincipalID: "group-1"},
			},
		},
		{
			name: "group with security policy",
			groups: []model.AccessGroup{
				{GroupID: "group-1", SecurityPolicyID: strPtr("sp-1")},
			},
			expected: []AccessInput{
				{PrincipalID: "group-1", SecurityPolicyID: strPtr("sp-1")},
			},
		},
		{
			name: "multiple groups",
			groups: []model.AccessGroup{
				{GroupID: "group-1"},
				{GroupID: "group-2", SecurityPolicyID: strPtr("sp-2")},
			},
			expected: []AccessInput{
				{PrincipalID: "group-1"},
				{PrincipalID: "group-2", SecurityPolicyID: strPtr("sp-2")},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := convertGroupsToAccessInput(c.groups)
			assert.Equal(t, len(c.expected), len(actual))
			for i := range c.expected {
				assert.Equal(t, c.expected[i].PrincipalID, actual[i].PrincipalID)
				assert.Equal(t, c.expected[i].SecurityPolicyID, actual[i].SecurityPolicyID)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}