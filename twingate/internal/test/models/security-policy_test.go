package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestSecurityPolicyModel(t *testing.T) {
	cases := []struct {
		policy model.SecurityPolicy

		expected interface{}
	}{
		{
			policy: model.SecurityPolicy{},
			expected: map[string]interface{}{
				"id":   "",
				"name": "",
			},
		},
		{
			policy: model.SecurityPolicy{
				ID:   "id",
				Name: "name",
			},
			expected: map[string]interface{}{
				"id":   "id",
				"name": "name",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.policy.ToTerraform())
		})
	}
}
