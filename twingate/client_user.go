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

type readUserQuery struct {
	User *struct {
		ID        graphql.ID
		FirstName graphql.String
		LastName  graphql.String
		Email     graphql.String
		Role      graphql.String
	} `graphql:"user(id: $id)"`
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
