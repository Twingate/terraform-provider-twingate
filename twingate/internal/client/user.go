package client

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

type UserEdge struct {
	Node *gqlUser
}

type Users struct {
	PaginatedResource[*UserEdge]
}

type readUsersQuery struct {
	Users Users
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

	err = response.Users.fetchPages(ctx, client.readUsersAfter, nil)
	if err != nil {
		return nil, err
	}

	return response.Users.ToModel(), nil
}

type readUsersAfter struct {
	Users Users `graphql:"users(after: $usersEndCursor)"`
}

func (client *Client) readUsersAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*UserEdge], error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	variables["usersEndCursor"] = cursor
	response := readUsersAfter{}

	err := client.GraphqlClient.NamedQuery(ctx, "readUsers", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, "All")
	}

	if len(response.Users.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", userResourceName, "All")
	}

	return &response.Users.PaginatedResource, nil
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
