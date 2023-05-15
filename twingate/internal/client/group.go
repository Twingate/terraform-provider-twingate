package client

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

func (client *Client) CreateGroup(ctx context.Context, input *model.Group) (*model.Group, error) {
	opr := resourceGroup.create()

	if input == nil || input.Name == "" {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	variables := newVars(
		gqlVar(input.Name, "name"),
		gqlIDs(input.Users, "userIds"),
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
		gqlNullable("", query.CursorUsers),
	)

	response := query.CreateGroup{}
	if err := client.mutate(ctx, &response, variables, opr, attr{name: input.Name}); err != nil {
		return nil, err
	}

	group := response.ToModel()
	group.Users = input.Users
	group.IsAuthoritative = input.IsAuthoritative

	return group, nil
}

func (client *Client) ReadGroup(ctx context.Context, groupID string) (*model.Group, error) {
	opr := resourceGroup.read()

	if groupID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(groupID),
		gqlNullable("", query.CursorUsers),
	)

	response := query.ReadGroup{}
	if err := client.query(ctx, &response, variables, opr, attr{id: groupID}); err != nil {
		return nil, err
	}

	if err := response.Group.Users.FetchPages(ctx, client.readGroupUsersAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) ReadGroups(ctx context.Context, filter *model.GroupsFilter) ([]*model.Group, error) {
	opr := resourceGroup.read()

	variables := newVars(
		gqlNullable(query.NewGroupFilterInput(filter), "filter"),
		gqlNullable("", query.CursorGroups),
		gqlNullable("", query.CursorUsers),
	)

	response := query.ReadGroups{}
	if err := client.query(ctx, &response, variables, opr.withCustomName("readGroups"),
		attr{id: "All", name: filter.GetName()}); err != nil {
		return nil, err
	}

	if err := response.FetchPages(ctx, client.readGroupsAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readGroupsAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.GroupEdge], error) {
	op := resourceGroup.read()

	variables[query.CursorGroups] = cursor

	response := query.ReadGroups{}
	if err := client.query(ctx, &response, variables, op.withCustomName("readGroups"), attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) UpdateGroup(ctx context.Context, input *model.Group) (*model.Group, error) {
	opr := resourceGroup.update()

	if input == nil || input.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if input.Name == "" {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	variables := newVars(
		gqlID(input.ID),
		gqlVar(input.Name, "name"),
		gqlIDs(input.Users, "addedUserIds"),
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
		gqlNullable("", query.CursorUsers),
	)

	response := query.UpdateGroup{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
		return nil, err
	}

	if err := response.Entity.Users.FetchPages(ctx, client.readGroupUsersAfter, newVars(gqlID(input.ID))); err != nil {
		return nil, err //nolint
	}

	group := response.Entity.ToModel()
	group.IsAuthoritative = input.IsAuthoritative

	return group, nil
}

func (client *Client) DeleteGroup(ctx context.Context, groupID string) error {
	opr := resourceGroup.delete()

	if groupID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteGroup{}

	return client.mutate(ctx, &response, newVars(gqlID(groupID)), opr, attr{id: groupID})
}

func (client *Client) DeleteGroupUsers(ctx context.Context, groupID string, userIDs []string) error {
	opr := resourceGroup.update()

	if len(userIDs) == 0 {
		return nil
	}

	if groupID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(groupID),
		gqlIDs(userIDs, "removedUserIds"),
		gqlNullable("", query.CursorUsers),
	)

	response := query.UpdateGroupRemoveUsers{}

	return client.mutate(ctx, &response, variables, opr, attr{id: groupID})
}

func (client *Client) readGroupUsersAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.UserEdge], error) {
	opr := resourceGroup.read()

	variables[query.CursorUsers] = cursor
	resourceID := fmt.Sprintf("%v", variables["id"])

	response := query.ReadGroup{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return &response.Group.Users.PaginatedResource, nil
}
