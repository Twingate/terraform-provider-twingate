package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

const userResourceName = "user"

func (client *Client) ReadUsers(ctx context.Context) ([]*model.User, error) {
	variables := newVars(gqlNullable("", query.CursorUsers))
	response := query.ReadUsers{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readUsers"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, nil
	}

	err = response.FetchPages(ctx, client.readUsersAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readUsersAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.UserEdge], error) {
	variables[query.CursorUsers] = cursor
	response := query.ReadUsers{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readUsers"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", userResourceName, "All")
	}

	return &response.PaginatedResource, nil
}

func (client *Client) ReadUser(ctx context.Context, userID string) (*model.User, error) {
	if userID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", userResourceName)
	}

	variables := newVars(gqlID(userID))
	response := query.ReadUser{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readUser"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", userResourceName, userID)
	}

	if response.User == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", userResourceName, userID)
	}

	return response.ToModel(), nil
}
