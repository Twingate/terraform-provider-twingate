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
	var (
		boolTrue  = true
		boolFalse = false
	)

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
									IsVisible:                true,
									IsBrowserShortcutEnabled: false,
								},
							},
						},
					},
				},
			},
			expected: []*model.Resource{
				{
					ID:                       "resource-id",
					Name:                     "resource-name",
					RemoteNetworkID:          "resource-network-id",
					IsVisible:                &boolTrue,
					IsBrowserShortcutEnabled: &boolFalse,
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

func optionalString(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

func optionalBool(val bool) *bool {
	return &val
}

func TestBuildGroupsFilter(t *testing.T) {
	defaultActive := BooleanFilterOperatorInput{Eq: true}
	defaultType := GroupTypeFilterOperatorInput{
		In: []graphql.String{model.GroupTypeManual,
			model.GroupTypeSynced,
			model.GroupTypeSystem},
	}

	testCases := []struct {
		filter   *model.GroupsFilter
		expected *GroupFilterInput
	}{
		{
			filter:   nil,
			expected: nil,
		},
		{
			filter: &model.GroupsFilter{Name: optionalString("Group")},
			expected: &GroupFilterInput{
				Name: &StringFilterOperationInput{
					Eq: "Group",
				},
				Type:     defaultType,
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{Type: optionalString("MANUAL")},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []graphql.String{model.GroupTypeManual},
				},
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{Type: optionalString("SYSTEM")},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []graphql.String{model.GroupTypeSystem},
				},
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{Type: optionalString("SYNCED")},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []graphql.String{model.GroupTypeSynced},
				},
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{IsActive: optionalBool(true)},
			expected: &GroupFilterInput{
				Type:     defaultType,
				IsActive: BooleanFilterOperatorInput{Eq: true},
			},
		},
		{
			filter: &model.GroupsFilter{IsActive: optionalBool(false)},
			expected: &GroupFilterInput{
				Type:     defaultType,
				IsActive: BooleanFilterOperatorInput{Eq: false},
			},
		},
		{
			filter: &model.GroupsFilter{
				Type:     optionalString("SYSTEM"),
				IsActive: optionalBool(false),
			},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []graphql.String{model.GroupTypeSystem},
				},
				IsActive: BooleanFilterOperatorInput{Eq: false},
			},
		},
		{
			filter: &model.GroupsFilter{
				Type:     optionalString("MANUAL"),
				IsActive: optionalBool(true),
			},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []graphql.String{model.GroupTypeManual},
				},
				IsActive: BooleanFilterOperatorInput{Eq: true},
			},
		},
		{
			filter: &model.GroupsFilter{
				Type:     optionalString("MANUAL"),
				IsActive: optionalBool(false),
			},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []graphql.String{model.GroupTypeManual},
				},
				IsActive: BooleanFilterOperatorInput{Eq: false},
			},
		},
	}

	for n, td := range testCases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, td.expected, NewGroupFilterInput(td.filter))
		})
	}
}
