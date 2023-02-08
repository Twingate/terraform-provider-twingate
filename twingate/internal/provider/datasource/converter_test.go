package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverterConnectorsToTerraform(t *testing.T) {
	cases := []struct {
		input    []*model.Connector
		expected []interface{}
	}{
		{
			input:    nil,
			expected: []interface{}{},
		},
		{
			input:    []*model.Connector{},
			expected: []interface{}{},
		},
		{
			input: []*model.Connector{
				{ID: "connector-id", Name: "connector-name", NetworkID: "network-id"},
			},
			expected: []interface{}{
				map[string]interface{}{
					"id":                "connector-id",
					"name":              "connector-name",
					"remote_network_id": "network-id",
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
		expected []interface{}
	}{
		{
			input:    nil,
			expected: []interface{}{},
		},
		{
			input:    []*model.Group{},
			expected: []interface{}{},
		},
		{
			input: []*model.Group{
				{ID: "group-id", Name: "group-name", Type: model.GroupTypeManual, IsActive: true, SecurityPolicyID: "policy-id"},
			},
			expected: []interface{}{
				map[string]interface{}{
					"id":                 "group-id",
					"name":               "group-name",
					"type":               "MANUAL",
					"is_active":          true,
					"security_policy_id": "policy-id",
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
				{ID: "user-id", FirstName: "Name", LastName: "Last", Email: "user@email.com", Role: "USER"},
				{ID: "admin-id", FirstName: "Admin", LastName: "Last", Email: "admin@email.com", Role: "ADMIN"},
			},
			expected: []interface{}{
				map[string]interface{}{
					"id":         "user-id",
					"first_name": "Name",
					"last_name":  "Last",
					"email":      "user@email.com",
					"is_admin":   false,
					"role":       "USER",
				},
				map[string]interface{}{
					"id":         "admin-id",
					"first_name": "Admin",
					"last_name":  "Last",
					"email":      "admin@email.com",
					"is_admin":   true,
					"role":       "ADMIN",
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
					"id":                "resource-id",
					"name":              "name",
					"address":           "address",
					"remote_network_id": "network-id",
					"protocols":         emptySlice,
				},
				map[string]interface{}{
					"id":                "resource-1",
					"name":              "",
					"address":           "",
					"remote_network_id": "",
					"protocols": []interface{}{
						map[string]interface{}{
							"allow_icmp": true,
							"tcp": []interface{}{
								map[string]interface{}{
									"policy": model.PolicyRestricted,
									"ports":  []string{"8000-8080"},
								},
							},
							"udp": []interface{}{
								map[string]interface{}{
									"policy": model.PolicyRestricted,
									"ports":  emptyStringSlice,
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
					"id":           "service-account-id",
					"name":         "service-account-name",
					"resource_ids": []string{"res-1", "res-2"},
					"key_ids":      []string{"key-1", "key-2"},
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
					"id":   "policy-id",
					"name": "policy-name",
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
