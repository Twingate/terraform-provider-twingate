package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

type ReadUser struct {
	User *gqlUser `graphql:"user(id: $id)"`
}

type gqlUser struct {
	ID        graphql.ID
	FirstName graphql.String
	LastName  graphql.String
	Email     graphql.String
	Role      graphql.String
}

func (u gqlUser) ToModel() *model.User {
	return &model.User{
		ID:        idToString(u.ID),
		FirstName: string(u.FirstName),
		LastName:  string(u.LastName),
		Email:     string(u.Email),
		Role:      string(u.Role),
	}
}

func (q ReadUser) ToModel() *model.User {
	if q.User == nil {
		return nil
	}

	return q.User.ToModel()
}
