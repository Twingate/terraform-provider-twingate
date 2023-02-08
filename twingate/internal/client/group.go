package client

import (
	"context"

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
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
	)
	response := query.CreateGroup{}

	err := client.GraphqlClient.NamedMutate(ctx, "createGroup", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", groupResourceName)
	}

	if !response.Ok {
		return nil, NewAPIError(NewMutationError(response.Error), "create", groupResourceName)
	}

	return response.ToModel(), nil
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

	return response.ToModel(), nil
}

func (client *Client) ReadGroups(ctx context.Context, filter *model.GroupsFilter) ([]*model.Group, error) {
	response := query.ReadGroups{}
	variables := newVars(
		gqlNullable(query.NewGroupFilterInput(filter), "filter"),
		gqlNullable("", query.CursorGroups),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
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

func (client *Client) UpdateGroup(ctx context.Context, input *model.Group) (*model.Group, error) {
	if input == nil || input.ID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "update", groupResourceName)
	}

	if input.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "update", groupResourceName)
	}

	variables := newVars(
		gqlID(input.ID),
		gqlVar(input.Name, "name"),
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
	)

	response := query.UpdateGroup{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateGroup", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", groupResourceName, input.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "update", groupResourceName, input.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", groupResourceName, input.ID)
	}

	return response.Entity.ToModel(), nil
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
