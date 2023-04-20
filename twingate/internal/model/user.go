package model

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"

const RoleAdmin = "ADMIN"

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Role      string
	Type      string
}

func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u User) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:        u.ID,
		attr.FirstName: u.FirstName,
		attr.LastName:  u.LastName,
		attr.Email:     u.Email,
		attr.IsAdmin:   u.IsAdmin(),
		attr.Role:      u.Role,
		attr.Type:      u.Type,
	}
}
