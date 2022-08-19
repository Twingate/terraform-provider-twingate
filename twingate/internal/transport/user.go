package transport

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

type readUsersQuery struct {
	Users *struct {
		Edges []*struct {
			Node *gqlUser
		}
	}
}

func (client *Client) ReadUsers(ctx context.Context) ([]*User, error) {
	response := readUsersQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readUsers", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, "All")
	}

	if response.Users == nil {
		return nil, nil
	}

	users := make([]*User, 0, len(response.Users.Edges))

	for _, item := range response.Users.Edges {
		if item == nil || item.Node == nil {
			continue
		}

		user := item.Node

		users = append(users, &User{
			ID:        user.ID.(string),
			FirstName: string(user.FirstName),
			LastName:  string(user.LastName),
			Email:     string(user.Email),
			Role:      string(user.Role),
		})
	}

	return users, nil
}

type readUserQuery struct {
	User *gqlUser `graphql:"user(id: $id)"`
}

func (client *Client) ReadUser(ctx context.Context, userID string) (*User, error) {
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
