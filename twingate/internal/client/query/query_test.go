package query

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func TestReadConnectorQueryToModel(t *testing.T) {
	cases := []struct {
		query    ReadConnector
		expected *model.Connector
	}{
		{
			query:    ReadConnector{},
			expected: nil,
		},
		{
			query: ReadConnector{
				Connector: &gqlConnector{
					IDName: IDName{
						ID:   "connector-id",
						Name: "connector-name",
					},
					RemoteNetwork: struct {
						ID graphql.ID
					}{
						ID: "connector-network-id",
					},
				},
			},
			expected: &model.Connector{
				ID:        "connector-id",
				Name:      "connector-name",
				NetworkID: "connector-network-id",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestReadResourcesByNameQueryToModel(t *testing.T) {
	cases := []struct {
		query    ReadResourcesByName
		expected []*model.Resource
	}{
		{
			query:    ReadResourcesByName{},
			expected: []*model.Resource{},
		},
		{
			query: ReadResourcesByName{
				Resources: Resources{
					PaginatedResource: PaginatedResource[*ResourceEdge]{
						Edges: []*ResourceEdge{
							{
								Node: &ResourceNode{
									IDName: IDName{
										ID:   "resource-id",
										Name: "resource-name",
									},
									RemoteNetwork: struct {
										ID graphql.ID
									}{
										ID: "resource-network-id",
									},
								},
							},
						},
					},
				},
			},
			expected: []*model.Resource{
				{
					ID:              "resource-id",
					Name:            "resource-name",
					RemoteNetworkID: "resource-network-id",
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestReadConnectorsQueryToModel(t *testing.T) {
	cases := []struct {
		query    ReadConnectors
		expected []*model.Connector
	}{
		{
			query:    ReadConnectors{},
			expected: nil,
		},
		{
			query: ReadConnectors{
				Connectors: Connectors{
					PaginatedResource: PaginatedResource[*ConnectorEdge]{
						Edges: []*ConnectorEdge{
							{
								Node: &gqlConnector{
									IDName: IDName{
										ID:   "connector-id",
										Name: "connector-name",
									},
									RemoteNetwork: struct {
										ID graphql.ID
									}{
										ID: "connector-network-id",
									},
								},
							},
						},
					},
				},
			},
			expected: []*model.Connector{
				{
					ID:        "connector-id",
					Name:      "connector-name",
					NetworkID: "connector-network-id",
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}
