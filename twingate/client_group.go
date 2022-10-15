package twingate

import (
	"context"
	"errors"

	"github.com/twingate/go-graphql-client"
)

const groupResourceName = "group"
const defaultPageSize = 100

type Group struct {
	ID        graphql.ID
	Name      graphql.String
	Type      graphql.String
	IsActive  graphql.Boolean
	Users     []graphql.ID
	Resources []graphql.ID
}

type createGroupQuery struct {
	GroupCreate struct {
		Entity struct {
			IDName
			Users struct {
				PageInfo struct {
					HasNextPage graphql.Boolean
				}
				Edges []*Edges
			} `graphql:"users(first: $usersPageSize)"`
			Resources struct {
				PageInfo struct {
					HasNextPage graphql.Boolean
				}
				Edges []*Edges
			} `graphql:"resources(first: $resourcesPageSize)"`
		}
		OkError
	} `graphql:"groupCreate(name: $name)"`
}

func (client *Client) createGroup(ctx context.Context, groupName graphql.String, users, resources []string) (*Group, error) {
	if groupName == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", groupResourceName)
	}

	variables := map[string]interface{}{
		"name":              groupName,
		"usersPageSize":     graphql.Int(defaultPageSize),
		"resourcesPageSize": graphql.Int(defaultPageSize),
	}

	if len(users) > 0 {
		variables["userIds"] = convertToGraphqlIDs(users)
	}

	if len(resources) > 0 {
		variables["resourceIds"] = convertToGraphqlIDs(resources)
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
		ID:        response.GroupCreate.Entity.ID,
		Name:      response.GroupCreate.Entity.Name,
		Users:     collectIDs(response.GroupCreate.Entity.Users.Edges),
		Resources: collectIDs(response.GroupCreate.Entity.Resources.Edges),
	}, nil
}

func collectIDs(edges []*Edges) []graphql.ID {
	if len(edges) == 0 {
		return nil
	}

	ids := make([]graphql.ID, 0, len(edges))
	for _, e := range edges {
		ids = append(ids, e.Node.ID)
	}

	return ids
}

func convertToGraphqlIDs(input []string) []graphql.ID {
	res := make([]graphql.ID, 0, len(input))

	for _, elem := range input {
		res = append(res, graphql.ID(elem))
	}

	return res
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
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, groupID)
	}

	group := Group{
		ID:       response.Group.ID,
		Name:     response.Group.Name,
		IsActive: response.Group.IsActive,
		Type:     response.Group.Type,
	}

	return &group, nil
}

type readGroupsQuery struct {
	Groups *struct {
		Edges []*struct {
			Node *struct {
				ID       graphql.ID
				Name     graphql.String
				Type     graphql.String
				IsActive graphql.Boolean
			}
		}
	}
}

func (client *Client) readGroups(ctx context.Context) (groups []*Group, err error) { //nolint
	response := readGroupsQuery{}

	err = client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if response.Groups == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
	}

	for _, g := range response.Groups.Edges {
		groups = append(groups, &Group{
			ID:       g.Node.ID,
			Name:     g.Node.Name,
			Type:     g.Node.Type,
			IsActive: g.Node.IsActive,
		})
	}

	return groups, nil
}

type readGroupsByNameQuery struct {
	Groups struct {
		Edges []*struct {
			Node *struct {
				ID       graphql.ID
				Name     graphql.String
				Type     graphql.String
				IsActive graphql.Boolean
			}
		}
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
		return nil, NewAPIErrorWithName(ErrGraphqlResultIsEmpty, "read", groupResourceName, groupName)
	}

	groups := make([]*Group, 0, len(response.Groups.Edges))

	for _, g := range response.Groups.Edges {
		groups = append(groups, &Group{
			ID:       g.Node.ID,
			Name:     g.Node.Name,
			Type:     g.Node.Type,
			IsActive: g.Node.IsActive,
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

func (client *Client) updateGroup(ctx context.Context, groupID graphql.ID, groupName graphql.String) (*Group, error) {
	if groupID.(string) == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "update", groupResourceName)
	}

	if groupName == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "update", groupResourceName)
	}

	variables := map[string]interface{}{
		"id":   groupID,
		"name": groupName,
	}

	response := updateGroupQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateGroup", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", groupResourceName, groupID)
	}

	if !response.GroupUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.GroupUpdate.Error), "update", groupResourceName, groupID)
	}

	return &Group{
		ID:   response.GroupUpdate.Entity.ID,
		Name: response.GroupUpdate.Entity.Name,
	}, nil
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

type GroupsFilter struct {
	Name     *string
	Type     *string
	IsActive *bool
}

func (f *GroupsFilter) HasName() bool {
	return f != nil && f.Name != nil && *f.Name != ""
}

func (f *GroupsFilter) Match(group *Group) bool {
	if f == nil {
		return true
	}

	if f.Type != nil && *f.Type != string(group.Type) {
		return false
	}

	if f.IsActive != nil && *f.IsActive != bool(group.IsActive) {
		return false
	}

	return true
}

func (client *Client) filterGroups(ctx context.Context, filter *GroupsFilter) ([]*Group, error) {
	var (
		groups []*Group
		err    error
	)

	if !filter.HasName() {
		groups, err = client.readGroups(ctx)
	} else {
		groups, err = client.readGroupsByName(ctx, *filter.Name)
	}

	if err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			return nil, nil
		}

		return nil, err
	}

	var filtered []*Group

	for _, g := range groups {
		if filter.Match(g) {
			filtered = append(filtered, g)
		}
	}

	return filtered, nil
}
