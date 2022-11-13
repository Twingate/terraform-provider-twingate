package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRemoteNetworkModel(t *testing.T) {
	cases := []struct {
		remoteNetwork model.RemoteNetwork

		expectedName string
		expectedID   string
	}{
		{
			remoteNetwork: model.RemoteNetwork{},
		},
		{
			remoteNetwork: model.RemoteNetwork{
				ID:   "id",
				Name: "name",
			},
			expectedID:   "id",
			expectedName: "name",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expectedID, c.remoteNetwork.GetID())
			assert.Equal(t, c.expectedName, c.remoteNetwork.GetName())
		})
	}
}
