package client

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewRoutingMode(t *testing.T) {
	through := model.RoutingModeThroughTwingate
	empty := ""

	bypassVal := RoutingMode(model.RoutingModeBypassTwingate)
	throughVal := RoutingMode(through)

	cases := []struct {
		name     string
		input    *string
		expected *RoutingMode
	}{
		{name: "nil", input: nil, expected: nil},
		{name: "empty", input: &empty, expected: nil},
		{name: "through", input: &through, expected: &throughVal},
		{name: "bypass", input: optionalString(model.RoutingModeBypassTwingate), expected: &bypassVal},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, newRoutingMode(c.input))
		})
	}
}

func TestNewPorts(t *testing.T) {
	var cases = []struct {
		ports    []*model.PortRange
		expected []*PortRangeInput
	}{
		{
			ports:    nil,
			expected: []*PortRangeInput{},
		},
		{
			ports: []*model.PortRange{
				{Start: 80, End: 90},
			},
			expected: []*PortRangeInput{
				{Start: 80, End: 90},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := newPorts(c.ports)

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestNewProtocol(t *testing.T) {
	var cases = []struct {
		protocol *model.Protocol
		expected *ProtocolInput
	}{
		{
			protocol: nil,
			expected: nil,
		},
		{
			protocol: &model.Protocol{
				Ports: []*model.PortRange{
					{Start: 80, End: 90},
				},
				Policy: model.PolicyRestricted,
			},
			expected: &ProtocolInput{
				Ports: []*PortRangeInput{
					{Start: 80, End: 90},
				},
				Policy: model.PolicyRestricted,
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := newProtocol(c.protocol)

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestNewAccessInputs(t *testing.T) {
	cases := []struct {
		name            string
		groups          []model.AccessGroup
		serviceAccounts []string
		expected        []AccessInput
	}{
		{
			name:     "nil groups - returns empty slice",
			groups:   nil,
			expected: []AccessInput{},
		},
		{
			name:     "empty groups - returns empty slice",
			groups:   []model.AccessGroup{},
			expected: []AccessInput{},
		},
		{
			name: "group without security policy or access policy",
			groups: []model.AccessGroup{
				{GroupID: "group-1"},
			},
			expected: []AccessInput{
				{PrincipalID: "group-1"},
			},
		},
		{
			name: "group with security policy",
			groups: []model.AccessGroup{
				{GroupID: "group-1", SecurityPolicyID: optionalString("sp-1")},
			},
			expected: []AccessInput{
				{PrincipalID: "group-1", SecurityPolicyID: optionalString("sp-1")},
			},
		},
		{
			name: "multiple groups",
			groups: []model.AccessGroup{
				{GroupID: "group-1"},
				{GroupID: "group-2", SecurityPolicyID: optionalString("sp-2")},
			},
			expected: []AccessInput{
				{PrincipalID: "group-1"},
				{PrincipalID: "group-2", SecurityPolicyID: optionalString("sp-2")},
			},
		},
		{
			name: "groups and service accounts",
			groups: []model.AccessGroup{
				{GroupID: "group-1", SecurityPolicyID: optionalString("sp-1")},
			},
			serviceAccounts: []string{"service-account-1"},
			expected: []AccessInput{
				{PrincipalID: "group-1", SecurityPolicyID: optionalString("sp-1")},
				{PrincipalID: "service-account-1"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := NewAccessInputs(c.groups, c.serviceAccounts)
			assert.Equal(t, len(c.expected), len(actual))
			for i := range c.expected {
				assert.Equal(t, c.expected[i].PrincipalID, actual[i].PrincipalID)
				assert.Equal(t, c.expected[i].SecurityPolicyID, actual[i].SecurityPolicyID)
			}
		})
	}
}
