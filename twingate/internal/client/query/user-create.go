package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

type CreateUser struct {
	UserEntityResponse `graphql:"userCreate(email: $email, firstName: $firstName, lastName: $lastName, role: $role, shouldSendInvite: $shouldSendInvite)"`
}

type UserEntityResponse struct {
	Entity *gqlUser
	OkError
}

func (q CreateUser) ToModel() *model.User {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}

func (q CreateUser) IsEmpty() bool {
	return q.Entity == nil
}
