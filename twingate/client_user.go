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

type readUsersQuery struct {
	Users *struct {
		Edges []*struct {
			Node *gqlUser
		}
	}
}

func (client *Client) readUsers(ctx context.Context) ([]*User, error) {
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
