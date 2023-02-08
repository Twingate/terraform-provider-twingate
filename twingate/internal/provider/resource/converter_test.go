package resource

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConvertProtocol(t *testing.T) {

	cases := []struct {
		input       []interface{}
		expected    *model.Protocol
		expectedErr error
	}{
		{},
		{
			input: []interface{}{
				map[string]interface{}{
					"policy": model.PolicyAllowAll,
					"ports": []interface{}{
						"-",
					},
				},
			},
			expectedErr: errors.New("failed to parse protocols port range"),
		},
		{
			input: []interface{}{
				map[string]interface{}{
					"policy": model.PolicyRestricted,
					"ports": []interface{}{
						"80-88",
					},
				},
			},
			expected: &model.Protocol{
				Policy: model.PolicyRestricted,
				Ports: []*model.PortRange{
					{Start: 80, End: 88},
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			protocol, err := convertProtocol(c.input)

			assert.Equal(t, c.expected, protocol)
			if c.expectedErr != nil {
				assert.ErrorContains(t, err, c.expectedErr.Error())
			}

		})
	}

}

func TestConvertPortsRangeToMap(t *testing.T) {
	cases := []struct {
		portsRange []*model.PortRange
		expected   map[int32]struct{}
	}{
		{
			portsRange: nil,
			expected:   map[int32]struct{}{},
		},
		{
			portsRange: []*model.PortRange{
				{
					Start: 70,
					End:   70,
				},
				{
					Start: 81,
					End:   85,
				},
			},
			expected: map[int32]struct{}{
				70: {},
				81: {},
				82: {},
				83: {},
				84: {},
				85: {},
			},
		},
		{
			portsRange: []*model.PortRange{
				{
					Start: 80,
					End:   83,
				},
				{
					Start: 81,
					End:   85,
				},
				{
					Start: 81,
					End:   82,
				},
			},
			expected: map[int32]struct{}{
				80: {},
				81: {},
				82: {},
				83: {},
				84: {},
				85: {},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertPortsRangeToMap(c.portsRange)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestEqualPorts(t *testing.T) {
	cases := []struct {
		inputA   []interface{}
		inputB   []interface{}
		expected bool
	}{
		{
			inputA:   []interface{}{""},
			inputB:   []interface{}{""},
			expected: false,
		},
		{
			inputA:   []interface{}{"80"},
			inputB:   []interface{}{""},
			expected: false,
		},
		{
			inputA:   []interface{}{"80"},
			inputB:   []interface{}{"90"},
			expected: false,
		},
		{
			inputA:   []interface{}{"80"},
			inputB:   []interface{}{"80"},
			expected: true,
		},
		{
			inputA:   []interface{}{"80-81"},
			inputB:   []interface{}{"80", "81"},
			expected: true,
		},
		{
			inputA:   []interface{}{"80-81", "70"},
			inputB:   []interface{}{"70", "80", "81"},
			expected: true,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := equalPorts(c.inputA, c.inputB)
			assert.Equal(t, c.expected, actual)
		})
	}
}
