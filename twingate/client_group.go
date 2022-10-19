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

type GroupUpdateRequest struct {
	ID           graphql.ID
	Name         graphql.String
	Type         graphql.String
	IsActive     graphql.Boolean
	NewUsers     []graphql.ID
	NewResources []graphql.ID
	OldUsers     []graphql.ID
	OldResources []graphql.ID
}

type IDNode struct {
	Node *struct {
		ID graphql.ID
	}
}

type PageInfo struct {
	EndCursor   graphql.String
	HasNextPage graphql.Boolean
}

type Users struct {
	PageInfo PageInfo
	Edges    []*IDNode
}

type Resources struct {
	PageInfo PageInfo
	Edges    []*IDNode
}

type gqlGroup struct {
	IDName
	IsActive  graphql.Boolean
	Type      graphql.String
	Users     Users     `graphql:"users(first: $usersPageSize)"`
	Resources Resources `graphql:"resources(first: $resourcesPageSize)"`
}

type readGroupUsersAndResourcesQuery struct {
	Group *gqlGroupUsersAndResources `graphql:"group(id: $id)"`
}

type gqlGroupUsersAndResources struct {
	ID        graphql.ID
	Users     Users     `graphql:"users(first: $usersPageSize, after: $usersCursor)"`
	Resources Resources `graphql:"resources(first: $resourcesPageSize, after: $resourcesCursor)"`
}

func (g *gqlGroup) toModel() *Group {
	return &Group{
		ID:        g.ID,
		Name:      g.Name,
		IsActive:  g.IsActive,
		Type:      g.Type,
		Users:     collectIDs(g.Users.Edges),
		Resources: collectIDs(g.Resources.Edges),
	}
}

type createGroupQuery struct {
	GroupCreate struct {
		Entity gqlGroup
		OkError
	} `graphql:"groupCreate(name: $name, userIds: $userIds, resourceIds: $resourceIds)"`
}

func (client *Client) createGroup(ctx context.Context, req *Group) (*Group, error) {
	if req.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", groupResourceName)
	}

	variables := map[string]interface{}{
		"name":              req.Name,
		"usersPageSize":     graphql.Int(defaultPageSize),
		"resourcesPageSize": graphql.Int(defaultPageSize),
		"userIds":           req.Users,
		"resourceIds":       req.Resources,
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

	group, err := client.readAllGroupUsersAndResources(ctx, &response.GroupCreate.Entity)
	if err != nil {
		return nil, err
	}

	return group.toModel(), err
}

func (client *Client) readAllGroupUsersAndResources(ctx context.Context, group *gqlGroup) (*gqlGroup, error) {
	usersPage := group.Users.PageInfo
	resourcesPage := group.Resources.PageInfo

	for usersPage.HasNextPage || resourcesPage.HasNextPage {
		resp, err := client.readGroupUsersAndResourcesAfter(ctx, group.ID, usersPage, resourcesPage)
		if err != nil {
			return nil, err
		}

		if len(resp.Users.Edges) > 0 {
			group.Users.Edges = append(group.Users.Edges, resp.Users.Edges...)
		}

		if len(resp.Resources.Edges) > 0 {
			group.Resources.Edges = append(group.Resources.Edges, resp.Resources.Edges...)
		}

		usersPage = resp.Users.PageInfo
		resourcesPage = resp.Resources.PageInfo
	}

	return group, nil
}

func (client *Client) readGroupUsersAndResourcesAfter(ctx context.Context, groupID graphql.ID, usersPage, resourcesPage PageInfo) (*gqlGroupUsersAndResources, error) {
	response := readGroupUsersAndResourcesQuery{}
	variables := map[string]interface{}{
		"id":              groupID,
		"groupsPageSize":  graphql.Int(readResourceQueryGroupsSize),
		"usersCursor":     "",
		"resourcesCursor": "",
	}

	if usersPage.HasNextPage {
		variables["usersCursor"] = usersPage.EndCursor
	}

	if resourcesPage.HasNextPage {
		variables["resourcesCursor"] = resourcesPage.EndCursor
	}

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, groupID)
	}

	if response.Group == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, groupID)
	}

	return response.Group, nil
}

func collectIDs(edges []*IDNode) []graphql.ID {
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
	result := make([]graphql.ID, 0)

	if len(input) == 0 {
		return result
	}

	res := make([]graphql.ID, 0, len(input))

	for _, elem := range input {
		res = append(res, graphql.ID(elem))
	}

	return res
}

type readGroupQuery struct {
	Group *gqlGroup `graphql:"group(id: $id)"`
}

func (client *Client) readGroup(ctx context.Context, groupID graphql.ID) (*Group, error) {
	if groupID.(string) == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", groupResourceName)
	}

	variables := map[string]interface{}{
		"id":                groupID,
		"usersPageSize":     graphql.Int(defaultPageSize),
		"resourcesPageSize": graphql.Int(defaultPageSize),
	}

	response := readGroupQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readGroup", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", groupResourceName, groupID)
	}

	if response.Group == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", groupResourceName, groupID)
	}

	group, err := client.readAllGroupUsersAndResources(ctx, response.Group)
	if err != nil {
		return nil, err
	}

	return group.toModel(), nil
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
		Entity gqlGroup
		OkError
	} `graphql:"groupUpdate(id: $id, name: $name, addedUserIds: $addedUserIds, removedUserIds: $removedUserIds, addedResourceIds: $addedResourceIds, removedResourceIds: $removedResourceIds)"`
}

func (client *Client) updateGroup(ctx context.Context, req *GroupUpdateRequest) (*Group, error) {
	if req.ID.(string) == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "update", groupResourceName)
	}

	if req.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "update", groupResourceName)
	}

	addedUsers, removedUsers := getDelta(req.OldUsers, req.NewUsers)
	addedResources, removedResources := getDelta(req.OldResources, req.NewResources)

	variables := map[string]interface{}{
		"id":                 req.ID,
		"name":               req.Name,
		"usersPageSize":      graphql.Int(defaultPageSize),
		"resourcesPageSize":  graphql.Int(defaultPageSize),
		"addedUserIds":       addedUsers,
		"removedUserIds":     removedUsers,
		"addedResourceIds":   addedResources,
		"removedResourceIds": removedResources,
	}

	response := updateGroupQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "updateGroup", &response, variables)

	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", groupResourceName, req.ID)
	}

	if !response.GroupUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.GroupUpdate.Error), "update", groupResourceName, req.ID)
	}

	group, err := client.readAllGroupUsersAndResources(ctx, &response.GroupUpdate.Entity)
	if err != nil {
		return nil, err
	}

	return group.toModel(), nil
}

func getDelta(oldList, newList []graphql.ID) ([]graphql.ID, []graphql.ID) {
	added := getNewIDs(oldList, newList)
	deleted := getNewIDs(newList, oldList)

	return added, deleted
}

func getNewIDs(oldList, newList []graphql.ID) []graphql.ID {
	current := make(map[graphql.ID]bool)

	for _, item := range oldList {
		current[item] = true
	}

	added := make([]graphql.ID, 0)

	for _, item := range newList {
		if !current[item] {
			added = append(added, item)
		}
	}

	return added
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
