package query

import (
	"errors"
	"fmt"
	"testing"
	"time"

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

func TestReadServiceAccountKeyToModel(t *testing.T) {
	today := time.Now().Add(time.Hour)

	cases := []struct {
		query         ReadServiceAccountKey
		expected      *model.ServiceKey
		expectedError error
	}{
		{
			query:         ReadServiceAccountKey{},
			expected:      nil,
			expectedError: nil,
		},
		{
			query: ReadServiceAccountKey{
				ServiceAccountKey: &gqlServiceKey{
					ExpiresAt: "invalid date",
				},
			},
			expectedError: errors.New("failed to parse expiration time `invalid date`: parsing time \"invalid date\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid date\" as \"2006\""),
		},
		{
			query: ReadServiceAccountKey{
				ServiceAccountKey: &gqlServiceKey{
					IDName: IDName{
						ID:   "service-key-id",
						Name: "service key name",
					},
					ExpiresAt: graphql.String(today.Format(time.RFC3339)),
					Status:    "OK",
					ServiceAccount: gqlServiceAccount{
						IDName{
							ID: "service-account-id",
						},
					},
				},
			},
			expected: &model.ServiceKey{
				ID:             "service-key-id",
				Name:           "service key name",
				Service:        "service-account-id",
				ExpirationTime: 1,
				Status:         "OK",
			},
		},
		{
			query: ReadServiceAccountKey{
				ServiceAccountKey: &gqlServiceKey{
					IDName: IDName{
						ID:   "service-key-id",
						Name: "service key name",
					},
					ExpiresAt: "",
					Status:    "OK",
					ServiceAccount: gqlServiceAccount{
						IDName{
							ID: "service-account-id",
						},
					},
				},
			},
			expected: &model.ServiceKey{
				ID:             "service-key-id",
				Name:           "service key name",
				Service:        "service-account-id",
				ExpirationTime: 0,
				Status:         "OK",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			actual, err := c.query.ToModel()

			if c.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, c.expectedError.Error())
			}

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestUpdateServiceAccountKeyToModel(t *testing.T) {
	today := time.Now().Add(time.Hour)

	cases := []struct {
		query         UpdateServiceAccountKey
		expected      *model.ServiceKey
		expectedError error
	}{
		{
			query:         UpdateServiceAccountKey{},
			expected:      nil,
			expectedError: nil,
		},
		{
			query: UpdateServiceAccountKey{
				ServiceAccountKeyEntityResponse{
					Entity: &gqlServiceKey{
						ExpiresAt: "invalid date",
					},
				},
			},
			expectedError: errors.New("failed to parse expiration time `invalid date`: parsing time \"invalid date\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid date\" as \"2006\""),
		},
		{
			query: UpdateServiceAccountKey{
				ServiceAccountKeyEntityResponse{
					Entity: &gqlServiceKey{
						IDName: IDName{
							ID:   "service-key-id",
							Name: "service key name",
						},
						ExpiresAt: graphql.String(today.Format(time.RFC3339)),
						Status:    "OK",
						ServiceAccount: gqlServiceAccount{
							IDName{
								ID: "service-account-id",
							},
						},
					},
				},
			},
			expected: &model.ServiceKey{
				ID:             "service-key-id",
				Name:           "service key name",
				Service:        "service-account-id",
				ExpirationTime: 1,
				Status:         "OK",
			},
		},
		{
			query: UpdateServiceAccountKey{
				ServiceAccountKeyEntityResponse{
					Entity: &gqlServiceKey{
						IDName: IDName{
							ID:   "service-key-id",
							Name: "service key name",
						},
						ExpiresAt: "",
						Status:    "OK",
						ServiceAccount: gqlServiceAccount{
							IDName{
								ID: "service-account-id",
							},
						},
					},
				},
			},
			expected: &model.ServiceKey{
				ID:             "service-key-id",
				Name:           "service key name",
				Service:        "service-account-id",
				ExpirationTime: 0,
				Status:         "OK",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			actual, err := c.query.ToModel()

			if c.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, c.expectedError.Error())
			}

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestCreateServiceAccountKeyToModel(t *testing.T) {
	today := time.Now().Add(time.Hour)

	cases := []struct {
		query         CreateServiceAccountKey
		expected      *model.ServiceKey
		expectedError error
	}{
		{
			query:         CreateServiceAccountKey{},
			expected:      nil,
			expectedError: nil,
		},
		{
			query: CreateServiceAccountKey{
				ServiceAccountKeyEntityCreateResponse{
					ServiceAccountKeyEntityResponse: ServiceAccountKeyEntityResponse{
						Entity: &gqlServiceKey{
							ExpiresAt: "invalid date",
						},
					},
				},
			},
			expectedError: errors.New("failed to parse expiration time `invalid date`: parsing time \"invalid date\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid date\" as \"2006\""),
		},
		{
			query: CreateServiceAccountKey{
				ServiceAccountKeyEntityCreateResponse{
					ServiceAccountKeyEntityResponse: ServiceAccountKeyEntityResponse{
						Entity: &gqlServiceKey{
							IDName: IDName{
								ID:   "service-key-id",
								Name: "service key name",
							},
							ExpiresAt: graphql.String(today.Format(time.RFC3339)),
							Status:    "OK",
							ServiceAccount: gqlServiceAccount{
								IDName{
									ID: "service-account-id",
								},
							},
						},
					},
					Token: "token",
				},
			},
			expected: &model.ServiceKey{
				ID:             "service-key-id",
				Name:           "service key name",
				Service:        "service-account-id",
				ExpirationTime: 1,
				Status:         "OK",
				Token:          "token",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			actual, err := c.query.ToModel()

			if c.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, c.expectedError.Error())
			}

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestCreateServiceAccountToModel(t *testing.T) {
	cases := []struct {
		query    CreateServiceAccount
		expected *model.ServiceAccount
	}{
		{
			query:    CreateServiceAccount{},
			expected: nil,
		},
		{
			query: CreateServiceAccount{
				ServiceAccountEntityResponse{
					Entity: &gqlServiceAccount{
						IDName{
							ID:   "service-account-id",
							Name: "service-account-name",
						},
					},
				},
			},
			expected: &model.ServiceAccount{
				ID:   "service-account-id",
				Name: "service-account-name",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestUpdateServiceAccountToModel(t *testing.T) {
	cases := []struct {
		query    UpdateServiceAccount
		expected *model.ServiceAccount
	}{
		{
			query:    UpdateServiceAccount{},
			expected: nil,
		},
		{
			query: UpdateServiceAccount{
				ServiceAccountEntityResponse{
					Entity: &gqlServiceAccount{
						IDName{
							ID:   "service-account-id",
							Name: "service-account-name",
						},
					},
				},
			},
			expected: &model.ServiceAccount{
				ID:   "service-account-id",
				Name: "service-account-name",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestUpdateRemoteNetworkToModel(t *testing.T) {
	cases := []struct {
		query    UpdateRemoteNetwork
		expected *model.RemoteNetwork
	}{
		{
			query:    UpdateRemoteNetwork{},
			expected: nil,
		},
		{
			query: UpdateRemoteNetwork{
				RemoteNetworkEntityResponse{
					Entity: &gqlRemoteNetwork{
						IDName: IDName{
							ID:   "network-id",
							Name: "network-name",
						},
						Location: "AWS",
					},
				},
			},
			expected: &model.RemoteNetwork{
				ID:       "network-id",
				Name:     "network-name",
				Location: "AWS",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestCreateRemoteNetworkToModel(t *testing.T) {
	cases := []struct {
		query    CreateRemoteNetwork
		expected *model.RemoteNetwork
	}{
		{
			query:    CreateRemoteNetwork{},
			expected: nil,
		},
		{
			query: CreateRemoteNetwork{
				RemoteNetworkEntityResponse{
					Entity: &gqlRemoteNetwork{
						IDName: IDName{
							ID:   "network-id",
							Name: "network-name",
						},
						Location: "AWS",
					},
				},
			},
			expected: &model.RemoteNetwork{
				ID:       "network-id",
				Name:     "network-name",
				Location: "AWS",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestCreateGroupToModel(t *testing.T) {
	cases := []struct {
		query    CreateGroup
		expected *model.Group
	}{
		{
			query:    CreateGroup{},
			expected: nil,
		},
		{
			query: CreateGroup{
				GroupEntityResponse{
					Entity: &gqlGroup{
						IDName: IDName{
							ID:   "group-id",
							Name: "group-name",
						},
						IsActive: true,
						Type:     "MANUAL",
					},
				},
			},
			expected: &model.Group{
				ID:       "group-id",
				Name:     "group-name",
				IsActive: true,
				Type:     "MANUAL",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}
