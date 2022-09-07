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
