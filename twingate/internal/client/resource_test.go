package client

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
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
