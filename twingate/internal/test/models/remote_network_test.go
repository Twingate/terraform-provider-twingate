package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRemoteNetworkModel(t *testing.T) {
	cases := []struct {
		remoteNetwork model.RemoteNetwork

		expected         interface{}
		expectedID       string
		expectedName     string
		expectedLocation string
	}{
		{
			remoteNetwork: model.RemoteNetwork{},
			expected: map[string]interface{}{
				attr.ID:       "",
				attr.Name:     "",
				attr.Location: "",
			},
		},
		{
			remoteNetwork: model.RemoteNetwork{
				ID:       "id",
				Name:     "name",
				Location: model.LocationGoogleCloud,
			},
			expected: map[string]interface{}{
				attr.ID:       "id",
				attr.Name:     "name",
				attr.Location: model.LocationGoogleCloud,
			},
			expectedID:       "id",
			expectedName:     "name",
			expectedLocation: model.LocationGoogleCloud,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.remoteNetwork.ToTerraform())
			assert.Equal(t, c.expectedID, c.remoteNetwork.GetID())
			assert.Equal(t, c.expectedName, c.remoteNetwork.GetName())
			assert.Equal(t, c.expectedLocation, c.remoteNetwork.Location)
		})
	}
}
