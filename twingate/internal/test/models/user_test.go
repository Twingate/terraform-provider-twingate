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
				attr.Type:      "",
			},
		},
		{
			user: model.User{
				ID:        "1",
				FirstName: "John",
				LastName:  "White",
				Email:     "john@white.com",
				Role:      "ADMIN",
				Type:      "MANUAL",
			},
			expected: map[string]interface{}{
				attr.ID:        "1",
				attr.FirstName: "John",
				attr.LastName:  "White",
				attr.Email:     "john@white.com",
				attr.IsAdmin:   true,
				attr.Role:      "ADMIN",
				attr.Type:      "MANUAL",
			},
		},
		{
			user: model.User{
				ID:        "2",
				FirstName: "Hue",
				LastName:  "Black",
				Email:     "hue@black.com",
				Role:      "USER",
				Type:      "SYNCED",
			},
			expected: map[string]interface{}{
				attr.ID:        "2",
				attr.FirstName: "Hue",
				attr.LastName:  "Black",
				attr.Email:     "hue@black.com",
				attr.IsAdmin:   false,
				attr.Role:      "USER",
				attr.Type:      "SYNCED",
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.user.ToTerraform())
		})
	}
}

func TestUserGetID(t *testing.T) {
	cases := []struct {
		user     model.User
		expected string
	}{
		{
			user:     model.User{},
			expected: "",
		},
		{
			user: model.User{
				ID: "1",
			},
			expected: "1",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.user.GetID())
		})
	}
}

func TestUserGetName(t *testing.T) {
	cases := []struct {
		user     model.User
		expected string
	}{
		{
			user:     model.User{},
			expected: "",
		},
		{
			user: model.User{
				Email:     "user-mail",
				FirstName: "Twin",
				LastName:  "Gate",
			},
			expected: "user-mail",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.user.GetName())
		})
	}
}

func TestUserState(t *testing.T) {
	cases := []struct {
		user     model.User
		expected string
	}{
		{
			user: model.User{
				IsActive: false,
			},
			expected: model.UserStateDisabled,
		},
		{
			user: model.User{
				IsActive: true,
			},
			expected: model.UserStateActive,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.user.State())
		})
	}
}

func TestUserUpdateState(t *testing.T) {
	valTrue := true
	valFalse := false

	cases := []struct {
		user     model.UserUpdate
		expected string
	}{
		{
			user:     model.UserUpdate{},
			expected: "",
		},
		{
			user: model.UserUpdate{
				IsActive: &valTrue,
			},
			expected: model.UserStateActive,
		},
		{
			user: model.UserUpdate{
				IsActive: &valFalse,
			},
			expected: model.UserStateDisabled,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, c.user.State())
		})
	}
}
