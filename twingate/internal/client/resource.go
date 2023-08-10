package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hasura/go-graphql-client"
)

type ProtocolsInput struct {
	UDP       *ProtocolInput `json:"udp"`
	TCP       *ProtocolInput `json:"tcp"`
	AllowIcmp bool           `json:"allowIcmp"`
}

type ProtocolInput struct {
	Ports  []*PortRangeInput `json:"ports"`
	Policy string            `json:"policy"`
}

type PortRangeInput struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func newProtocolsInput(protocols *model.Protocols) *ProtocolsInput {
	if protocols == nil {
		return nil
	}

	return &ProtocolsInput{
		UDP:       newProtocol(protocols.UDP),
		TCP:       newProtocol(protocols.TCP),
		AllowIcmp: protocols.AllowIcmp,
	}
}

func newProtocol(protocol *model.Protocol) *ProtocolInput {
	if protocol == nil {
		return nil
	}

	return &ProtocolInput{
		Ports:  newPorts(protocol.Ports),
		Policy: protocol.Policy,
	}
}

func newPorts(ports []*model.PortRange) []*PortRangeInput {
	return utils.Map[*model.PortRange, *PortRangeInput](ports, func(port *model.PortRange) *PortRangeInput {
		return &PortRangeInput{
			Start: port.Start,
			End:   port.End,
		}
	})
}

