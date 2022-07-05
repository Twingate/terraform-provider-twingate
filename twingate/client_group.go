package twingate

import (
	"context"

	"github.com/twingate/go-graphql-client"
)

const groupResourceName = "group"

type Group struct {
	ID       graphql.ID
	Name     graphql.String
	IsActive graphql.Boolean
	Type     graphql.String
}

type createGroupQuery struct {
	GroupCreate struct {
		Entity IDName
		OkError
	} `graphql:"groupCreate(name: $name)"`
}

func (client *Client) createGroup(ctx context.Context, groupName graphql.String) (*Group, error) {
	if groupName == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", groupResourceName)
	}

	variables := map[string]interface{}{
		"name": groupName,
	}
	response := createGroupQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "createGroup", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", groupResourceName)
	}

	if !response.GroupCreate.Ok {
		message := response.GroupCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", groupResourceName)
	}

	return &Group{
		ID:   response.GroupCreate.Entity.ID,
		Name: response.GroupCreate.Entity.Name,
	}, nil
}

type readGroupQuery struct {
	Group *struct {
		IDName
		IsActive graphql.Boolean
		Type     graphql.String
	} `graphql:"group(id: $id)"`
}

func (client *Client) readGroup(ctx context.Context, groupID graphql.ID) (*Group, error) {
	if groupID.(string) == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", groupResourceName)
	}

	variables := map[string]interface{}{
		"id": groupID,
	}

	response := readGroupQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readGroup", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, groupID)
	}

	if response.Group == nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, groupID)
	}

	group := Group{
		ID:       response.Group.ID,
		Name:     response.Group.Name,
		IsActive: response.Group.IsActive,
		Type:     response.Group.Type,
	}

	return &group, nil
}

type readGroupsQuery struct { //nolint
	Groups *struct {
		Edges []*Edges
	}
}

func (client *Client) readGroups(ctx context.Context) (groups []*Group, err error) { //nolint
	response := readGroupsQuery{}

	err = client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if response.Groups == nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	for _, g := range response.Groups.Edges {
		groups = append(groups, &Group{
			ID:   g.Node.ID,
			Name: g.Node.Name,
		})
	}

	return groups, nil
}

type readGroupsByNameQuery struct {
	Groups struct {
		Edges []*Edges
	} `graphql:"groups(filter: {name: {eq: $name}})"`
}

func (client *Client) readGroupsByName(ctx context.Context, groupName string) ([]*Group, error) {
	if groupName == "" {
		return nil, NewAPIError(ErrGraphqlGroupNameIsEmpty, "read", groupResourceName)
	}

	response := readGroupsByNameQuery{}

	variables := map[string]interface{}{
		"name": graphql.String(groupName),
	}

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, "read", groupResourceName, groupName)
	}

	if len(response.Groups.Edges) == 0 {
		return nil, NewAPIErrorWithName(ErrGraphqlResourceNotFound, "read", groupResourceName, groupName)
	}

	groups := make([]*Group, 0, len(response.Groups.Edges))

	for _, g := range response.Groups.Edges {
		groups = append(groups, &Group{
			ID:   g.Node.ID,
			Name: g.Node.Name,
		})
	}

	return groups, nil
}

type updateGroupQuery struct {
	GroupUpdate struct {
		Entity IDName
		OkError
	} `graphql:"groupUpdate(id: $id, name: $name)"`
}

func (client *Client) updateGroup(ctx context.Context, groupID graphql.ID, groupName graphql.String) error {
	if groupID.(string) == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "update", groupResourceName)
	}

	if groupName == "" {
		return NewAPIError(ErrGraphqlNameIsEmpty, "update", groupResourceName)
	}

	variables := map[string]interface{}{
		"id":   groupID,
		"name": groupName,
	}

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

func (client *Client) deleteGroup(ctx context.Context, groupID graphql.ID) error {
	if groupID.(string) == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", groupResourceName)
	}

	variables := map[string]interface{}{
		"id": groupID,
	}

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
