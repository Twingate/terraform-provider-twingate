package transport

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const groupResourceName = "group"

type gqlGroup struct {
	IDName
	IsActive graphql.Boolean
	Type     graphql.String
}

type gqlGroups struct {
	Edges []*struct {
		Node *gqlGroup
	}
}

type createGroupQuery struct {
	GroupCreate struct {
		Entity IDName
		OkError
	} `graphql:"groupCreate(name: $name)"`
}

func (client *Client) CreateGroup(ctx context.Context, groupName string) (*model.Group, error) {
	if groupName == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", groupResourceName)
	}

	variables := newVars(gqlField(groupName, "name"))
	response := createGroupQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "createGroup", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", groupResourceName)
	}

	if !response.GroupCreate.Ok {
		message := response.GroupCreate.Error
		return nil, NewAPIError(NewMutationError(message), "create", groupResourceName)
	}

	return response.ToModel(), nil
}

type readGroupQuery struct {
	Group *gqlGroup `graphql:"group(id: $id)"`
}

func (client *Client) ReadGroup(ctx context.Context, groupID string) (*model.Group, error) {
	if groupID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", groupResourceName)
	}

	variables := newVars(gqlID(groupID))
	response := readGroupQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readGroup", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, groupID)
	}

	if response.Group == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, groupID)
	}

	return response.ToModel(), nil
}

type readGroupsQuery struct {
	Groups gqlGroups
}

func (client *Client) ReadGroups(ctx context.Context) ([]*model.Group, error) {
	response := readGroupsQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if len(response.Groups.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
	}

	return response.ToModel(), nil
}

type readGroupsByNameQuery struct {
	Groups gqlGroups `graphql:"groups(filter: {name: {eq: $name}})"`
}

func (client *Client) ReadGroupsByName(ctx context.Context, groupName string) ([]*model.Group, error) {
	if groupName == "" {
		return nil, NewAPIError(ErrGraphqlGroupNameIsEmpty, "read", groupResourceName)
	}

	response := readGroupsByNameQuery{}
	variables := newVars(gqlField(groupName, "name"))

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, "read", groupResourceName, groupName)
	}

	if len(response.Groups.Edges) == 0 {
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "read", groupResourceName, groupName)
	}

	return response.ToModel(), nil
}

type updateGroupQuery struct {
	GroupUpdate struct {
		Entity IDName
		OkError
	} `graphql:"groupUpdate(id: $id, name: $name)"`
}

func (client *Client) UpdateGroup(ctx context.Context, groupID, groupName string) error {
	if groupID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "update", groupResourceName)
	}

	if groupName == "" {
		return NewAPIError(ErrGraphqlNameIsEmpty, "update", groupResourceName)
	}

	variables := newVars(
		gqlID(groupID),
		gqlField(groupName, "name"),
	)

	response := updateGroupQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateGroup", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "update", groupResourceName, groupID)
	}

	if !response.GroupUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(response.GroupUpdate.Error), "update", groupResourceName, groupID)
	}

	return nil
}

type deleteGroupQuery struct {
	GroupDelete *OkError `graphql:"groupDelete(id: $id)" json:"groupDelete"`
}

func (client *Client) DeleteGroup(ctx context.Context, groupID string) error {
	if groupID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", groupResourceName)
	}

	variables := newVars(gqlID(groupID))
	response := deleteGroupQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "deleteGroup", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", groupResourceName, groupID)
	}

	if !response.GroupDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.GroupDelete.Error), "delete", groupResourceName, groupID)
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
