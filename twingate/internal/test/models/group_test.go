package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGroupModel(t *testing.T) {
	cases := []struct {
		group model.Group

		expectedName string
		expectedID   string
		expected     interface{}
	}{
		{
			group: model.Group{},
			expected: map[string]interface{}{
				"id":                 "",
				"name":               "",
				"type":               "",
				"is_active":          false,
				"security_policy_id": "",
			},
		},
		{
			group: model.Group{
				ID:               "id",
				Name:             "name",
				Type:             "type",
				IsActive:         true,
				SecurityPolicyID: "policy-id",
			},
			expectedID:   "id",
			expectedName: "name",
			expected: map[string]interface{}{
				"id":                 "id",
				"name":               "name",
				"type":               "type",
				"is_active":          true,
				"security_policy_id": "policy-id",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expectedID, c.group.GetID())
			assert.Equal(t, c.expectedName, c.group.GetName())
			assert.Equal(t, c.expected, c.group.ToTerraform())
		})
	}
}
