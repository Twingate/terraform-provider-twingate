package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

func (client *Client) ReadUsers(ctx context.Context) ([]*model.User, error) {
	op := resourceUser.read()

	variables := newVars(gqlNullable("", query.CursorUsers))

	response := query.ReadUsers{}
	if err := client.query(ctx, &response, variables, op.withCustomName("readUsers"), attr{id: "All"}); err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			return nil, nil
		}

		return nil, err
	}

	if err := response.FetchPages(ctx, client.readUsersAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readUsersAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.UserEdge], error) {
	op := resourceUser.read()

	variables[query.CursorUsers] = cursor
	response := query.ReadUsers{}

	if err := client.query(ctx, &response, variables, op.withCustomName("readUsers"), attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) ReadUser(ctx context.Context, userID string) (*model.User, error) {
	opr := resourceUser.read()

	if userID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(gqlID(userID))
	response := query.ReadUser{}

	if err := client.query(ctx, &response, variables, opr, attr{id: userID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}
