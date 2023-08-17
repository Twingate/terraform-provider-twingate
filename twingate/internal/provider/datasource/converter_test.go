package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestConverterConnectorsToTerraform(t *testing.T) {
	boolTrue := true

	cases := []struct {
		input    []*model.Connector
		expected []connectorModel
	}{
		{
			input:    nil,
			expected: []connectorModel{},
		},
		{
			input:    []*model.Connector{},
			expected: []connectorModel{},
		},
		{
			input: []*model.Connector{
				{ID: "connector-id", Name: "connector-name", NetworkID: "network-id", StatusUpdatesEnabled: &boolTrue},
			},
			expected: []connectorModel{
				{
					Name:                 types.StringValue("connector-name"),
					RemoteNetworkID:      types.StringValue("network-id"),
					StatusUpdatesEnabled: types.BoolValue(true),
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertConnectorsToTerraform(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestConverterGroupsToTerraform(t *testing.T) {
	cases := []struct {
		input    []*model.Group
		expected []groupModel
	}{
		{
			input:    nil,
			expected: []groupModel{},
		},
		{
			input:    []*model.Group{},
			expected: []groupModel{},
		},
		{
			input: []*model.Group{
				{ID: "group-id", Name: "group-name", Type: model.GroupTypeManual, IsActive: true, SecurityPolicyID: "policy-id"},
			},
			expected: []groupModel{
				{
					ID:               types.StringValue("group-id"),
					Name:             types.StringValue("group-name"),
					Type:             types.StringValue(model.GroupTypeManual),
					SecurityPolicyID: types.StringValue("policy-id"),
					IsActive:         types.BoolValue(true),
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertGroupsToTerraform(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestConverterUsersToTerraform(t *testing.T) {
	cases := []struct {
		input    []*model.User
		expected []interface{}
	}{
		{
			input:    nil,
			expected: []interface{}{},
		},
		{
			input:    []*model.User{},
			expected: []interface{}{},
		},
		{
			input: []*model.User{
				{ID: "user-id", FirstName: "Name", LastName: "Last", Email: "user@email.com", Role: "USER", Type: "SYNCED"},
				{ID: "admin-id", FirstName: "Admin", LastName: "Last", Email: "admin@email.com", Role: model.UserRoleAdmin, Type: "MANUAL"},
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.ID:        "user-id",
					attr.FirstName: "Name",
					attr.LastName:  "Last",
					attr.Email:     "user@email.com",
					attr.IsAdmin:   false,
					attr.Role:      "USER",
					attr.Type:      "SYNCED",
				},
				map[string]interface{}{
					attr.ID:        "admin-id",
					attr.FirstName: "Admin",
					attr.LastName:  "Last",
					attr.Email:     "admin@email.com",
					attr.IsAdmin:   true,
					attr.Role:      model.UserRoleAdmin,
					attr.Type:      "MANUAL",
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertUsersToTerraform(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestConverterResourcesToTerraform(t *testing.T) {
	var emptySlice []interface{}
	var emptyStringSlice []string

	cases := []struct {
		input    []*model.Resource
		expected []interface{}
	}{
		{
			input:    nil,
			expected: []interface{}{},
		},
		{
			input:    []*model.Resource{},
			expected: []interface{}{},
		},
		{
			input: []*model.Resource{
				{ID: "resource-id", Name: "name", Address: "address", RemoteNetworkID: "network-id"},
				{
					ID: "resource-1",
					Protocols: &model.Protocols{
						AllowIcmp: true,
						TCP: &model.Protocol{
							Policy: model.PolicyRestricted,
							Ports: []*model.PortRange{
								{Start: 8000, End: 8080},
							},
						},
						UDP: &model.Protocol{
							Policy: model.PolicyRestricted,
						},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.ID:              "resource-id",
					attr.Name:            "name",
					attr.Address:         "address",
					attr.RemoteNetworkID: "network-id",
					attr.Protocols:       emptySlice,
				},
				map[string]interface{}{
					attr.ID:              "resource-1",
					attr.Name:            "",
					attr.Address:         "",
					attr.RemoteNetworkID: "",
					attr.Protocols: []interface{}{
						map[string]interface{}{
							attr.AllowIcmp: true,
							attr.TCP: []interface{}{
								map[string]interface{}{
									attr.Policy: model.PolicyRestricted,
									attr.Ports:  []string{"8000-8080"},
								},
							},
							attr.UDP: []interface{}{
								map[string]interface{}{
									attr.Policy: model.PolicyDenyAll,
									attr.Ports:  emptyStringSlice,
								},
							},
						},
					},
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertResourcesToTerraform(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestTerraformServicesDatasourceID(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "",
			expected: "all-services",
		},
		{
			input:    "hello",
			expected: "service-by-name-hello",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := terraformServicesDatasourceID(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestConvertServicesToTerraform(t *testing.T) {
	cases := []struct {
		input    []*model.ServiceAccount
		expected []interface{}
	}{
		{
			input:    nil,
			expected: []interface{}{},
		},
		{
			input: []*model.ServiceAccount{
				{
					ID:        "service-account-id",
					Name:      "service-account-name",
					Resources: []string{"res-1", "res-2"},
					Keys:      []string{"key-1", "key-2"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.ID:          "service-account-id",
					attr.Name:        "service-account-name",
					attr.ResourceIDs: []string{"res-1", "res-2"},
					attr.KeyIDs:      []string{"key-1", "key-2"},
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertServicesToTerraform(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestConvertSecurityPoliciesToTerraform(t *testing.T) {
	cases := []struct {
		input    []*model.SecurityPolicy
		expected []interface{}
	}{
		{
			input:    nil,
			expected: []interface{}{},
		},
		{
			input: []*model.SecurityPolicy{
				{
					ID:   "policy-id",
					Name: "policy-name",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.ID:   "policy-id",
					attr.Name: "policy-name",
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertSecurityPoliciesToTerraform(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}
