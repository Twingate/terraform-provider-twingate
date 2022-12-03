package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func Test_idToString(t *testing.T) {
	cases := []struct {
		id       graphql.ID
		expected string
	}{
		{
			id:       nil,
			expected: "",
		},
		{
			id:       graphql.ID("123"),
			expected: "123",
		},
		{
			id:       graphql.ID(101),
			expected: "101",
		},
		{
			id:       graphql.ID(101.5),
			expected: "101.5",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("test case #%d", i+1), func(t *testing.T) {
			actual := idToString(c.id)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func Test_convertToGQL(t *testing.T) {
	cases := []struct {
		val      interface{}
		expected interface{}
	}{
		{
			val:      nil,
			expected: nil,
		},
		{
			val:      "123",
			expected: graphql.String("123"),
		},
		{
			val:      101,
			expected: graphql.Int(101),
		},
		{
			val:      int32(102),
			expected: graphql.Int(102),
		},
		{
			val:      int64(103),
			expected: graphql.Int(103),
		},
		{
			val:      101.5,
			expected: graphql.Float(101.5),
		},
		{
			val:      float32(101.25),
			expected: graphql.Float(101.25),
		},
		{
			val:      true,
			expected: graphql.Boolean(true),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("test case #%d", i+1), func(t *testing.T) {
			actual := convertToGQL(c.val)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadConnectorQueryToModel(t *testing.T) {
	cases := []struct {
		query    readConnectorQuery
		expected *model.Connector
	}{
		{
			query:    readConnectorQuery{},
			expected: nil,
		},
		{
			query: readConnectorQuery{
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
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestReadConnectorsQueryToModel(t *testing.T) {
	cases := []struct {
		query    readConnectorsQuery
		expected []*model.Connector
	}{
		{
			query:    readConnectorsQuery{},
			expected: nil,
		},
		{
			query: readConnectorsQuery{
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
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestReadResourcesByNameQueryToModel(t *testing.T) {
	cases := []struct {
		query    readResourcesByNameQuery
		expected []*model.Resource
	}{
		{
			query:    readResourcesByNameQuery{},
			expected: []*model.Resource{},
		},
		{
			query: readResourcesByNameQuery{
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
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestGqlRemoteNetworksToModel(t *testing.T) {
	cases := []struct {
		query    gqlRemoteNetworks
		expected []*model.RemoteNetwork
	}{
		{
			query:    gqlRemoteNetworks{},
			expected: []*model.RemoteNetwork{},
		},
		{
			query: gqlRemoteNetworks{
				Edges: []*gqlRemoteNetworkEdge{
					{
						Node: gqlRemoteNetwork{
							ID:   "network-id",
							Name: "network-name",
						},
					},
				},
			},
			expected: []*model.RemoteNetwork{
				{
					ID:   "network-id",
					Name: "network-name",
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestIsDefaultValue(t *testing.T) {
	cases := []struct {
		val      interface{}
		expected bool
	}{
		{
			val:      nil,
			expected: true,
		},
		{
			val:      "",
			expected: true,
		},
		{
			val:      "a",
			expected: false,
		},
		{
			val:      0,
			expected: true,
		},
		{
			val:      1,
			expected: false,
		},
		{
			val:      int32(0),
			expected: true,
		},
		{
			val:      int32(1),
			expected: false,
		},
		{
			val:      int64(0),
			expected: true,
		},
		{
			val:      int64(1),
			expected: false,
		},
		{
			val:      false,
			expected: true,
		},
		{
			val:      true,
			expected: false,
		},
		{
			val:      float64(0),
			expected: true,
		},
		{
			val:      float64(1),
			expected: false,
		},
		{
			val:      float32(0),
			expected: true,
		},
		{
			val:      float32(1),
			expected: false,
		},
		{
			val:      []interface{}{},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, isDefaultValue(c.val))
		})
	}
}

func TestGetDefaultGQLValue(t *testing.T) {
	var (
		defaultString *graphql.String
		defaultInt    *graphql.Int
		defaultBool   *graphql.Boolean
		defaultFloat  *graphql.Float
	)

	cases := []struct {
		val      interface{}
		expected interface{}
	}{
		{
			val:      nil,
			expected: nil,
		},
		{
			val:      "str",
			expected: defaultString,
		},
		{
			val:      true,
			expected: defaultBool,
		},
		{
			val:      1,
			expected: defaultInt,
		},
		{
			val:      int32(1),
			expected: defaultInt,
		},
		{
			val:      int64(1),
			expected: defaultInt,
		},
		{
			val:      float32(1.0),
			expected: defaultFloat,
		},
		{
			val:      1.0,
			expected: defaultFloat,
		},
		{
			val:      []interface{}{},
			expected: nil,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, getDefaultGQLValue(c.val))
		})
	}
}

func TestReadGroupQueryToModel(t *testing.T) {
	cases := []struct {
		query    readGroupQuery
		expected *model.Group
	}{
		{
			query:    readGroupQuery{},
			expected: nil,
		},
		{
			query: readGroupQuery{
				Group: &gqlGroup{
					IDName: IDName{
						ID:   "group-id",
						Name: "group-name",
					},
					IsActive: true,
					Type:     "new type",
				},
			},
			expected: &model.Group{
				ID:       "group-id",
				Name:     "group-name",
				IsActive: true,
				Type:     "new type",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestNewProtocolFromModel(t *testing.T) {
	cases := []struct {
		protocol *model.Protocol
		expected *ProtocolInput
	}{
		{},
		{
			protocol: &model.Protocol{
				Ports: []*model.PortRange{
					{Start: 80, End: 80},
				},
				Policy: "allow",
			},
			expected: &ProtocolInput{
				Ports: []*PortRangeInput{
					{Start: 80, End: 80},
				},
				Policy: "allow",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, newProtocol(c.protocol))
		})
	}
}

func TestProtocolToModel(t *testing.T) {
	cases := []struct {
		protocol *Protocol
		expected *model.Protocol
	}{
		{},
		{
			protocol: &Protocol{},
			expected: &model.Protocol{
				Ports: []*model.PortRange{},
			},
		},
		{
			protocol: &Protocol{
				Ports: []*PortRange{nil},
			},
			expected: &model.Protocol{
				Ports: []*model.PortRange{nil},
			},
		},
		{
			protocol: &Protocol{
				Ports: []*PortRange{
					{Start: 80, End: 80},
				},
				Policy: "allow",
			},
			expected: &model.Protocol{
				Ports: []*model.PortRange{
					{Start: 80, End: 80},
				},
				Policy: "allow",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, protocolToModel(c.protocol))
		})
	}
}

func TestReadServiceAccountQueryToModel(t *testing.T) {
	cases := []struct {
		query    readServiceAccountQuery
		expected *model.ServiceAccount
	}{
		{
			query:    readServiceAccountQuery{},
			expected: nil,
		},
		{
			query: readServiceAccountQuery{
				ServiceAccount: &gqlServiceAccount{
					IDName{
						ID:   "service-id",
						Name: "service-name",
					},
				},
			},
			expected: &model.ServiceAccount{
				ID:   "service-id",
				Name: "service-name",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestReadServiceKeyQueryToModel(t *testing.T) {
	_, parseErr := time.Parse(time.RFC3339, "hello")

	timeInHour := time.Now().Add(time.Hour).Format(time.RFC3339)

	cases := []struct {
		query       gqlServiceKey
		expected    *model.ServiceKey
		expectedErr error
	}{
		{
			query:    gqlServiceKey{},
			expected: &model.ServiceKey{},
		},
		{
			query: gqlServiceKey{
				ExpiresAt: "hello",
			},
			expectedErr: fmt.Errorf("failed to parse expiration time `hello`: %w", parseErr),
		},
		{
			query: gqlServiceKey{
				ExpiresAt: graphql.String(timeInHour),
			},
			expected: &model.ServiceKey{
				ExpirationTime: 1,
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			key, err := c.query.ToModel()
			assert.Equal(t, c.expected, key)
			assert.Equal(t, c.expectedErr, err)
		})
	}
}
