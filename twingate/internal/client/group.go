package client

import (
	"context"
	"fmt"
	"log"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

func (client *Client) CreateGroup(ctx context.Context, input *model.Group) (*model.Group, error) {
	opr := resourceGroup.create()

	if input == nil || input.Name == "" {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	variables := newVars(
		gqlVar(input.Name, "name"),
		gqlIDs(input.Users, "userIds"),
		cursor(query.CursorUsers),
		pageLimit(client.pageLimit),
	)

	response := query.CreateGroup{}
	if err := client.mutate(ctx, &response, variables, opr, attr{name: input.Name}); err != nil {
		return nil, err
	}

	group := response.ToModel()
	group.Users = input.Users
	group.IsAuthoritative = input.IsAuthoritative

	setResource(group)

	return group, nil
}

func (client *Client) ReadGroup(ctx context.Context, groupID string) (*model.Group, error) {
	opr := resourceGroup.read()

	if groupID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if res, ok := getResource[*model.Group](groupID); ok {
		log.Printf("[DEBUG] ReadGroup: found group in cache: %v", res.Name)

		return res, nil
	}

	log.Println("[DEBUG] ReadGroup: group not found in cache: fallback to query API")

	variables := newVars(
		gqlID(groupID),
		cursor(query.CursorUsers),
		pageLimit(client.pageLimit),
	)

	response := query.ReadGroup{}
	if err := client.query(ctx, &response, variables, opr, attr{id: groupID}); err != nil {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.Group.Users.FetchPages(oprCtx, client.readGroupUsersAfter, variables); err != nil {
		return nil, fmt.Errorf("%s: failed to read users for group %s: %w", opr.String(), groupID, err)
	}

	group := response.ToModel()

	setResource(group)

	return group, nil
}

func (client *Client) ReadGroups(ctx context.Context, filter *model.GroupsFilter) ([]*model.Group, error) {
	opr := resourceGroup.read().withCustomName("readGroups")

	// cache is not used when cache filter config set or cache disabled
	if isCacheReady[*model.Group]() {
		if matched := matchResources[*model.Group](filter); len(matched) > 0 {
			log.Printf(
				"[DEBUG] ReadGroups: matched #%d groups from cache: %v",
				len(matched), utils.Map(matched, func(item *model.Group) string {
					return item.Name
				}))

			return matched, nil
		}

		log.Println("[DEBUG] ReadGroups: no matched groups in cache: fallback to query API")
	}

	variables := newVars(
		gqlNullable(query.NewGroupFilterInput(filter), "filter"),
		cursor(query.CursorGroups),
		cursor(query.CursorUsers),
		pageLimit(client.pageLimit),
	)

	response := query.ReadGroups{}
	if err := client.query(ctx, &response, variables, opr,
		attr{id: "All", name: filter.GetName()}); err != nil {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.FetchPages(oprCtx, client.readGroupsAfter, variables); err != nil {
		return nil, err //nolint
	}

	for i, group := range response.Edges {
		if err := response.Edges[i].Node.Users.FetchPages(oprCtx, client.readGroupUsersAfter, newVars(pageLimit(client.pageLimit), gqlID(group.Node.ID))); err != nil {
			return nil, fmt.Errorf("%s: failed to read users for group %s: %w", opr.String(), group.Node.ID, err)
		}
	}

	return response.ToModel(), nil
}

func (client *Client) readGroupsAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.GroupEdge], error) {
	opr := resourceGroup.read().withCustomName("readGroupsAfter")

	variables[query.CursorGroups] = cursor

	response := query.ReadGroups{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) ReadFullGroupsByName(ctx context.Context, filter *model.GroupsFilter) ([]*model.Group, error) {
	opr := resourceGroup.read().withCustomName("readFullGroupsByName")

	variables := newVars(
		gqlNullable(query.NewGroupFilterInput(filter), "filter"),
		cursor(query.CursorGroups),
		cursor(query.CursorUsers),
		pageLimit(extendedPageLimit),
	)

	response := query.ReadGroups{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.FetchPages(oprCtx, client.readGroupsAfter, variables); err != nil {
		return nil, err //nolint
	}

	for i, group := range response.Edges {
		if err := response.Edges[i].Node.Users.FetchPages(oprCtx, client.readGroupUsersAfter, newVars(pageLimit(client.pageLimit), gqlID(group.Node.ID))); err != nil {
			return nil, fmt.Errorf("%s: failed to read users for group %s: %w", opr.String(), group.Node.ID, err)
		}
	}

	return response.ToModel(), nil
}

func (client *Client) ReadFullGroups(ctx context.Context) ([]*model.Group, error) {
	opr := resourceGroup.read().withCustomName("readFullGroups")

	variables := newVars(
		gqlNullable(query.NewGroupFilterInput(nil), "filter"),
		cursor(query.CursorGroups),
		cursor(query.CursorUsers),
		pageLimit(extendedPageLimit),
	)

	response := query.ReadGroups{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.FetchPages(oprCtx, client.readGroupsAfter, variables); err != nil {
		return nil, err //nolint
	}

	for i, group := range response.Edges {
		if err := response.Edges[i].Node.Users.FetchPages(oprCtx, client.readGroupUsersAfter, newVars(pageLimit(client.pageLimit), gqlID(group.Node.ID))); err != nil {
			return nil, fmt.Errorf("%s: failed to read users for group %s: %w", opr.String(), group.Node.ID, err)
		}
	}

	return response.ToModel(), nil
}

func (client *Client) UpdateGroup(ctx context.Context, input *model.Group) (*model.Group, error) {
	opr := resourceGroup.update()

	if input == nil || input.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if input.Name == "" {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	invalidateResource[*model.Group](input.ID)

	variables := newVars(
		gqlID(input.ID),
		gqlVar(input.Name, "name"),
		gqlIDs(input.Users, "addedUserIds"),
		cursor(query.CursorUsers),
		pageLimit(client.pageLimit),
	)

	response := query.UpdateGroup{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.Entity.Users.FetchPages(oprCtx, client.readGroupUsersAfter, newVars(pageLimit(client.pageLimit), gqlID(input.ID))); err != nil {
		return nil, fmt.Errorf("%s: failed to read users for group %s: %w", opr.String(), input.ID, err)
	}

	group := response.Entity.ToModel()
	group.IsAuthoritative = input.IsAuthoritative

	setResource(group)

	return group, nil
}

func (client *Client) DeleteGroup(ctx context.Context, groupID string) error {
	opr := resourceGroup.delete()

	if groupID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	invalidateResource[*model.Group](groupID)

	response := query.DeleteGroup{}

	return client.mutate(ctx, &response, newVars(gqlID(groupID)), opr, attr{id: groupID})
}

func (client *Client) DeleteGroupUsers(ctx context.Context, groupID string, userIDs []string) error {
	opr := resourceGroup.update()

	if len(userIDs) == 0 {
		return nil
	}

	if groupID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	invalidateResource[*model.Group](groupID)

	variables := newVars(
		gqlID(groupID),
		gqlIDs(userIDs, "removedUserIds"),
		cursor(query.CursorUsers),
		pageLimit(client.pageLimit),
	)

	response := query.UpdateGroupRemoveUsers{}

	return client.mutate(ctx, &response, variables, opr, attr{id: groupID})
}

func (client *Client) readGroupUsersAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.UserEdge], error) {
	opr := resourceGroup.read().withCustomName("readGroupUsersAfter")

	variables[query.CursorUsers] = cursor
	resourceID := fmt.Sprintf("%v", variables["id"])

	response := query.ReadGroup{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return &response.Group.Users.PaginatedResource, nil
}
