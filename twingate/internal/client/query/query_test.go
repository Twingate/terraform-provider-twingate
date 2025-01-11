package query

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

var (
	boolTrue  = true
	boolFalse = false
)

func TestOkError(t *testing.T) {
	cases := []struct {
		query         OkError
		expectedOk    bool
		expectedError string
	}{
		{
			query: OkError{},
		},
		{
			query: OkError{
				Ok: true,
			},
			expectedOk: true,
		},
		{
			query: OkError{
				Ok:    false,
				Error: "some error",
			},
			expectedOk:    false,
			expectedError: "some error",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expectedOk, c.query.OK())
			assert.Equal(t, c.expectedError, c.query.ErrorStr())
		})
	}
}

func TestDeleteConnectorQuery(t *testing.T) {
	cases := []struct {
		query    DeleteConnector
		expected bool
	}{
		{
			query:    DeleteConnector{},
			expected: false,
		},
		{
			query: DeleteConnector{
				OkError{
					Ok: true,
				},
			},
			expected: false,
		},
		{
			query: DeleteConnector{
				OkError{
					Ok: false,
				},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.IsEmpty())
		})
	}
}

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
			assert.Equal(t, c.expected == nil, c.query.IsEmpty())
		})
	}
}

func TestCreateConnectorQueryResponse(t *testing.T) {
	cases := []struct {
		query    CreateConnector
		expected bool
	}{
		{
			query:    CreateConnector{},
			expected: true,
		},
		{
			query: CreateConnector{
				ConnectorEntityResponse{
					Entity: &gqlConnector{},
				},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.IsEmpty())
		})
	}
}

