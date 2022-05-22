package twingate

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

const groupResourceName = "group"

type Group struct {
	ID       graphql.ID
	Name     graphql.String
	IsActive graphql.Boolean
}

type createGroupQuery struct {
	GroupCreate struct {
		Entity IDName
		OkError
	} `graphql:"groupCreate(name: $name)"`
}

func (client *Client) createGroup(ctx context.Context, groupName string) (*Group, error) {
	if groupName == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", groupResourceName)
	}

	variables := map[string]interface{}{
		"name": graphql.String(groupName),
	}
	response := createGroupQuery{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("createGroup"))
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
	} `graphql:"group(id: $id)"`
}

func (client *Client) readGroup(ctx context.Context, groupID string) (*Group, error) {
	if groupID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", groupResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(groupID),
	}

	response := readGroupQuery{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readGroup"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, groupID)
	}

	if response.Group == nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, groupID)
	}

	group := Group{
		ID:   response.Group.ID,
		Name: response.Group.Name,
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

type updateGroupQuery struct {
	GroupUpdate struct {
		Entity IDName
		OkError
	} `graphql:"groupUpdate(id: $id, name: $name)"`
}

func (client *Client) updateGroup(ctx context.Context, groupID, groupName string) error {
	if groupID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "update", groupResourceName)
	}

	if groupName == "" {
		return NewAPIError(ErrGraphqlNameIsEmpty, "update", groupResourceName)
	}

	variables := map[string]interface{}{
		"id":   graphql.ID(groupID),
		"name": graphql.String(groupName),
	}

	response := updateGroupQuery{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("updateGroup"))
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

func (client *Client) deleteGroup(ctx context.Context, groupID string) error {
	if groupID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", groupResourceName)
	}

	variables := map[string]interface{}{
		"id": graphql.ID(groupID),
	}

	response := deleteGroupQuery{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("deleteGroup"))
	if err != nil {
		return NewAPIErrorWithID(err, "delete", groupResourceName, groupID)
	}

	if !response.GroupDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.GroupDelete.Error), "delete", groupResourceName, groupID)
	}

	return nil
}
