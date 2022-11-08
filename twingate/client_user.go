package twingate

import (
	"context"

	"github.com/twingate/go-graphql-client"
)

const (
	userResourceName = "user"
	adminRole        = "ADMIN"
)

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

type gqlUser struct {
	ID        graphql.ID
	FirstName graphql.String
	LastName  graphql.String
	Email     graphql.String
	Role      graphql.String
}

type Users struct {
	PaginatedResource[*UserEdge]
}

func (u *Users) toList() []*User {
	return toList[*UserEdge, *User](u.Edges, func(edge *UserEdge) *User {
		user := edge.Node

		return &User{
			ID:        user.ID.(string),
			FirstName: string(user.FirstName),
			LastName:  string(user.LastName),
			Email:     string(user.Email),
			Role:      string(user.Role),
		}
	})
}

type UserEdge struct {
	Node *gqlUser
}

type readUsersQuery struct {
	Users Users
}

func (client *Client) readUsers(ctx context.Context) ([]*User, error) {
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

	return response.Users.toList(), nil
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

func (client *Client) readUser(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", userResourceName)
	}

	variables := map[string]interface{}{
		"id": userID,
	}

	response := readUserQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readUser", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, userID)
	}

	if response.User == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", userResourceName, userID)
	}

	return &User{
		ID:        response.User.ID.(string),
		FirstName: string(response.User.FirstName),
		LastName:  string(response.User.LastName),
		Email:     string(response.User.Email),
		Role:      string(response.User.Role),
	}, nil
}
