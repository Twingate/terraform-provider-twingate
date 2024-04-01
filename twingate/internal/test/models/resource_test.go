package models

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewPortRange(t *testing.T) {
	invalidPortsRange := func(str ...string) error {
		port, input := str[0], str[0]
		if len(str) > 1 {
			port = str[1]
		}

		return fmt.Errorf("failed to parse protocols port range \"%s\": port `%s` is not a valid integer: strconv.ParseInt: parsing \"%s\": invalid syntax", input, port, port)
	}

	cases := []struct {
		input       string
		expected    *model.PortRange
		expectedErr error
	}{
		{
			input:    "80",
			expected: &model.PortRange{Start: 80, End: 80},
		},
		{
			input:    "80-90",
			expected: &model.PortRange{Start: 80, End: 90},
		},
		{
			input:       "",
			expectedErr: invalidPortsRange(""),
		},
		{
			input:       " ",
			expectedErr: invalidPortsRange(" "),
		},
		{
			input:       "foo",
			expectedErr: invalidPortsRange("foo"),
		},
		{
			input:       "80-",
			expectedErr: invalidPortsRange("80-", ""),
		},
		{
			input:       "-80",
			expectedErr: invalidPortsRange("-80", ""),
		},
		{
			input:       "80-90-100",
			expectedErr: errors.New("failed to parse protocols port range \"80-90-100\": port range expects 2 values"),
		},
		{
			input:       "80-70",
			expectedErr: errors.New("failed to parse protocols port range \"80-70\": ports 80, 70 needs to be in a rising sequence"),
		},
		{
			input:       "0-65536",
			expectedErr: errors.New("failed to parse protocols port range \"0-65536\": port 0 not in the range of 1-65535"),
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual, err := model.NewPortRange(c.input)

			assert.Equal(t, c.expected, actual)

			if c.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, c.expectedErr.Error())
			}
		})
	}
}

func TestResourceModel(t *testing.T) {
	var (
		emptySlice       []interface{}
		emptyStringSlice []string
	)

	cases := []struct {
		resource model.Resource

		expectedName string
		expectedID   string
		expected     interface{}
	}{
		{
			resource: model.Resource{},
			expected: map[string]interface{}{
				attr.ID:              "",
				attr.Name:            "",
				attr.Address:         "",
				attr.RemoteNetworkID: "",
				attr.Protocols:       emptySlice,
			},
		},
		{
			resource: model.Resource{
				ID:              "id",
				Name:            "name",
				Address:         "address",
				RemoteNetworkID: "network-id",
				Protocols: &model.Protocols{
					AllowIcmp: true,
					UDP: &model.Protocol{
						Policy: model.PolicyAllowAll,
					},
					TCP: &model.Protocol{
						Ports: []*model.PortRange{
							{Start: 80, End: 80},
						},
						Policy: model.PolicyRestricted,
					},
				},
			},
			expectedID:   "id",
			expectedName: "name",
			expected: map[string]interface{}{
				attr.ID:              "id",
				attr.Name:            "name",
				attr.Address:         "address",
				attr.RemoteNetworkID: "network-id",
				attr.Protocols: []interface{}{
					map[string]interface{}{
						attr.AllowIcmp: true,
						attr.TCP: []interface{}{
							map[string]interface{}{
								attr.Policy: "RESTRICTED",
								attr.Ports:  []string{"80"},
							},
						},
						attr.UDP: []interface{}{
							map[string]interface{}{
								attr.Policy: "ALLOW_ALL",
								attr.Ports:  emptyStringSlice,
							},
						},
					},
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expectedID, c.resource.GetID())
			assert.Equal(t, c.expectedName, c.resource.GetName())
			assert.Equal(t, c.expected, c.resource.ToTerraform())
		})
	}
}

func TestProtocolToTerraform(t *testing.T) {
	var emptySlice []interface{}
	var emptyStringSlice []string

	cases := []struct {
		protocol *model.Protocol

		expected interface{}
	}{
		{
			protocol: nil,
			expected: emptySlice,
		},
		{
			protocol: &model.Protocol{
				Policy: model.PolicyAllowAll,
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.Policy: "ALLOW_ALL",
					attr.Ports:  emptyStringSlice,
				},
			},
		},
		{
			protocol: &model.Protocol{
				Policy: model.PolicyRestricted,
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.Policy: "DENY_ALL",
					attr.Ports:  emptyStringSlice,
				},
			},
		},
		{
			protocol: &model.Protocol{
				Policy: model.PolicyRestricted,
				Ports: []*model.PortRange{
					{Start: 80, End: 80},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.Policy: "RESTRICTED",
					attr.Ports:  []string{"80"},
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.protocol.ToTerraform())
		})
	}
}

func TestProtocols(t *testing.T) {
	t.Run("Test Twingate Resource : Protocols", func(t *testing.T) {
		protocols := model.DefaultProtocols()

		assert.EqualValues(t, model.PolicyAllowAll, protocols.TCP.Policy)
		assert.EqualValues(t, model.PolicyAllowAll, protocols.UDP.Policy)
		assert.Nil(t, protocols.UDP.Ports)
		assert.Nil(t, protocols.TCP.Ports)

		port := &model.PortRange{Start: 1, End: 18000}
		protocols.TCP.Ports = append(protocols.TCP.Ports, port)
		protocols.UDP.Ports = append(protocols.UDP.Ports, port)
		udpPorts := protocols.UDP.PortsToString()
		tcpPorts := protocols.TCP.PortsToString()
		assert.EqualValues(t, "1-18000", tcpPorts[0])
		assert.EqualValues(t, "1-18000", udpPorts[0])
	})
}

func TestResourceAccessToTerraform(t *testing.T) {
	cases := []struct {
		resource model.Resource

		expected []interface{}
	}{
		{
			resource: model.Resource{},
			expected: nil,
		},
		{
			resource: model.Resource{
				GroupsAccess: []model.AccessGroup{
					{GroupID: "group-1"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.GroupIDs: []string{"group-1"},
				},
			},
		},
		{
			resource: model.Resource{
				ServiceAccounts: []string{"service-1"},
				IsAuthoritative: true,
			},
			expected: []interface{}{
				map[string]interface{}{
					attr.ServiceAccountIDs: []string{"service-1"},
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.resource.AccessToTerraform())
		})
	}
}
