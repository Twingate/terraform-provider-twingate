package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	cases := []struct {
		user     model.User
		expected interface{}
	}{
		{
			user: model.User{},
			expected: map[string]interface{}{
				"id":         "",
				"first_name": "",
				"last_name":  "",
				"email":      "",
				"is_admin":   false,
				"role":       "",
			},
		},
		{
			user: model.User{
				ID:        "1",
				FirstName: "John",
				LastName:  "White",
				Email:     "john@white.com",
				Role:      "ADMIN",
			},
			expected: map[string]interface{}{
				"id":         "1",
				"first_name": "John",
				"last_name":  "White",
				"email":      "john@white.com",
				"is_admin":   true,
				"role":       "ADMIN",
			},
		},
		{
			user: model.User{
				ID:        "2",
				FirstName: "Hue",
				LastName:  "Black",
				Email:     "hue@black.com",
				Role:      "USER",
			},
			expected: map[string]interface{}{
				"id":         "2",
				"first_name": "Hue",
				"last_name":  "Black",
				"email":      "hue@black.com",
				"is_admin":   false,
				"role":       "USER",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.user.ToTerraform())
		})
	}
}
