package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
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
		expected     any
	}{
		{
			connector: model.Connector{
				StatusUpdatesEnabled: &boolFalse,
			},
			expected: map[string]any{
				attr.ID:                   "",
				attr.Name:                 "",
				attr.RemoteNetworkID:      "",
				attr.StatusUpdatesEnabled: false,
				attr.State:                "",
				attr.Version:              "",
				attr.Hostname:             "",
				attr.PublicIP:             "",
				attr.PrivateIPs:           []string(nil),
			},
		},
		{
			connector: model.Connector{
				ID:                   "id",
				Name:                 "name",
				NetworkID:            "network-id",
				StatusUpdatesEnabled: &boolTrue,
				State:                "DEAD_NO_HEARTBEAT",
				Version:              "0.1",
				Hostname:             "127.0.0.1",
				PublicIP:             "127.0.0.1",
				PrivateIPs:           []string{"127.0.0.1"},
			},
			expectedID:   "id",
			expectedName: "name",
			expected: map[string]any{
				attr.ID:                   "id",
				attr.Name:                 "name",
				attr.RemoteNetworkID:      "network-id",
				attr.StatusUpdatesEnabled: true,
				attr.State:                "DEAD_NO_HEARTBEAT",
				attr.Version:              "0.1",
				attr.Hostname:             "127.0.0.1",
				attr.PublicIP:             "127.0.0.1",
				attr.PrivateIPs:           []string{"127.0.0.1"},
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
