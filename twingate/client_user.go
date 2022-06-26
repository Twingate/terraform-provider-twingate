package twingate

import (
	"context"

	"github.com/twingate/go-graphql-client"
)

const userResourceName = "user"

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	IsAdmin   bool
}

type gqlUser struct {
	ID        graphql.ID
	FirstName graphql.String
	LastName  graphql.String
	Email     graphql.String
	IsAdmin   graphql.Boolean
}

type readUsersQuery struct {
	Users *struct {
		Edges []*struct {
			Node *gqlUser
		}
	}
}

func (client *Client) readUsers(ctx context.Context) (users map[string]*User, err error) {
	response := readUsersQuery{}

	err = client.GraphqlClient.NamedQuery(ctx, "readUsers", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, "All")
	}

	if response.Users == nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, "All")
	}

	users = make(map[string]*User)

	for _, item := range response.Users.Edges {
		if item == nil {
			continue
		}

		user := item.Node
		if user == nil {
			continue
		}

		users[string(user.Email)] = &User{
			ID:        user.ID.(string),
			FirstName: string(user.FirstName),
			LastName:  string(user.LastName),
			Email:     string(user.Email),
			IsAdmin:   bool(user.IsAdmin),
		}
	}

	return users, nil
}
