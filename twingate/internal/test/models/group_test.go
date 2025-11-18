package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGroupModel(t *testing.T) {
	cases := []struct {
		group model.Group

		expectedName string
		expectedID   string
		expected     any
	}{
		{
			group: model.Group{},
			expected: map[string]any{
				attr.ID:               "",
				attr.Name:             "",
				attr.Type:             "",
				attr.IsActive:         false,
				attr.SecurityPolicyID: "",
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
			expected: map[string]any{
				attr.ID:               "id",
				attr.Name:             "name",
				attr.Type:             "type",
				attr.IsActive:         true,
				attr.SecurityPolicyID: "policy-id",
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
