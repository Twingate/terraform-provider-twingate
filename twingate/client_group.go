package twingate

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

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

type gqlGroup struct {
	IDName
	IsActive graphql.Boolean
	Type     graphql.String
	Users    struct {
		PageInfo struct {
			HasNextPage graphql.Boolean
		}
		Edges []*IDNode
	} `graphql:"users(first: $usersPageSize)"`
	Resources struct {
		PageInfo struct {
			HasNextPage graphql.Boolean
		}
		Edges []*IDNode
	} `graphql:"resources(first: $resourcesPageSize)"`
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
	} `graphql:"groupCreate(name: $name)"`
}

func (client *Client) createGroup(ctx context.Context, req *Group) (*Group, error) {
	if req.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", groupResourceName)
	}

	variables := map[string]interface{}{
		"name":              req.Name,
		"usersPageSize":     graphql.Int(defaultPageSize),
		"resourcesPageSize": graphql.Int(defaultPageSize),
	}

	addIDsIfNotEmpty(variables, "userIds", req.Users)
	addIDsIfNotEmpty(variables, "resourceIds", req.Resources)

	response := createGroupQuery{}
	newTag := newQuery("groupCreate", variables, []string{"usersPageSize", "resourcesPageSize"})
	patchedResponse := patchGraphqlResponseStruct(&response, "groupCreate", newTag)

	err := client.GraphqlClient.NamedMutate(ctx, "createGroup", patchedResponse, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", groupResourceName)
	}

	if !response.GroupCreate.Ok {
		message := response.GroupCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", groupResourceName)
	}

	return response.GroupCreate.Entity.toModel(), err
}

func addIDsIfNotEmpty(variables map[string]interface{}, key string, ids []graphql.ID) {
	if len(ids) == 0 {
		return
	}

	variables[key] = ids
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
	if len(input) == 0 {
		return nil
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

	return response.Group.toModel(), nil
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
	} `graphql:"groupUpdate(id: $id, name: $name)"`
}

func patchGraphqlResponseStruct(req interface{}, prefix, newTag string) interface{} {
	ptr := reflect.ValueOf(req)
	v := reflect.Indirect(ptr)
	st := v.Type()

	fields := make([]reflect.StructField, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		sf := st.Field(i)
		if strings.HasPrefix(sf.Tag.Get("graphql"), prefix) {
			sf.Tag = reflect.StructTag(fmt.Sprintf(`graphql:"%s"`, newTag))
		}
		fields = append(fields, sf)
	}

	newType := reflect.StructOf(fields)
	newPtrVal := reflect.NewAt(newType, unsafe.Pointer(ptr.Pointer()))
	return newPtrVal.Interface()
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
		"id":                req.ID,
		"name":              req.Name,
		"usersPageSize":     graphql.Int(defaultPageSize),
		"resourcesPageSize": graphql.Int(defaultPageSize),
	}
	addIDsIfNotEmpty(variables, "addedUserIds", addedUsers)
	addIDsIfNotEmpty(variables, "removedUserIds", removedUsers)
	addIDsIfNotEmpty(variables, "addedResourceIds", addedResources)
	addIDsIfNotEmpty(variables, "removedResourceIds", removedResources)

	response := updateGroupQuery{}
	newTag := newQuery("groupUpdate", variables, []string{"usersPageSize", "resourcesPageSize"})
	patchedResponse := patchGraphqlResponseStruct(&response, "groupUpdate", newTag)

	err := client.GraphqlClient.NamedMutate(ctx, "updateGroup", patchedResponse, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", groupResourceName, req.ID)
	}

	if !response.GroupUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.GroupUpdate.Error), "update", groupResourceName, req.ID)
	}

	return response.GroupUpdate.Entity.toModel(), nil
}

func newQuery(queryName string, variables map[string]interface{}, ignoredVariables []string) string {
	vars := make([]string, 0, len(variables))

	ignored := make(map[string]bool, len(ignoredVariables))
	for _, v := range ignoredVariables {
		ignored[v] = true
	}

	for key := range variables {
		if !ignored[key] {
			vars = append(vars, fmt.Sprintf("%s: $%s", key, key))
		}
	}

	return fmt.Sprintf("%s(%s)", queryName, strings.Join(vars, ", "))
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

	var added []graphql.ID
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
