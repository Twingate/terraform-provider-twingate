package twingate

import (
	"context"
	"errors"

	"github.com/twingate/go-graphql-client"
)

const groupResourceName = "group"

type Group struct {
	ID       graphql.ID
	Name     graphql.String
	Type     graphql.String
	IsActive graphql.Boolean
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

type PageInfo struct {
	EndCursor   graphql.String
	HasNextPage graphql.Boolean
}

type GroupNode struct {
	Node *Group
}

type Groups struct {
	PageInfo PageInfo
	Edges    []*GroupNode
}

func (g *Groups) toList() []*Group {
	return toList[*GroupNode, *Group](g.Edges,
		func(edge *GroupNode) *Group {
			return edge.Node
		},
	)
}

func toList[E, O any](edges []E, getObj func(edge E) O) []O {
	out := make([]O, 0, len(edges))
	for _, elem := range edges {
		out = append(out, getObj(elem))
	}

	return out
}

type readGroupsQuery struct {
	Groups Groups
}

func (client *Client) readGroups(ctx context.Context) ([]*Group, error) {
	response := readGroupsQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if len(response.Groups.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
	}

	groups, err := client.readAllGroups(ctx, &response.Groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (client *Client) readAllGroups(ctx context.Context, groups *Groups) ([]*Group, error) {
	page := groups.PageInfo
	for page.HasNextPage {
		resp, err := client.readGroupsAfter(ctx, page.EndCursor)
		if err != nil {
			return nil, err
		}

		groups.Edges = append(groups.Edges, resp.Edges...)
		page = resp.PageInfo
	}

	return groups.toList(), nil
}

type readGroupsAfter struct {
	Groups Groups `graphql:"groups(after: $groupsEndCursor)"`
}

func (client *Client) readGroupsAfter(ctx context.Context, cursor graphql.String) (*Groups, error) {
	response := readGroupsAfter{}
	variables := map[string]interface{}{
		"groupsEndCursor": cursor,
	}

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if len(response.Groups.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
	}

	return &response.Groups, nil
}

type readGroupsByNameQuery struct {
	Groups Groups `graphql:"groups(filter: {name: {eq: $name}})"`
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

	groups, err := client.readAllGroupsByName(ctx, &response.Groups, variables)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (client *Client) readAllGroupsByName(ctx context.Context, groups *Groups, variables map[string]interface{}) ([]*Group, error) {
	page := groups.PageInfo
	for page.HasNextPage {
		resp, err := client.readGroupsByNameAfter(ctx, page.EndCursor, variables)
		if err != nil {
			return nil, err
		}

		groups.Edges = append(groups.Edges, resp.Edges...)
		page = resp.PageInfo
	}

	return groups.toList(), nil
}

type readGroupsByNameAfter struct {
	Groups Groups `graphql:"groups(filter: {name: {eq: $name}}, after: $groupsEndCursor)"`
}

func (client *Client) readGroupsByNameAfter(ctx context.Context, cursor graphql.String, variables map[string]interface{}) (*Groups, error) {
	response := readGroupsByNameAfter{}
	variables["groupsEndCursor"] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readGroups", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, "All")
	}

	if len(response.Groups.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, "All")
	}

	return &response.Groups, nil
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
