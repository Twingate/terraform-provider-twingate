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
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
		cursor(query.CursorAccess),
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

	if input.SecurityPolicyID == nil {
		resource.SecurityPolicyID = nil
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
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)

	response := query.ReadResource{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	if err := response.Resource.Access.FetchPages(ctx, client.readResourceAccessAfter, newVars(gqlID(resourceID))); err != nil {
		return nil, err //nolint
	}

	return response.Resource.ToModel(), nil
}

func (client *Client) readResourceAccessAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
	opr := resourceResource.read()

	resourceID := string(variables["id"].(graphql.ID))
	variables[query.CursorAccess] = cursor
	pageLimit(client.pageLimit)(variables)

	response := query.ReadResourceAccess{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return &response.Resource.Access.PaginatedResource, nil
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
		gqlVar(input.IsActive, "isActive"),
		gqlNullable(input.IsVisible, "isVisible"),
		gqlNullable(input.IsBrowserShortcutEnabled, "isBrowserShortcutEnabled"),
		gqlNullable(input.Alias, "alias"),
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)

	response := query.UpdateResource{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
		return nil, err
	}

	if err := response.Entity.Access.FetchPages(ctx, client.readResourceAccessAfter, newVars(gqlID(input.ID))); err != nil {
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

	if input.SecurityPolicyID == nil {
		resource.SecurityPolicyID = nil
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
