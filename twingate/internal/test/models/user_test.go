package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
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
				attr.ID:        "",
				attr.FirstName: "",
				attr.LastName:  "",
				attr.Email:     "",
				attr.IsAdmin:   false,
				attr.Role:      "",
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
				attr.ID:        "1",
				attr.FirstName: "John",
				attr.LastName:  "White",
				attr.Email:     "john@white.com",
				attr.IsAdmin:   true,
				attr.Role:      "ADMIN",
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
				attr.ID:        "2",
				attr.FirstName: "Hue",
				attr.LastName:  "Black",
				attr.Email:     "hue@black.com",
				attr.IsAdmin:   false,
				attr.Role:      "USER",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.user.ToTerraform())
		})
	}
}