func (client *Client) CreateResource(ctx context.Context, input *model.Resource) (*model.Resource, error) {
	opr := resourceResource.create()

	variables := newVars(
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlIDs(input.Groups, "groupIds"),
		gqlVar(input.Name, "name"),
		gqlVar(input.Address, "address"),
		gqlVar(newProtocolsInput(input.Protocols), "protocols"),
		gqlNullable(input.IsVisible, "isVisible"),
		gqlNullable(input.IsBrowserShortcutEnabled, "isBrowserShortcutEnabled"),
		gqlNullable(input.Alias, "alias"),
		cursor(query.CursorUsers),
		cursor(query.CursorGroups),
		cursor(query.CursorServices),
		pageLimit(client.pageLimit),
	)

	response := query.CreateResource{}
	if err := client.mutate(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	resource := response.Entity.ToModel()
	resource.Groups = input.Groups
	resource.ServiceAccounts = input.ServiceAccounts
	resource.IsAuthoritative = input.IsAuthoritative

	if input.IsVisible == nil {
		resource.IsVisible = nil
	}

	if input.IsBrowserShortcutEnabled == nil {
		resource.IsBrowserShortcutEnabled = nil
	}

	return resource, nil
}

func (client *Client) ReadResource(ctx context.Context, resourceID string) (*model.Resource, error) {
	opr := resourceResource.read()

	if resourceID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(resourceID),
		cursor(query.CursorUsers),
		cursor(query.CursorGroups),
		cursor(query.CursorServices),
		pageLimit(client.pageLimit),
	)

	response := query.ReadResource{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	if err := response.Resource.Groups.FetchPages(ctx, client.readResourceGroupsAfter, newVars(gqlID(resourceID))); err != nil {
		return nil, err //nolint
	}

	if err := response.Resource.ServiceAccounts.FetchPages(ctx, client.readResourceServiceAccountsAfter, newVars(gqlID(resourceID))); err != nil {
		return nil, err //nolint
	}

	return response.Resource.ToModel(), nil
}

func (client *Client) readResourceGroupsAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.GroupEdge], error) {
	opr := resourceResource.read()

	resourceID := string(variables["id"].(graphql.ID))
	variables[query.CursorGroups] = cursor
	gqlNullable("", query.CursorUsers)(variables)
	pageLimit(client.pageLimit)(variables)

	response := query.ReadResourceGroups{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return &response.Resource.Groups.PaginatedResource, nil
}

func (client *Client) readResourceServiceAccountsAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.ServiceAccountEdge], error) {
	opr := resourceResource.read()

	resourceID := string(variables["id"].(graphql.ID))
	variables[query.CursorServices] = cursor
	pageLimit(client.pageLimit)(variables)

	response := query.ReadResourceServiceAccounts{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return &response.Resource.ServiceAccounts.PaginatedResource, nil
}

func (client *Client) ReadResources(ctx context.Context) ([]*model.Resource, error) {
	opr := resourceResource.read()

	variables := newVars(
		cursor(query.CursorResources),
		pageLimit(client.pageLimit),
	)

	response := query.ReadResources{}
	if err := client.query(ctx, &response, variables, opr.withCustomName("readResources"), attr{id: "All"}); err != nil && !errors.Is(err, ErrGraphqlResultIsEmpty) {
		return nil, err
	}

	if err := response.FetchPages(ctx, client.readResourcesAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readResourcesAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.ResourceEdge], error) {
	opr := resourceResource.read()

	variables[query.CursorResources] = cursor

	response := query.ReadResources{}
	if err := client.query(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) UpdateResource(ctx context.Context, input *model.Resource) (*model.Resource, error) {
	opr := resourceResource.update()

	variables := newVars(
		gqlID(input.ID),
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlVar(input.Name, "name"),
		gqlVar(input.Address, "address"),
		gqlVar(newProtocolsInput(input.Protocols), "protocols"),
		gqlNullable(input.IsVisible, "isVisible"),
		gqlNullable(input.IsBrowserShortcutEnabled, "isBrowserShortcutEnabled"),
		gqlNullable(input.Alias, "alias"),
		cursor(query.CursorUsers),
		cursor(query.CursorGroups),
		cursor(query.CursorServices),
		pageLimit(client.pageLimit),
	)

	response := query.UpdateResource{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
		return nil, err
	}

	if err := response.Entity.Groups.FetchPages(ctx, client.readResourceGroupsAfter, newVars(gqlID(input.ID))); err != nil {
		return nil, err //nolint
	}

	if err := response.Entity.ServiceAccounts.FetchPages(ctx, client.readResourceServiceAccountsAfter, newVars(gqlID(input.ID))); err != nil {
		return nil, err //nolint
	}

	resource := response.Entity.ToModel()
	resource.IsAuthoritative = input.IsAuthoritative

	if input.IsVisible == nil {
		resource.IsVisible = nil
	}

	if input.IsBrowserShortcutEnabled == nil {
		resource.IsBrowserShortcutEnabled = nil
	}

	return resource, nil
}

func (client *Client) DeleteResource(ctx context.Context, resourceID string) error {
	opr := resourceResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}

func (client *Client) UpdateResourceActiveState(ctx context.Context, resource *model.Resource) error {
	opr := resourceResource.update()

	variables := newVars(
		gqlID(resource.ID),
		gqlVar(resource.IsActive, "isActive"),
	)

	response := query.UpdateResourceActiveState{}

	return client.mutate(ctx, &response, variables, opr, attr{id: resource.ID})
}

func (client *Client) ReadResourcesByName(ctx context.Context, name string) ([]*model.Resource, error) {
	opr := resourceResource.read()

	variables := newVars(
		gqlVar(name, "name"),
		cursor(query.CursorResources),
		pageLimit(client.pageLimit),
	)

	response := query.ReadResourcesByName{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	if err := response.FetchPages(ctx, client.readResourcesByNameAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readResourcesByNameAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.ResourceEdge], error) {
	opr := resourceResource.read()

	variables[query.CursorResources] = cursor

	response := query.ReadResourcesByName{}
	if err := client.query(ctx, &response, variables, opr.withCustomName("readResources"), attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) DeleteResourceServiceAccounts(ctx context.Context, resourceID string, deleteServiceAccountIDs []string) error {
	opr := resourceResource.update()

	if len(deleteServiceAccountIDs) == 0 {
		return nil
	}

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	resourcesToDelete := []string{resourceID}

	for _, serviceAccountID := range deleteServiceAccountIDs {
		if err := client.UpdateServiceAccountRemoveResources(ctx, serviceAccountID, resourcesToDelete); err != nil {
			return err
		}
	}

	return nil
}

func (client *Client) AddResourceGroups(ctx context.Context, resource *model.Resource) error {
	opr := resourceResource.update()

	if len(resource.Groups) == 0 {
		return nil
	}

	if resource.ID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(resource.ID),
		gqlIDs(resource.Groups, "groupIds"),
	)

	response := query.AddResourceGroups{}

	return client.mutate(ctx, &response, variables, opr, attr{id: resource.ID})
}

func (client *Client) DeleteResourceGroups(ctx context.Context, resourceID string, deleteGroupIDs []string) error {
	opr := resourceResource.update()

	if len(deleteGroupIDs) == 0 {
		return nil
	}

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(resourceID),
		gqlIDs(deleteGroupIDs, "removedGroupIds"),
		cursor(query.CursorGroups),
		cursor(query.CursorUsers),
		cursor(query.CursorServices),
		pageLimit(client.pageLimit),
	)

	response := query.UpdateResourceRemoveGroups{}

	return client.mutate(ctx, &response, variables, opr, attr{id: resourceID})
}

func (client *Client) ReadResourceServiceAccounts(ctx context.Context, resourceID string) ([]string, error) {
	serviceAccounts, err := client.ReadServiceAccounts(ctx)
	if err != nil {
		return nil, err
	}

	serviceAccountIDs := make([]string, 0, len(serviceAccounts))

	for _, account := range serviceAccounts {
		if utils.Contains(account.Resources, resourceID) {
			serviceAccountIDs = append(serviceAccountIDs, account.ID)
		}
	}

	return serviceAccountIDs, nil
}

func (client *Client) AddResourceServiceAccountIDs(ctx context.Context, resource *model.Resource) error {
	for _, serviceAccountID := range resource.ServiceAccounts {
		_, err := client.UpdateServiceAccount(ctx, &model.ServiceAccount{
			ID:        serviceAccountID,
			Resources: []string{resource.ID},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (client *Client) RemoveResourceAccess(ctx context.Context, resourceID string, principalIDs []string) error {
	opr := resourceResourceAccess.delete()

	if len(principalIDs) == 0 {
		return nil
	}

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(resourceID),
		gqlIDs(principalIDs, "principalIds"),
	)

	response := query.RemoveResourceAccess{}

	return client.mutate(ctx, &response, variables, opr, attr{id: resourceID})
}

type AccessInput struct {
	PrincipalID      string  `json:"principalId"`
	SecurityPolicyID *string `json:"securityPolicyId"`
}

func (client *Client) AddResourceAccess(ctx context.Context, resourceID string, principalIDs []string) error {
	opr := resourceResourceAccess.update()

	if len(principalIDs) == 0 {
		return nil
	}

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	access := utils.Map(principalIDs, func(id string) AccessInput {
		return AccessInput{PrincipalID: id}
	})

	variables := newVars(
		gqlID(resourceID),
		gqlNullable(access, "access"),
	)

	response := query.AddResourceAccess{}

	return client.mutate(ctx, &response, variables, opr, attr{id: resourceID})
}