func TestReadResourcesByNameQueryToModel(t *testing.T) {
	var (
		boolTrue  = true
		boolFalse = false
	)

	cases := []struct {
		query         ReadResourcesByName
		expected      []*model.Resource
		expectedEmpty bool
	}{
		{
			query:         ReadResourcesByName{},
			expected:      []*model.Resource{},
			expectedEmpty: true,
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
			expectedEmpty: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, c.query.ToModel())
			assert.Equal(t, c.expectedEmpty, c.query.IsEmpty())
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

			if c.expected == nil {
				assert.True(t, c.query.IsEmpty())
			} else {
				assert.False(t, c.query.IsEmpty())
			}
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

			if c.expected == nil && c.expectedError == nil {
				assert.True(t, c.query.IsEmpty())
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

			if c.expected == nil && c.expectedError == nil {
				assert.True(t, c.query.IsEmpty())
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

			if c.expected == nil {
				assert.True(t, c.query.IsEmpty())
			}
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

			if c.expected == nil {
				assert.True(t, c.query.IsEmpty())
			}
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

			if c.expected == nil {
				assert.True(t, c.query.IsEmpty())
			}
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

			if c.expected == nil {
				assert.True(t, c.query.IsEmpty())
			}
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
		{
			query: ReadGroup{
				Group: &gqlGroup{
					IDName: IDName{
						ID:   "group-1",
						Name: "group-name",
					},
					IsActive: true,
					Type:     "MANUAL",
					Users: Users{
						PaginatedResource[*UserEdge]{
							Edges: []*UserEdge{
								{
									Node: &gqlUser{
										ID:        "user-1",
										FirstName: "First",
										LastName:  "Last",
										Email:     "email",
										Role:      "ADMIN",
									},
								},
								{
									Node: &gqlUser{
										ID:        "user-2",
										FirstName: "Second",
										LastName:  "Last",
										Email:     "email",
										Role:      "ADMIN",
									},
								},
							},
						},
					},
				},
			},
			expected: &model.Group{
				ID:       "group-1",
				Name:     "group-name",
				IsActive: true,
				Type:     "MANUAL",
				Users:    []string{"user-1", "user-2"},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())

			assert.Equal(t, c.expected == nil, c.query.IsEmpty())
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

			if c.expected == nil {
				assert.True(t, c.query.IsEmpty())
			}
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

			if c.expected == nil {
				assert.True(t, c.query.IsEmpty())
			}
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
			filter: &model.GroupsFilter{Types: []string{"MANUAL"}},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []string{model.GroupTypeManual},
				},
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{Types: []string{"SYSTEM"}},
			expected: &GroupFilterInput{
				Type: GroupTypeFilterOperatorInput{
					In: []string{model.GroupTypeSystem},
				},
				IsActive: defaultActive,
			},
		},
		{
			filter: &model.GroupsFilter{Types: []string{"SYNCED"}},
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
				Types:    []string{"SYSTEM"},
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
				Types:    []string{"MANUAL"},
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
				Types:    []string{"MANUAL"},
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

func TestGenerateConnectorTokensToModel(t *testing.T) {
	cases := []struct {
		query    GenerateConnectorTokens
		expected *model.ConnectorTokens
	}{
		{
			query: GenerateConnectorTokens{},
			expected: &model.ConnectorTokens{
				AccessToken:  "",
				RefreshToken: "",
			},
		},
		{
			query: GenerateConnectorTokens{
				ConnectorTokensResponse: ConnectorTokensResponse{
					ConnectorTokens: gqlConnectorTokens{
						AccessToken:  "test-access-token",
						RefreshToken: "test-refresh-token",
					},
				},
			},
			expected: &model.ConnectorTokens{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())

			if c.expected.AccessToken == "" && c.expected.RefreshToken == "" {
				assert.True(t, c.query.IsEmpty())
			} else {
				assert.False(t, c.query.IsEmpty())
			}
		})
	}
}

func TestNewConnectorFilterInput(t *testing.T) {
	cases := []struct {
		name     string
		filter   string
		expected *ConnectorFilterInput
	}{
		{
			name:     "",
			filter:   "",
			expected: &ConnectorFilterInput{},
		},
		{
			name:   "Empty filter",
			filter: "",
			expected: &ConnectorFilterInput{
				Name: &StringFilterOperationInput{
					Eq: optionalString("Empty filter"),
				},
			},
		},
		{
			name:   "Valid filter",
			filter: "_regexp",
			expected: &ConnectorFilterInput{
				Name: &StringFilterOperationInput{
					Regexp: optionalString("Valid filter"),
				},
			},
		},
		{
			name:   "Prefix filter",
			filter: "_prefix",
			expected: &ConnectorFilterInput{
				Name: &StringFilterOperationInput{
					StartsWith: optionalString("Prefix filter"),
				},
			},
		},
		{
			name:   "Suffix filter",
			filter: "_suffix",
			expected: &ConnectorFilterInput{
				Name: &StringFilterOperationInput{
					EndsWith: optionalString("Suffix filter"),
				},
			},
		},
		{
			name:   "Contains filter",
			filter: "_contains",
			expected: &ConnectorFilterInput{
				Name: &StringFilterOperationInput{
					Contains: optionalString("Contains filter"),
				},
			},
		},
		{
			name:   "Exclude filter",
			filter: "_exclude",
			expected: &ConnectorFilterInput{
				Name: &StringFilterOperationInput{
					Ne: optionalString("Exclude filter"),
				},
			},
		},
		{
			name:   "Unknown filter type",
			filter: "_unknown",
			expected: &ConnectorFilterInput{
				Name: &StringFilterOperationInput{
					Eq: optionalString("Unknown filter type"),
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := NewConnectorFilterInput(c.name, c.filter)

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestDNSFilteringProfileEntityResponse_IsEmpty(t *testing.T) {
	cases := []struct {
		query    *DNSFilteringProfileEntityResponse
		expected bool
	}{
		{
			query:    nil,
			expected: true,
		},
		{
			query: &DNSFilteringProfileEntityResponse{
				Entity: nil,
			},
			expected: true,
		},
		{
			query: &DNSFilteringProfileEntityResponse{
				Entity: &gqlDNSFilteringProfile{},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			actual := c.query.IsEmpty()
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestDeleteDNSFilteringProfile(t *testing.T) {
	cases := []struct {
		query    *DeleteDNSFilteringProfile
		expected bool
	}{
		{
			query:    &DeleteDNSFilteringProfile{},
			expected: false,
		},
		{
			query: &DeleteDNSFilteringProfile{
				OkError{
					Ok: true,
				},
			},
			expected: false,
		},
		{
			query: &DeleteDNSFilteringProfile{
				OkError{
					Ok: false,
				},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			actual := c.query.IsEmpty()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadDNSFilteringProfile_IsEmpty(t *testing.T) {
	cases := []struct {
		query    ReadDNSFilteringProfile
		expected bool
	}{
		{
			query: ReadDNSFilteringProfile{
				DNSFilteringProfile: nil,
			},
			expected: true,
		},
		{
			query: ReadDNSFilteringProfile{
				DNSFilteringProfile: &gqlDNSFilteringProfile{},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			actual := c.query.IsEmpty()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadDNSFilteringProfile_ToModel(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadDNSFilteringProfile
		expected *model.DNSFilteringProfile
	}{
		{
			name: "Nil DNSFilteringProfile",
			query: ReadDNSFilteringProfile{
				DNSFilteringProfile: nil,
			},
			expected: nil,
		},
		{
			name: "Valid DNSFilteringProfile with PrivacyCategoryConfig",
			query: ReadDNSFilteringProfile{
				DNSFilteringProfile: &gqlDNSFilteringProfile{
					IDName:         IDName{ID: "123", Name: "Test Profile"},
					Priority:       1.0,
					FallbackMethod: "block",
					AllowedDomains: []string{"example.com", "example.org"},
					DeniedDomains:  []string{"malicious.com"},
					PrivacyCategoryConfig: &PrivacyCategoryConfig{
						BlockAffiliate:         true,
						BlockDisguisedTrackers: false,
						BlockAdsAndTrackers:    true,
					},
					Groups: gqlGroupIDs{
						PaginatedResource[*GroupIDEdge]{
							Edges: []*GroupIDEdge{
								{Node: &gqlGroupID{IDName: IDName{ID: "group1"}}},
								{Node: &gqlGroupID{IDName: IDName{ID: "group2"}}},
							},
						},
					},
				},
			},
			expected: &model.DNSFilteringProfile{
				ID:             "123",
				Name:           "Test Profile",
				Priority:       1.0,
				FallbackMethod: "block",
				AllowedDomains: []string{"example.com", "example.org"},
				DeniedDomains:  []string{"malicious.com"},
				PrivacyCategories: &model.PrivacyCategories{
					BlockAffiliate:         true,
					BlockDisguisedTrackers: false,
					BlockAdsAndTrackers:    true,
				},
				Groups: []string{"group1", "group2"},
			},
		},
		{
			name: "DNSFilteringProfile with no Optional Configs",
			query: ReadDNSFilteringProfile{
				DNSFilteringProfile: &gqlDNSFilteringProfile{
					IDName:         IDName{ID: "456", Name: "Another Profile"},
					Priority:       2.0,
					FallbackMethod: "monitor",
					AllowedDomains: []string{"test.com"},
					DeniedDomains:  nil,
					Groups: gqlGroupIDs{
						PaginatedResource[*GroupIDEdge]{
							Edges: []*GroupIDEdge{
								{Node: &gqlGroupID{IDName: IDName{ID: "group3"}}},
							},
						},
					},
					PrivacyCategoryConfig:  nil,
					SecurityCategoryConfig: nil,
					ContentCategoryConfig:  nil,
				},
			},
			expected: &model.DNSFilteringProfile{
				ID:                 "456",
				Name:               "Another Profile",
				Priority:           2.0,
				FallbackMethod:     "monitor",
				AllowedDomains:     []string{"test.com"},
				DeniedDomains:      nil,
				Groups:             []string{"group3"},
				PrivacyCategories:  nil,
				SecurityCategories: nil,
				ContentCategories:  nil,
			},
		},
		{
			name: "Valid DNSFilteringProfile with Full Configs",
			query: ReadDNSFilteringProfile{
				DNSFilteringProfile: &gqlDNSFilteringProfile{
					IDName:         IDName{ID: "123", Name: "Test Profile"},
					Priority:       1.0,
					FallbackMethod: "block",
					AllowedDomains: []string{"example.com", "example.org"},
					DeniedDomains:  []string{"malicious.com"},
					PrivacyCategoryConfig: &PrivacyCategoryConfig{
						BlockAffiliate:         true,
						BlockDisguisedTrackers: false,
						BlockAdsAndTrackers:    true,
					},
					SecurityCategoryConfig: &SecurityCategoryConfig{
						EnableThreatIntelligenceFeeds:   true,
						EnableGoogleSafeBrowsing:        true,
						BlockCryptojacking:              true,
						BlockIdnHomographs:              false,
						BlockTyposquatting:              false,
						BlockDnsRebinding:               true,
						BlockNewlyRegisteredDomains:     true,
						BlockDomainGenerationAlgorithms: false,
						BlockParkedDomains:              false,
					},
					ContentCategoryConfig: &ContentCategoryConfig{
						BlockGambling:               true,
						BlockDating:                 false,
						BlockAdultContent:           true,
						BlockSocialMedia:            false,
						BlockGames:                  false,
						BlockStreaming:              true,
						BlockPiracy:                 true,
						EnableYoutubeRestrictedMode: false,
						EnableSafeSearch:            true,
					},
					Groups: gqlGroupIDs{
						PaginatedResource[*GroupIDEdge]{
							Edges: []*GroupIDEdge{
								{Node: &gqlGroupID{IDName: IDName{ID: "group1"}}},
							},
						},
					},
				},
			},
			expected: &model.DNSFilteringProfile{
				ID:             "123",
				Name:           "Test Profile",
				Priority:       1.0,
				FallbackMethod: "block",
				AllowedDomains: []string{"example.com", "example.org"},
				DeniedDomains:  []string{"malicious.com"},
				PrivacyCategories: &model.PrivacyCategories{
					BlockAffiliate:         true,
					BlockDisguisedTrackers: false,
					BlockAdsAndTrackers:    true,
				},
				SecurityCategories: &model.SecurityCategory{
					EnableThreatIntelligenceFeeds:   true,
					EnableGoogleSafeBrowsing:        true,
					BlockCryptojacking:              true,
					BlockIdnHomographs:              false,
					BlockTyposquatting:              false,
					BlockDNSRebinding:               true,
					BlockNewlyRegisteredDomains:     true,
					BlockDomainGenerationAlgorithms: false,
					BlockParkedDomains:              false,
				},
				ContentCategories: &model.ContentCategory{
					BlockGambling:               true,
					BlockDating:                 false,
					BlockAdultContent:           true,
					BlockSocialMedia:            false,
					BlockGames:                  false,
					BlockStreaming:              true,
					BlockPiracy:                 true,
					EnableYoutubeRestrictedMode: false,
					EnableSafeSearch:            true,
				},
				Groups: []string{"group1"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.query.ToModel()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadDNSFilteringProfileGroups_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadDNSFilteringProfileGroups
		expected bool
	}{
		{
			name: "Nil DNSFilteringProfile",
			query: ReadDNSFilteringProfileGroups{
				DNSFilteringProfile: nil,
			},
			expected: true,
		},
		{
			name: "Non-nil DNSFilteringProfile",
			query: ReadDNSFilteringProfileGroups{
				DNSFilteringProfile: &gqlDNSFilteringProfileGroups{},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.query.IsEmpty()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadDNSFilteringProfiles_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadDNSFilteringProfiles
		expected bool
	}{
		{
			name: "Empty DNS Filtering Profiles",
			query: ReadDNSFilteringProfiles{
				DNSFilteringProfiles: nil,
			},
			expected: true,
		},
		{
			name: "Non-empty DNS Filtering Profiles",
			query: ReadDNSFilteringProfiles{
				DNSFilteringProfiles: []*gqlShallowDNSFilteringProfile{
					{IDName: IDName{ID: "123", Name: "Profile 1"}, Priority: 1.0},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.query.IsEmpty()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadDNSFilteringProfiles_ToModel(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadDNSFilteringProfiles
		expected []*model.DNSFilteringProfile
	}{
		{
			name: "Empty DNS Filtering Profiles",
			query: ReadDNSFilteringProfiles{
				DNSFilteringProfiles: nil,
			},
			expected: []*model.DNSFilteringProfile{},
		},
		{
			name: "Single Profile Conversion",
			query: ReadDNSFilteringProfiles{
				DNSFilteringProfiles: []*gqlShallowDNSFilteringProfile{
					{IDName: IDName{ID: "123", Name: "Profile 1"}, Priority: 1.0},
				},
			},
			expected: []*model.DNSFilteringProfile{
				{ID: "123", Name: "Profile 1", Priority: 1.0},
			},
		},
		{
			name: "Multiple Profiles Conversion",
			query: ReadDNSFilteringProfiles{
				DNSFilteringProfiles: []*gqlShallowDNSFilteringProfile{
					{IDName: IDName{ID: "123", Name: "Profile 1"}, Priority: 1.0},
					{IDName: IDName{ID: "456", Name: "Profile 2"}, Priority: 2.0},
				},
			},
			expected: []*model.DNSFilteringProfile{
				{ID: "123", Name: "Profile 1", Priority: 1.0},
				{ID: "456", Name: "Profile 2", Priority: 2.0},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.query.ToModel()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestDeleteGroup_IsEmpty(t *testing.T) {
	cases := []struct {
		query    DeleteGroup
		expected bool
	}{
		{
			query:    DeleteGroup{},
			expected: false,
		},
		{
			query: DeleteGroup{
				OkError: OkError{
					Ok: true,
				},
			},
			expected: false,
		},
		{
			query: DeleteGroup{
				OkError: OkError{
					Ok: false,
				},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := c.query.IsEmpty()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestUpdateGroup_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    UpdateGroup
		expected bool
	}{
		{
			name: "UpdateGroup with nil Entity",
			query: UpdateGroup{
				GroupEntityResponse: GroupEntityResponse{
					Entity: nil,
				},
			},
			expected: true,
		},
		{
			name: "UpdateGroup with non-nil Entity",
			query: UpdateGroup{
				GroupEntityResponse: GroupEntityResponse{
					Entity: &gqlGroup{},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.query.IsEmpty()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestUpdateGroupRemoveUsers_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    UpdateGroupRemoveUsers
		expected bool
	}{
		{
			name: "UpdateGroupRemoveUsers with nil Entity",
			query: UpdateGroupRemoveUsers{
				GroupEntityResponse: GroupEntityResponse{
					Entity: nil,
				},
			},
			expected: true,
		},
		{
			name: "UpdateGroupRemoveUsers with non-nil Entity",
			query: UpdateGroupRemoveUsers{
				GroupEntityResponse: GroupEntityResponse{
					Entity: &gqlGroup{},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.query.IsEmpty()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadGroups_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadGroups
		expected bool
	}{
		{
			name: "No edges in groups (empty)",
			query: ReadGroups{
				Groups: Groups{
					PaginatedResource: PaginatedResource[*GroupEdge]{
						Edges: nil,
					},
				},
			},
			expected: true,
		},
		{
			name: "Edges present in groups (non-empty)",
			query: ReadGroups{
				Groups: Groups{
					PaginatedResource: PaginatedResource[*GroupEdge]{
						Edges: []*GroupEdge{
							{Node: &gqlGroup{}},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.query.IsEmpty()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadGroups_ToModel(t *testing.T) {
	cases := []struct {
		name     string
		groups   ReadGroups
		expected []*model.Group
	}{
		{
			name: "No groups",
			groups: ReadGroups{
				Groups: Groups{
					PaginatedResource: PaginatedResource[*GroupEdge]{
						Edges: nil,
					},
				},
			},
			expected: []*model.Group{},
		},
		{
			name: "One group",
			groups: ReadGroups{
				Groups: Groups{
					PaginatedResource: PaginatedResource[*GroupEdge]{
						Edges: []*GroupEdge{
							{
								Node: &gqlGroup{
									IDName: IDName{
										ID:   "group1",
										Name: "Group 1",
									},
								},
							},
						},
					},
				},
			},
			expected: []*model.Group{
				{
					ID:    "group1",
					Name:  "Group 1",
					Users: []string{},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.groups.ToModel()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestReadRemoteNetworkByName_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadRemoteNetworkByName
		expected bool
	}{
		{
			name: "No edges in RemoteNetworks",
			query: ReadRemoteNetworkByName{
				RemoteNetworks: gqlRemoteNetworks{
					Edges: nil, // No edges
				},
			},
			expected: true,
		},
		{
			name: "Edges slice is empty",
			query: ReadRemoteNetworkByName{
				RemoteNetworks: gqlRemoteNetworks{
					Edges: []*RemoteNetworkEdge{}, // Empty edges slice
				},
			},
			expected: true,
		},
		{
			name: "First edge is nil",
			query: ReadRemoteNetworkByName{
				RemoteNetworks: gqlRemoteNetworks{
					Edges: []*RemoteNetworkEdge{nil}, // First edge is nil
				},
			},
			expected: true,
		},
		{
			name: "Edges contain valid data",
			query: ReadRemoteNetworkByName{
				RemoteNetworks: gqlRemoteNetworks{
					Edges: []*RemoteNetworkEdge{
						{
							Node: gqlRemoteNetwork{
								IDName: IDName{
									ID:   "network1",
									Name: "Network 1",
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestDeleteRemoteNetwork_IsEmpty(t *testing.T) {
	cases := []struct {
		query    DeleteRemoteNetwork
		expected bool
	}{
		{
			query:    DeleteRemoteNetwork{},
			expected: false,
		},
		{
			query: DeleteRemoteNetwork{
				OkError: OkError{
					Ok: true,
				},
			},
			expected: false,
		},
		{
			query: DeleteRemoteNetwork{
				OkError: OkError{
					Ok: false,
				},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestReadRemoteNetworks(t *testing.T) {
	cases := []struct {
		name          string
		query         ReadRemoteNetworks
		expectedEmpty bool
		expected      []*model.RemoteNetwork
	}{
		{
			name: "No edges in RemoteNetworks",
			query: ReadRemoteNetworks{
				RemoteNetworks: RemoteNetworks{
					PaginatedResource: PaginatedResource[*RemoteNetworkEdge]{
						Edges: nil, // No edges present
					},
				},
			},
			expectedEmpty: true,
			expected:      []*model.RemoteNetwork{},
		},
		{
			name: "Edges slice is empty",
			query: ReadRemoteNetworks{
				RemoteNetworks: RemoteNetworks{
					PaginatedResource: PaginatedResource[*RemoteNetworkEdge]{
						Edges: []*RemoteNetworkEdge{}, // Empty edges
					},
				},
			},
			expectedEmpty: true,
			expected:      []*model.RemoteNetwork{},
		},
		{
			name: "Edges contain data",
			query: ReadRemoteNetworks{
				RemoteNetworks: RemoteNetworks{
					PaginatedResource: PaginatedResource[*RemoteNetworkEdge]{
						Edges: []*RemoteNetworkEdge{
							{
								Node: gqlRemoteNetwork{
									IDName: IDName{
										ID:   "network-id",
										Name: "network-name",
									},
								},
							},
						},
					},
				},
			},
			expectedEmpty: false,
			expected: []*model.RemoteNetwork{
				{
					ID:   "network-id",
					Name: "network-name",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expectedEmpty, c.query.IsEmpty())
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}

func TestRemoteNetworkFilter(t *testing.T) {
	cases := []struct {
		name           string
		inputName      string
		inputFilter    string
		expectedFilter *StringFilterOperationInput
	}{
		{
			name:        "Basic name and filter",
			inputName:   "network1",
			inputFilter: "",
			expectedFilter: &StringFilterOperationInput{
				Eq: optionalString("network1"),
			},
		},
		{
			name:        "Prefix name",
			inputName:   "name",
			inputFilter: "_prefix",
			expectedFilter: &StringFilterOperationInput{
				StartsWith: optionalString("name"),
			},
		},
		{
			name:           "Both name and filter empty",
			inputName:      "",
			inputFilter:    "",
			expectedFilter: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := NewRemoteNetworkFilterInput(c.inputName, c.inputFilter)

			assert.Equal(t, c.expectedFilter, result.Name)
		})
	}
}

func TestAddResourceAccess_IsEmpty(t *testing.T) {
	cases := []struct {
		query    AddResourceAccess
		expected bool
	}{
		{
			query:    AddResourceAccess{},
			expected: false,
		},
		{
			query: AddResourceAccess{
				OkError{Ok: true},
			},
			expected: false,
		},
		{
			query: AddResourceAccess{
				OkError{Ok: false},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestRemoveResourceAccess_IsEmpty(t *testing.T) {
	cases := []struct {
		query    RemoveResourceAccess
		expected bool
	}{
		{
			query:    RemoveResourceAccess{},
			expected: false,
		},
		{
			query: RemoveResourceAccess{
				OkError{Ok: true},
			},
			expected: false,
		},
		{
			query: RemoveResourceAccess{
				OkError{Ok: false},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestDeleteResource_IsEmpty(t *testing.T) {
	cases := []struct {
		query    DeleteResource
		expected bool
	}{
		{
			query:    DeleteResource{},
			expected: false,
		},
		{
			query: DeleteResource{
				OkError{Ok: true},
			},
			expected: false,
		},
		{
			query: DeleteResource{
				OkError{Ok: false},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestUpdateResourceActiveState_IsEmpty(t *testing.T) {
	cases := []struct {
		query    UpdateResourceActiveState
		expected bool
	}{
		{
			query:    UpdateResourceActiveState{},
			expected: false,
		},
		{
			query: UpdateResourceActiveState{
				OkError{Ok: true},
			},
			expected: false,
		},
		{
			query: UpdateResourceActiveState{
				OkError{Ok: false},
			},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestReadResourceAccess_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadResourceAccess
		expected bool
	}{
		{
			name: "Resource is nil",
			query: ReadResourceAccess{
				Resource: nil,
			},
			expected: true,
		},
		{
			name: "Resource is not nil",
			query: ReadResourceAccess{
				Resource: &gqlResourceAccess{
					ID:     "123",
					Access: Access{},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestCreateResource_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    CreateResource
		expected bool
	}{
		{
			name: "Resource is nil",
			query: CreateResource{
				ResourceEntityResponse: ResourceEntityResponse{
					Entity: nil,
					OkError: OkError{
						Ok: true,
					},
				},
			},
			expected: true,
		},
		{
			name: "Resource is not nil",
			query: CreateResource{
				ResourceEntityResponse{
					Entity: &gqlResource{
						ResourceNode: ResourceNode{
							IDName: IDName{
								ID:   "123",
								Name: "Resource 1",
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestReadResource_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadResource
		expected bool
	}{
		{
			name: "Resource is nil",
			query: ReadResource{
				Resource: nil,
			},
			expected: true,
		},
		{
			name: "Resource is not nil",
			query: ReadResource{
				Resource: &gqlResource{
					ResourceNode: ResourceNode{
						IDName: IDName{
							ID:   "123",
							Name: "Resource 1",
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestUpdateResource_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    UpdateResource
		expected bool
	}{
		{
			name: "Resource is nil",
			query: UpdateResource{
				ResourceEntityResponse: ResourceEntityResponse{
					Entity: nil,
				},
			},
			expected: true,
		},
		{
			name: "Resource is not nil",
			query: UpdateResource{
				ResourceEntityResponse: ResourceEntityResponse{
					Entity: &gqlResource{
						ResourceNode: ResourceNode{
							IDName: IDName{
								ID:   "123",
								Name: "Resource 1",
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}
func TestUpdateResourceRemoveGroups_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    UpdateResourceRemoveGroups
		expected bool
	}{
		{
			name: "Resource is nil",
			query: UpdateResourceRemoveGroups{
				ResourceEntityResponse: ResourceEntityResponse{
					Entity: nil,
				},
			},
			expected: true,
		},
		{
			name: "Resource is not nil",
			query: UpdateResourceRemoveGroups{
				ResourceEntityResponse: ResourceEntityResponse{
					Entity: &gqlResource{
						ResourceNode: ResourceNode{
							IDName: IDName{
								ID:   "123",
								Name: "Resource 1",
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			isEmpty := c.query.IsEmpty()

			assert.Equal(t, c.expected, isEmpty)
		})
	}
}

func TestReadResource_ToModel(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadResource
		expected *model.Resource
	}{
		{
			name: "Resource is nil",
			query: ReadResource{
				Resource: nil,
			},
			expected: nil,
		},
		{
			name: "Resource with no access edges",
			query: ReadResource{
				Resource: &gqlResource{
					ResourceNode: ResourceNode{
						IDName: IDName{
							ID:   "resource123",
							Name: "Resource Name",
						},
					},
					Access: Access{
						PaginatedResource: PaginatedResource[*AccessEdge]{
							Edges: nil,
						},
					},
				},
			},
			expected: &model.Resource{
				ID:   "resource123",
				Name: "Resource Name",
				Protocols: &model.Protocols{
					TCP: &model.Protocol{
						Policy: model.PolicyAllowAll,
					},
					UDP: &model.Protocol{
						Policy: model.PolicyAllowAll,
					},
					AllowIcmp: true,
				},
				IsVisible:                optionalBool(false),
				IsBrowserShortcutEnabled: optionalBool(false),
			},
		},
		{
			name: "Resource with multiple access edges",
			query: ReadResource{
				Resource: &gqlResource{
					ResourceNode: ResourceNode{
						IDName: IDName{
							ID:   "resource456",
							Name: "Another Resource",
						},
						SecurityPolicy: &gqlSecurityPolicy{
							IDName{ID: "policy123", Name: "Policy 1"},
						},
						Protocols: &Protocols{
							TCP: &Protocol{
								Ports: []*PortRange{
									{Start: 100, End: 200},
								},
								Policy: model.PolicyRestricted,
							},
							UDP: &Protocol{
								Policy: model.PolicyDenyAll,
							},
							AllowIcmp: false,
						},
					},
					Access: Access{
						PaginatedResource: PaginatedResource[*AccessEdge]{
							Edges: []*AccessEdge{
								{
									Node: Principal{
										Type: "Group",
										Node: Node{ID: "group123"},
									},
									SecurityPolicy: &gqlSecurityPolicy{
										IDName{
											ID: "policy789",
										},
									},
									UsageBasedAutolockDurationDays: optionalInt64(30),
								},
								{
									Node: Principal{
										Type: "ServiceAccount",
										Node: Node{ID: "serviceAccount456"},
									},
								},
							},
						},
					},
				},
			},
			expected: &model.Resource{
				ID:               "resource456",
				Name:             "Another Resource",
				SecurityPolicyID: optionalString("policy123"),
				GroupsAccess: []model.AccessGroup{
					{
						GroupID:            "group123",
						SecurityPolicyID:   optionalString("policy789"),
						UsageBasedDuration: optionalInt64(30),
					},
				},
				ServiceAccounts: []string{"serviceAccount456"},
				Protocols: &model.Protocols{
					TCP: &model.Protocol{
						Ports: []*model.PortRange{
							{Start: 100, End: 200},
						},
						Policy: model.PolicyRestricted,
					},
					UDP: &model.Protocol{
						Ports:  []*model.PortRange{},
						Policy: model.PolicyDenyAll,
					},
					AllowIcmp: false,
				},
				IsVisible:                optionalBool(false),
				IsBrowserShortcutEnabled: optionalBool(false),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !c.query.IsEmpty() {
				assert.Equal(t, c.expected, c.query.Resource.ToModel())
			}
		})
	}
}

func TestReadResourcesByName_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadResourcesByName
		expected bool
	}{
		{
			name: "No edges - should be empty",
			query: ReadResourcesByName{
				Resources: Resources{
					PaginatedResource: PaginatedResource[*ResourceEdge]{},
				},
			},
			expected: true,
		},
		{
			name: "Edges present - should not be empty",
			query: ReadResourcesByName{
				Resources: Resources{
					PaginatedResource: PaginatedResource[*ResourceEdge]{
						Edges: []*ResourceEdge{
							{
								Node: &ResourceNode{
									IDName: IDName{
										ID:   "123",
										Name: "TestResource",
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.IsEmpty())
		})
	}
}

func TestResourceFilter(t *testing.T) {
	cases := []struct {
		name           string
		inputName      string
		inputFilter    string
		expectedFilter *StringFilterOperationInput
	}{
		{
			name:        "Basic name and filter",
			inputName:   "network1",
			inputFilter: "",
			expectedFilter: &StringFilterOperationInput{
				Eq: optionalString("network1"),
			},
		},
		{
			name:        "Prefix name",
			inputName:   "name",
			inputFilter: "_prefix",
			expectedFilter: &StringFilterOperationInput{
				StartsWith: optionalString("name"),
			},
		},
		{
			name:           "Both name and filter empty",
			inputName:      "",
			inputFilter:    "",
			expectedFilter: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := NewResourceFilterInput(c.inputName, c.inputFilter)

			assert.Equal(t, c.expectedFilter, result.Name)
		})
	}
}

func TestReadResources_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadResources
		expected bool
	}{
		{
			name: "No edges - resources should be empty",
			query: ReadResources{
				Resources: Resources{
					PaginatedResource: PaginatedResource[*ResourceEdge]{
						Edges: nil, // No edges
					},
				},
			},
			expected: true,
		},
		{
			name: "Empty edges list - resources should be empty",
			query: ReadResources{
				Resources: Resources{
					PaginatedResource: PaginatedResource[*ResourceEdge]{
						Edges: []*ResourceEdge{}, // Empty edges list
					},
				},
			},
			expected: true,
		},
		{
			name: "Edges present - resources should not be empty",
			query: ReadResources{
				Resources: Resources{
					PaginatedResource: PaginatedResource[*ResourceEdge]{
						Edges: []*ResourceEdge{
							{
								Node: &ResourceNode{
									IDName: IDName{
										ID:   "123",
										Name: "TestResource",
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.IsEmpty())
		})
	}
}

func TestReadFullResources_IsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadFullResources
		expected bool
	}{
		{
			name: "No edges - resources should be empty",
			query: ReadFullResources{
				FullResources: FullResources{
					PaginatedResource: PaginatedResource[*FullResourceEdge]{
						Edges: nil, // No edges
					},
				},
			},
			expected: true,
		},
		{
			name: "Empty edges list - resources should be empty",
			query: ReadFullResources{
				FullResources: FullResources{
					PaginatedResource: PaginatedResource[*FullResourceEdge]{
						Edges: []*FullResourceEdge{}, // Empty edges list
					},
				},
			},
			expected: true,
		},
		{
			name: "Edges present - resources should not be empty",
			query: ReadFullResources{
				FullResources: FullResources{
					PaginatedResource: PaginatedResource[*FullResourceEdge]{
						Edges: []*FullResourceEdge{
							{
								Node: &gqlResource{
									ResourceNode: ResourceNode{
										IDName: IDName{
											ID:   "123",
											Name: "TestResource",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.IsEmpty())
		})
	}
}

func TestReadFullResources_ToModel(t *testing.T) {
	cases := []struct {
		name     string
		query    ReadFullResources
		expected []*model.Resource
	}{
		{
			name: "No edges - should return empty list",
			query: ReadFullResources{
				FullResources: FullResources{
					PaginatedResource: PaginatedResource[*FullResourceEdge]{
						Edges: nil, // No edges
					},
				},
			},
			expected: []*model.Resource{},
		},
		{
			name: "Empty edges list - should return empty list",
			query: ReadFullResources{
				FullResources: FullResources{
					PaginatedResource: PaginatedResource[*FullResourceEdge]{
						Edges: []*FullResourceEdge{}, // Empty edges list
					},
				},
			},
			expected: []*model.Resource{},
		},
		{
			name: "Edges present - should map to model.Resource",
			query: ReadFullResources{
				FullResources: FullResources{
					PaginatedResource: PaginatedResource[*FullResourceEdge]{
						Edges: []*FullResourceEdge{
							{
								Node: &gqlResource{
									ResourceNode: ResourceNode{
										IDName: IDName{
											ID:   "123",
											Name: "Resource A",
										},
									},
								},
							},
							{
								Node: &gqlResource{
									ResourceNode: ResourceNode{
										IDName: IDName{
											ID:   "456",
											Name: "Resource B",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []*model.Resource{
				{
					ID:   "123",
					Name: "Resource A",
					Protocols: &model.Protocols{
						TCP: &model.Protocol{
							Policy: model.PolicyAllowAll,
						},
						UDP: &model.Protocol{
							Policy: model.PolicyAllowAll,
						},
						AllowIcmp: true,
					},
					IsVisible:                optionalBool(false),
					IsBrowserShortcutEnabled: optionalBool(false),
				},
				{
					ID:   "456",
					Name: "Resource B",
					Protocols: &model.Protocols{
						TCP: &model.Protocol{
							Policy: model.PolicyAllowAll,
						},
						UDP: &model.Protocol{
							Policy: model.PolicyAllowAll,
						},
						AllowIcmp: true,
					},
					IsVisible:                optionalBool(false),
					IsBrowserShortcutEnabled: optionalBool(false),
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.query.ToModel())
		})
	}
}
