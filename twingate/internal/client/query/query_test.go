package query

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

var (
	boolTrue  = true
	boolFalse = false
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
				ID:                   "connector-id",
				Name:                 "connector-name",
				NetworkID:            "connector-network-id",
				StatusUpdatesEnabled: &boolFalse,
			},
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
					HasStatusNotificationsEnabled: true,
				},
			},
			expected: &model.Connector{
				ID:                   "connector-id",
				Name:                 "connector-name",
				NetworkID:            "connector-network-id",
				StatusUpdatesEnabled: &boolTrue,
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
					Protocols:                model.DefaultProtocols(),
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
					ID:                   "connector-id",
					Name:                 "connector-name",
					NetworkID:            "connector-network-id",
					StatusUpdatesEnabled: &boolFalse,
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
					ExpiresAt: today.Format(time.RFC3339),
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
						ExpiresAt: today.Format(time.RFC3339),
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
							ExpiresAt: today.Format(time.RFC3339),
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

func TestReadShallowServiceAccount(t *testing.T) {
	cases := []struct {
		query    ReadShallowServiceAccount
		expected *model.ServiceAccount
	}{
		{
			query:    ReadShallowServiceAccount{},
			expected: nil,
		},
		{
			query: ReadShallowServiceAccount{
				ServiceAccount: &gqlServiceAccount{
					IDName{
						ID:   "service-account-id",
						Name: "service-account-name",
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

func TestReadRemoteNetworkByIDToModel(t *testing.T) {
	cases := []struct {
		query    ReadRemoteNetworkByID
		expected *model.RemoteNetwork
	}{
		{
			query:    ReadRemoteNetworkByID{},
			expected: nil,
		},
		{
			query: ReadRemoteNetworkByID{
				RemoteNetwork: &gqlRemoteNetwork{
					IDName: IDName{
						ID:   "network-id",
						Name: "network-name",
					},
					Location: "AWS",
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
				Users:    []string{},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestReadGroupToModel(t *testing.T) {
	cases := []struct {
		query    ReadGroup
		expected *model.Group
	}{
		{
			query:    ReadGroup{},
			expected: nil,
		},
		{
			query: ReadGroup{
				Group: &gqlGroup{
					IDName: IDName{
						ID:   "group-id",
						Name: "group-name",
					},
					IsActive: true,
					Type:     "MANUAL",
				},
			},
			expected: &model.Group{
				ID:       "group-id",
				Name:     "group-name",
				IsActive: true,
				Type:     "MANUAL",
				Users:    []string{},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestReadSecurityPolicy(t *testing.T) {
	cases := []struct {
		query    ReadSecurityPolicy
		expected *model.SecurityPolicy
	}{
		{
			query:    ReadSecurityPolicy{},
			expected: nil,
		},
		{
			query: ReadSecurityPolicy{
				SecurityPolicy: &gqlSecurityPolicy{
					IDName{
						ID:   "policy-id",
						Name: "policy-name",
					},
				},
			},
			expected: &model.SecurityPolicy{
				ID:   "policy-id",
				Name: "policy-name",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestReadUserToModel(t *testing.T) {
	cases := []struct {
		query    ReadUser
		expected *model.User
	}{
		{
			query:    ReadUser{},
			expected: nil,
		},
		{
			query: ReadUser{
				User: &gqlUser{
					ID:        "user-id",
					FirstName: "First",
					LastName:  "Last",
					Email:     "email",
					Role:      "ADMIN",
				},
			},
			expected: &model.User{
				ID:        "user-id",
				FirstName: "First",
				LastName:  "Last",
				Email:     "email",
				Role:      "ADMIN",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestProtocolToModel(t *testing.T) {
	cases := []struct {
		protocol *Protocol
		expected *model.Protocol
	}{
		{
			protocol: nil,
			expected: model.DefaultProtocol(),
		},
		{
			protocol: &Protocol{
				Ports: []*PortRange{
					{Start: 80, End: 80},
				},
				Policy: "policy",
			},
			expected: &model.Protocol{
				Ports: []*model.PortRange{
					{Start: 80, End: 80},
				},
				Policy: "policy",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, protocolToModel(c.protocol))
		})
	}

}

func optionalBool(val bool) *bool {
	return &val
}

func TestBuildGroupsFilter(t *testing.T) {
	defaultActive := BooleanFilterOperatorInput{Eq: true}
	defaultType := GroupTypeFilterOperatorInput{
		In: []string{model.GroupTypeManual,
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
					Eq: optionalString("Group"),
				},
				Type:     defaultType,
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{Type: optionalString("MANUAL")},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []string{model.GroupTypeManual},
				},
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{Type: optionalString("SYSTEM")},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []string{model.GroupTypeSystem},
				},
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{Type: optionalString("SYNCED")},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []string{model.GroupTypeSynced},
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
					In: []string{model.GroupTypeSystem},
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
					In: []string{model.GroupTypeManual},
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
					In: []string{model.GroupTypeManual},
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

func TestPortsRangeToModel(t *testing.T) {
	cases := []struct {
		ports    []*PortRange
		expected []*model.PortRange
	}{
		{
			ports:    nil,
			expected: []*model.PortRange{},
		},
		{
			ports: []*PortRange{
				nil,
			},
			expected: []*model.PortRange{nil},
		},
		{
			ports: []*PortRange{
				{Start: 80, End: 90},
			},
			expected: []*model.PortRange{
				{Start: 80, End: 90},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, portsRangeToModel(c.ports))
		})
	}
}
