package transport

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const userResourceName = "user"

type gqlUser struct {
	ID        graphql.ID
	FirstName graphql.String
	LastName  graphql.String
	Email     graphql.String
	Role      graphql.String
}

type gqlUsers struct {
	Edges []*struct {
		Node *gqlUser
	}
}

type readUsersQuery struct {
	Users gqlUsers
}

func (client *Client) ReadUsers(ctx context.Context) ([]*model.User, error) {
	response := readUsersQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readUsers", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, "All")
	}

	if len(response.Users.Edges) == 0 {
		return nil, nil
	}

	return response.Users.ToModel(), nil
}

type readUserQuery struct {
	User *gqlUser `graphql:"user(id: $id)"`
}

func (client *Client) ReadUser(ctx context.Context, userID string) (*model.User, error) {
	if userID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", userResourceName)
	}

	variables := newVars(gqlID(userID))
	response := readUserQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readUser", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, userID)
	}

	if response.User == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", userResourceName, userID)
	}

	return response.User.ToModel(), nil
}
