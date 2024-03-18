package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestServiceAccountModel(t *testing.T) {
	cases := []struct {
		remoteNetwork model.ServiceAccount

		expectedName string
		expectedID   string
	}{
		{
			remoteNetwork: model.ServiceAccount{},
		},
		{
			remoteNetwork: model.ServiceAccount{
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

func TestServiceAccountToTerraform(t *testing.T) {
	var emptyStringSlice []string

	cases := []struct {
		serviceAccount model.ServiceAccount

		expected interface{}
	}{
		{
			serviceAccount: model.ServiceAccount{},
			expected: map[string]interface{}{
				attr.ID:          "",
				attr.Name:        "",
				attr.ResourceIDs: emptyStringSlice,
				attr.KeyIDs:      emptyStringSlice,
			},
		},
		{
			serviceAccount: model.ServiceAccount{
				ID:        "service-id",
				Name:      "service-name",
				Resources: []string{"res-1"},
				Keys:      []string{"key-1"},
			},
			expected: map[string]interface{}{
				attr.ID:          "service-id",
				attr.Name:        "service-name",
				attr.ResourceIDs: []string{"res-1"},
				attr.KeyIDs:      []string{"key-1"},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.serviceAccount.ToTerraform())
		})
	}
}
