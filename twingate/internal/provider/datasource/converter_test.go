package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
				{ID: "connector-id", Name: "connector-name", NetworkID: "network-id", StatusUpdatesEnabled: &boolTrue, State: "ALIVE"},
			},
			expected: []connectorModel{
				{
					ID:                   types.StringValue("connector-id"),
					Name:                 types.StringValue("connector-name"),
					RemoteNetworkID:      types.StringValue("network-id"),
					StatusUpdatesEnabled: types.BoolValue(true),
					State:                types.StringValue("ALIVE"),
					Hostname:             types.StringValue(""),
					Version:              types.StringValue(""),
					PublicIP:             types.StringValue(""),
					PrivateIPs:           types.SetValueMust(types.StringType, []attr.Value{}),
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
		expected []userModel
	}{
		{
			input:    nil,
			expected: []userModel{},
		},
		{
			input:    []*model.User{},
			expected: []userModel{},
		},
		{
			input: []*model.User{
				{ID: "user-id", FirstName: "Name", LastName: "Last", Email: "user@email.com", Role: "USER", Type: "SYNCED"},
				{ID: "admin-id", FirstName: "Admin", LastName: "Last", Email: "admin@email.com", Role: model.UserRoleAdmin, Type: "MANUAL"},
			},
			expected: []userModel{
				{
					ID:        types.StringValue("user-id"),
					FirstName: types.StringValue("Name"),
					LastName:  types.StringValue("Last"),
					Email:     types.StringValue("user@email.com"),
					Role:      types.StringValue("USER"),
					Type:      types.StringValue("SYNCED"),
				},
				{
					ID:        types.StringValue("admin-id"),
					FirstName: types.StringValue("Admin"),
					LastName:  types.StringValue("Last"),
					Email:     types.StringValue("admin@email.com"),
					Role:      types.StringValue(model.UserRoleAdmin),
					Type:      types.StringValue("MANUAL"),
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
	cases := []struct {
		input    []*model.Resource
		expected []resourceModel
	}{
		{
			input:    nil,
			expected: []resourceModel{},
		},
		{
			input:    []*model.Resource{},
			expected: []resourceModel{},
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
							Ports:  []*model.PortRange{},
						},
					},
				},
			},
			expected: []resourceModel{
				{
					ID:              types.StringValue("resource-id"),
					Name:            types.StringValue("name"),
					Address:         types.StringValue("address"),
					RemoteNetworkID: types.StringValue("network-id"),
					Protocols:       nil,
					Tags:            types.MapNull(types.StringType),
				},
				{
					ID:              types.StringValue("resource-1"),
					Name:            types.StringValue(""),
					Address:         types.StringValue(""),
					RemoteNetworkID: types.StringValue(""),
					Protocols: &protocolsModel{
						AllowIcmp: types.BoolValue(true),
						TCP: &protocolModel{
							Policy: types.StringValue(model.PolicyRestricted),
							Ports:  []types.String{types.StringValue("8000-8080")},
						},
						UDP: &protocolModel{
							Policy: types.StringValue(model.PolicyRestricted),
							Ports:  []types.String{},
						},
					},
					Tags: types.MapNull(types.StringType),
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
		expected []serviceAccountModel
	}{
		{
			input:    nil,
			expected: []serviceAccountModel{},
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
			expected: []serviceAccountModel{
				{
					ID:          types.StringValue("service-account-id"),
					Name:        types.StringValue("service-account-name"),
					ResourceIDs: []types.String{types.StringValue("res-1"), types.StringValue("res-2")},
					KeyIDs:      []types.String{types.StringValue("key-1"), types.StringValue("key-2")},
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
		expected []securityPolicyModel
	}{
		{
			input:    nil,
			expected: []securityPolicyModel{},
		},
		{
			input: []*model.SecurityPolicy{
				{
					ID:   "policy-id",
					Name: "policy-name",
				},
			},
			expected: []securityPolicyModel{
				{
					ID:   types.StringValue("policy-id"),
					Name: types.StringValue("policy-name"),
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

func TestConvertRemoteNetworksToTerraform(t *testing.T) {
	cases := []struct {
		input    []*model.RemoteNetwork
		expected []remoteNetworkModel
	}{
		{
			input:    nil,
			expected: []remoteNetworkModel{},
		},
		{
			input: []*model.RemoteNetwork{
				{
					ID:       "network-id",
					Name:     "network-name",
					Location: "network-location",
					Type:     "network-type",
				},
			},
			expected: []remoteNetworkModel{
				{
					ID:       types.StringValue("network-id"),
					Name:     types.StringValue("network-name"),
					Location: types.StringValue("network-location"),
					Type:     types.StringValue("network-type"),
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertRemoteNetworksToTerraform(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestConvertDomainsToTerraform(t *testing.T) {
	cases := []struct {
		input    []string
		expected *domainsModel
	}{
		{
			input: nil,
			expected: &domainsModel{
				Domains: types.SetValueMust(types.StringType, []attr.Value{}),
			},
		},
		{
			input: []string{
				"domain-1",
				"domain-2",
			},
			expected: &domainsModel{
				Domains: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("domain-1"),
					types.StringValue("domain-2"),
				}),
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertDomainsToTerraform(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}
