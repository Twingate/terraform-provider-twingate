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
		{name: "bypass", input: strPtr(model.RoutingModeBypassTwingate), expected: &bypassVal},
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
	autoLock := model.AccessPolicyModeAutoLock
	duration := "2h"
	approvalMode := model.ApprovalModeManual
	seconds := int64(7200)
	approval := AccessApprovalMode(approvalMode)

	cases := []struct {
		name            string
		serviceAccounts []string
		groups          []model.AccessGroup
		expected        []AccessInput
	}{
		{
			name:     "nil access - returns empty slice",
			expected: []AccessInput{},
		},
		{
			name:   "group without security policy or access policy",
			groups: []model.AccessGroup{{GroupID: "group-1"}},
			expected: []AccessInput{
				{PrincipalID: "group-1", AccessPolicy: &AccessPolicyInput{Mode: model.AccessPolicyModeManual}},
			},
		},
		{
			name:   "group with security policy",
			groups: []model.AccessGroup{{GroupID: "group-1", SecurityPolicyID: strPtr("sp-1")}},
			expected: []AccessInput{
				{PrincipalID: "group-1", SecurityPolicyID: strPtr("sp-1"), AccessPolicy: &AccessPolicyInput{Mode: model.AccessPolicyModeManual}},
			},
		},
		{
			name: "group with access policy",
			groups: []model.AccessGroup{
				{GroupID: "group-1", AccessPolicy: &model.AccessPolicy{Mode: &autoLock, Duration: &duration, ApprovalMode: &approvalMode}},
			},
			expected: []AccessInput{
				{
					PrincipalID:  "group-1",
					ApprovalMode: &approval,
					AccessPolicy: &AccessPolicyInput{Mode: model.AccessPolicyModeAutoLock, DurationSeconds: &seconds},
				},
			},
		},
		{
			name:            "service accounts only",
			serviceAccounts: []string{"sa-1", "sa-2"},
			expected: []AccessInput{
				{PrincipalID: "sa-1"},
				{PrincipalID: "sa-2"},
			},
		},
		{
			name:            "service accounts go before groups",
			serviceAccounts: []string{"sa-1"},
			groups:          []model.AccessGroup{{GroupID: "group-1"}, {GroupID: "group-2", SecurityPolicyID: strPtr("sp-2")}},
			expected: []AccessInput{
				{PrincipalID: "sa-1"},
				{PrincipalID: "group-1", AccessPolicy: &AccessPolicyInput{Mode: model.AccessPolicyModeManual}},
				{PrincipalID: "group-2", SecurityPolicyID: strPtr("sp-2"), AccessPolicy: &AccessPolicyInput{Mode: model.AccessPolicyModeManual}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, NewAccessInputs(c.serviceAccounts, c.groups))
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestResourceVars(t *testing.T) {
	visible := true
	base := model.BaseResource{
		Name:            "resource",
		Address:         "example.com",
		RemoteNetworkID: "network-id",
		IsVisible:       &visible,
		Tags:            map[string]string{"env": "prod"},
	}

	baseKeys := []string{"name", "address", "remoteNetworkId", "isVisible", "alias", "securityPolicyId", "tags", "accessPolicy", "approvalMode"}

	t.Run("base", func(t *testing.T) {
		vars := newVars(baseResourceVars(base)...)

		assert.Len(t, vars, len(baseKeys))

		for _, key := range baseKeys {
			assert.Contains(t, vars, key)
		}

		assert.Equal(t, "resource", vars["name"])
		assert.Equal(t, "example.com", vars["address"])
	})

	t.Run("PAM adds the gateway and resource specific variables", func(t *testing.T) {
		vars := pamVars(model.PAMResource{BaseResource: base, GatewayID: "gateway-id"}, gqlVar(int64(22), "port"))

		assert.Len(t, vars, len(baseKeys)+2)
		assert.Contains(t, vars, "gatewayId")
		assert.Equal(t, int64(22), vars["port"])
	})
}
