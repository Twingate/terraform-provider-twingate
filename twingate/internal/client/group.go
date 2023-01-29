package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const groupResourceName = "group"

type PageInfo struct {
	EndCursor   graphql.String
	HasNextPage graphql.Boolean
}

func (client *Client) CreateGroup(ctx context.Context, input *model.Group) (*model.Group, error) {
	if input == nil || input.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", groupResourceName)
	}

	variables := newVars(
		gqlVar(input.Name, "name"),
		gqlIDs(input.Users, "userIds"),
	)
	response := query.CreateGroup{}

	err := client.GraphqlClient.NamedMutate(ctx, "createGroup", &response, variables)
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

	err := client.GraphqlClient.NamedQuery(ctx, "readGroup", &response, variables)
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

func (client *Client) ReadGroups(ctx context.Context) ([]*model.Group, error) {
	response := query.ReadGroups{}
	variables := newVars(gqlNullable("", query.CursorGroups))

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
	}

	err = response.FetchPages(ctx, client.readGroupsAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readGroupsAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.GroupEdge], error) {
	variables[query.CursorGroups] = cursor
	response := query.ReadGroups{}

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
	}

	return &response.PaginatedResource, nil
}

func (client *Client) ReadGroupsByName(ctx context.Context, groupName string) ([]*model.Group, error) {
	if groupName == "" {
		return nil, NewAPIError(ErrGraphqlGroupNameIsEmpty, "read", groupResourceName)
	}

	response := query.ReadGroupsByName{}
	variables := newVars(
		gqlVar(groupName, "name"),
		gqlNullable("", query.CursorGroups),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, "read", groupResourceName, groupName)
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "read", groupResourceName, groupName)
	}

	err = response.FetchPages(ctx, client.readGroupsByNameAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readGroupsByNameAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.GroupEdge], error) {
	response := query.ReadGroupsByName{}
	variables[query.CursorGroups] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
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
	)

	response := query.UpdateGroup{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateGroup", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationUpdate, groupResourceName, input.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), operationUpdate, groupResourceName, input.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationUpdate, groupResourceName, input.ID)
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

	err := client.GraphqlClient.NamedMutate(ctx, "deleteGroup", &response, variables)
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

	err := client.GraphqlClient.NamedMutate(ctx, "updateGroup", &response, variables)
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

type GroupsFilter struct {
	Name     *string
	Type     *string
	IsActive *bool
}

func (f *GroupsFilter) HasName() bool {
	return f != nil && f.Name != nil && *f.Name != ""
}

func (f *GroupsFilter) Match(group *model.Group) bool {
	if f == nil {
		return true
	}

	if f.Type != nil && *f.Type != group.Type {
		return false
	}

	if f.IsActive != nil && *f.IsActive != group.IsActive {
		return false
	}

	return true
}

func (client *Client) FilterGroups(ctx context.Context, filter *GroupsFilter) ([]*model.Group, error) {
	var (
		groups []*model.Group
		err    error
	)

	if !filter.HasName() {
		groups, err = client.ReadGroups(ctx)
	} else {
		groups, err = client.ReadGroupsByName(ctx, *filter.Name)
	}

	if err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			return nil, nil
		}

		return nil, err
	}

	var filtered []*model.Group

	for _, g := range groups {
		if filter.Match(g) {
			filtered = append(filtered, g)
		}
	}

	return filtered, nil
}

func (client *Client) readGroupUsersAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.UserEdge], error) {
	response := query.ReadGroup{}
	resourceID := variables["id"]
	variables[query.CursorUsers] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readGroup", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, groupResourceName, resourceID)
	}

	if response.Group == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, groupResourceName, resourceID)
	}

	return &response.Group.Users.PaginatedResource, nil
}
