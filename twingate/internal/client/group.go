package client

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

const groupResourceName = "group"

type PageInfo struct {
	EndCursor   string
	HasNextPage bool
}

func (client *Client) CreateGroup(ctx context.Context, input *model.Group) (*model.Group, error) {
	if input == nil || input.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", groupResourceName)
	}

	variables := newVars(
		gqlVar(input.Name, "name"),
		gqlIDs(input.Users, "userIds"),
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
	)
	response := query.CreateGroup{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("createGroup"))
	if err != nil {
		return nil, NewAPIError(err, "create", groupResourceName)
	}

	if !response.Ok {
		return nil, NewAPIError(NewMutationError(response.Error), "create", groupResourceName)
	}

	group := response.ToModel()
	group.Users = input.Users
	group.IsAuthoritative = input.IsAuthoritative

	return group, nil
}

func (client *Client) ReadGroup(ctx context.Context, groupID string) (*model.Group, error) {
	if groupID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", groupResourceName)
	}

	variables := newVars(gqlID(groupID))
	response := query.ReadGroup{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readGroup"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, groupID)
	}

	if response.Group == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, groupID)
	}

	err = response.Group.Users.FetchPages(ctx, client.readGroupUsersAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) ReadGroups(ctx context.Context, filter *model.GroupsFilter) ([]*model.Group, error) {
	response := query.ReadGroups{}

	variables := newVars(
		gqlNullable(query.NewGroupFilterInput(filter), "filter"),
		gqlNullable("", query.CursorGroups),
	)

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readGroups"))
	if err != nil {
		if filter.HasName() {
			return nil, NewAPIErrorWithName(err, "read", groupResourceName, *filter.Name)
		} else {
			return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
		}
	}

	if len(response.Edges) == 0 {
		if filter.HasName() {
			return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "read", groupResourceName, *filter.Name)
		} else {
			return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
		}
	}

	err = response.FetchPages(ctx, client.readGroupsAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readGroupsAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.GroupEdge], error) {
	variables[query.CursorGroups] = cursor
	response := query.ReadGroups{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readGroups"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
	}

	return &response.PaginatedResource, nil
}

func (client *Client) UpdateGroup(ctx context.Context, input *model.Group) (*model.Group, error) {
	if input == nil || input.ID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, operationUpdate, groupResourceName)
	}

	if input.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, operationUpdate, groupResourceName)
	}

	variables := newVars(
		gqlID(input.ID),
		gqlVar(input.Name, "name"),
		gqlIDs(input.Users, "addedUserIds"),
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
	)

	response := query.UpdateGroup{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("updateGroup"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationUpdate, groupResourceName, input.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), operationUpdate, groupResourceName, input.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationUpdate, groupResourceName, input.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "update", groupResourceName, input.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", groupResourceName, input.ID)
	}

	err = response.Entity.Users.FetchPages(ctx, client.readGroupUsersAfter, newVars(gqlID(input.ID)))
	if err != nil {
		return nil, err //nolint
	}

	group := response.Entity.ToModel()
	group.IsAuthoritative = input.IsAuthoritative

	return group, nil
}

func (client *Client) DeleteGroup(ctx context.Context, groupID string) error {
	if groupID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", groupResourceName)
	}

	variables := newVars(gqlID(groupID))
	response := query.DeleteGroup{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("deleteGroup"))
	if err != nil {
		return NewAPIErrorWithID(err, "delete", groupResourceName, groupID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), "delete", groupResourceName, groupID)
	}

	return nil
}

func (client *Client) DeleteGroupUsers(ctx context.Context, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	if groupID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, operationUpdate, groupResourceName)
	}

	variables := newVars(
		gqlID(groupID),
		gqlIDs(userIDs, "removedUserIds"),
	)

	response := query.UpdateGroupRemoveUsers{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("updateGroup"))
	if err != nil {
		return NewAPIErrorWithID(err, operationUpdate, groupResourceName, groupID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), operationUpdate, groupResourceName, groupID)
	}

	if response.Entity == nil {
		return NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationUpdate, groupResourceName, groupID)
	}

	return nil
}

func (client *Client) readGroupUsersAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.UserEdge], error) {
	response := query.ReadGroup{}
	resourceID := fmt.Sprintf("%v", variables["id"])
	variables[query.CursorUsers] = cursor

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readGroup"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, groupResourceName, resourceID)
	}

	if response.Group == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, groupResourceName, resourceID)
	}

	return &response.Group.Users.PaginatedResource, nil
}
