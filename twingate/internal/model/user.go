package model

const adminRole = "ADMIN"

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Role      string
}

func (u User) IsAdmin() bool {
	return u.Role == adminRole
}

func (u User) ToTerraform() interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"email":      u.Email,
		"is_admin":   u.IsAdmin(),
		"role":       u.Role,
	}
}
