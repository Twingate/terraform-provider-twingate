package query

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

type UserRole string
type UserStateUpdateInput string

func NewUserStateUpdateInput(val string) *UserStateUpdateInput {
	if val == "" {
		return nil
	}

	state := UserStateUpdateInput(val)

	return &state
}

type UpdateUser struct {
	UserEntityResponse `graphql:"userDetailsUpdate(id: $id, firstName: $firstName, lastName: $lastName, state: $state)"`
}

func (q UpdateUser) ToModel() *model.User {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}

func (q UpdateUser) IsEmpty() bool {
	return q.Entity == nil
}

type UpdateUserRole struct {
	UserEntityResponse `graphql:"userRoleUpdate(id: $id, role: $role)"`
}

func (q UpdateUserRole) ToModel() *model.User {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}

func (q UpdateUserRole) IsEmpty() bool {
	return q.Entity == nil
}
