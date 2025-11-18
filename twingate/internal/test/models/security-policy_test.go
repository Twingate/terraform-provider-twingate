package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestSecurityPolicyModel(t *testing.T) {
	cases := []struct {
		policy model.SecurityPolicy

		expected any
	}{
		{
			policy: model.SecurityPolicy{},
			expected: map[string]any{
				attr.ID:   "",
				attr.Name: "",
			},
		},
		{
			policy: model.SecurityPolicy{
				ID:   "id",
				Name: "name",
			},
			expected: map[string]any{
				attr.ID:   "id",
				attr.Name: "name",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.policy.ToTerraform())
		})
	}
}
