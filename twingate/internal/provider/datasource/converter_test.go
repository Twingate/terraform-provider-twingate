package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverter(t *testing.T) {
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
