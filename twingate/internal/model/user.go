package model

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"

const (
	UserRoleAdmin   = "ADMIN"
	UserRoleDevops  = "DEVOPS"
	UserRoleSupport = "SUPPORT"
	UserRoleMember  = "MEMBER"

	UserStateActive   = "ACTIVE"
	UserStatePending  = "PENDING"
	UserStateDisabled = "DISABLED"

	UserTypeManual = "MANUAL"
	UserTypeSynced = "SYNCED"
)

//nolint:gochecknoglobals
var (
	UserRoles = []string{UserRoleAdmin, UserRoleDevops, UserRoleSupport, UserRoleMember}
	UserTypes = []string{UserTypeManual, UserTypeSynced}
)

type User struct {
	ID         string
	FirstName  string
	LastName   string
	Email      string
	Role       string
	Type       string
	SendInvite bool
	IsActive   bool
}

func (u User) GetID() string {
	return u.ID
}

func (u User) GetName() string {
	// that's used only in sweeper tests
	return u.Email
}

func (u User) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:        u.ID,
		attr.FirstName: u.FirstName,
		attr.LastName:  u.LastName,
		attr.Email:     u.Email,
		attr.Role:      u.Role,
		attr.Type:      u.Type,
	}
}

func (u User) State() string {
	if u.IsActive {
		return UserStateActive
	}

	return UserStateDisabled
}

type UserUpdate struct {
	ID        string
	FirstName *string
	LastName  *string
	Role      *string
	IsActive  *bool
}

func (u UserUpdate) State() string {
	if u.IsActive == nil {
		return ""
	}

	if *u.IsActive {
		return UserStateActive
	}

	return UserStateDisabled
}
