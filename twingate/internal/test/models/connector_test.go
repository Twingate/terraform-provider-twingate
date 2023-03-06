package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConnectorModel(t *testing.T) {
	var (
		boolTrue  = true
		boolFalse = false
	)

	cases := []struct {
		connector model.Connector

		expectedName string
		expectedID   string
		expected     interface{}
	}{
		{
			connector: model.Connector{
				StatusUpdatesEnabled: &boolFalse,
			},
			expected: map[string]interface{}{
				"id":                      "",
				"name":                    "",
				"remote_network_id":       "",
				attr.StatusUpdatesEnabled: false,
			},
		},
		{
			connector: model.Connector{
				ID:                   "id",
				Name:                 "name",
				NetworkID:            "network-id",
				StatusUpdatesEnabled: &boolTrue,
			},
			expectedID:   "id",
			expectedName: "name",
			expected: map[string]interface{}{
				"id":                      "id",
				"name":                    "name",
				"remote_network_id":       "network-id",
				attr.StatusUpdatesEnabled: true,
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expectedID, c.connector.GetID())
			assert.Equal(t, c.expectedName, c.connector.GetName())
			assert.Equal(t, c.expected, c.connector.ToTerraform())
		})
	}
}
