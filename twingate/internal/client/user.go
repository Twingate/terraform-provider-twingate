package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

func (client *Client) ReadUsers(ctx context.Context) ([]*model.User, error) {
	opr := resourceUser.read()

	variables := newVars(
		gqlNullable("", query.CursorUsers),
		gqlVar(client.pageLimit, query.PageLimitUsers),
	)

	response := query.ReadUsers{}
	if err := client.query(ctx, &response, variables, opr.withCustomName("readUsers"), attr{id: "All"}); err != nil {
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
	opr := resourceUser.read()

	variables[query.CursorUsers] = cursor
	response := query.ReadUsers{}

	if err := client.query(ctx, &response, variables, opr.withCustomName("readUsers"), attr{id: "All"}); err != nil {
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

func (client *Client) CreateUser(ctx context.Context, input *model.User) (*model.User, error) {
	opr := resourceUser.create()

	if input == nil || input.Email == "" {
		return nil, opr.apiError(ErrGraphqlEmailIsEmpty)
	}

	variables := newVars(
		gqlVar(input.Email, "email"),
		gqlNullable(input.FirstName, "firstName"),
		gqlNullable(input.LastName, "lastName"),
		gqlVar(query.UserRole(input.Role), "role"),
		gqlVar(input.SendInvite, "shouldSendInvite"),
	)
	response := query.CreateUser{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: input.Email}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) UpdateUser(ctx context.Context, input *model.UserUpdate) (*model.User, error) {
	opr := resourceUser.update()

	if input == nil || input.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if input.FirstName != nil || input.LastName != nil || input.IsActive != nil {
		variables := newVars(
			gqlID(input.ID),
			gqlNullable(input.FirstName, "firstName"),
			gqlNullable(input.LastName, "lastName"),
			gqlNullable(query.NewUserStateUpdateInput(input.State()), "state"),
		)

		response := query.UpdateUser{}

		if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
			return nil, err
		}

		user := response.ToModel()

		if input.Role == nil {
			return user, nil
		}
	}

	return client.UpdateUserRole(ctx, input)
}

func (client *Client) UpdateUserRole(ctx context.Context, input *model.UserUpdate) (*model.User, error) {
	opr := resourceUser.update()

	if input == nil || input.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if input.Role == nil {
		return client.ReadUser(ctx, input.ID)
	}

	variables := newVars(
		gqlID(input.ID),
		gqlVar(query.UserRole(*input.Role), "role"),
	)
	response := query.UpdateUserRole{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteUser(ctx context.Context, userID string) error {
	opr := resourceUser.delete()

	if userID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteUser{}

	return client.mutate(ctx, &response, newVars(gqlID(userID)), opr, attr{id: userID})
}
