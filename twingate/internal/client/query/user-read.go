package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadUser struct {
	User *gqlUser `graphql:"user(id: $id)"`
}

func (q ReadUser) IsEmpty() bool {
	return q.User == nil
}

type gqlUser struct {
	ID        graphql.ID
	FirstName string
	LastName  string
	Email     string
	Role      string
}

func (u gqlUser) ToModel() *model.User {
	return &model.User{
		ID:        string(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Role:      u.Role,
	}
}

func (q ReadUser) ToModel() *model.User {
	if q.User == nil {
		return nil
	}

	return q.User.ToModel()
}
