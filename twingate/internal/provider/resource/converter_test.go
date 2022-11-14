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
