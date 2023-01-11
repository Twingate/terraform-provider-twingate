package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/twingate/go-graphql-client"
)

const (
	resourceResourceName        = "resource"
	readResourceQueryGroupsSize = 50
)

type ProtocolsInput struct {
	UDP       *ProtocolInput  `json:"udp"`
	TCP       *ProtocolInput  `json:"tcp"`
	AllowIcmp graphql.Boolean `json:"allowIcmp"`
}

type ProtocolInput struct {
	Ports  []*PortRangeInput `json:"ports"`
	Policy graphql.String    `json:"policy"`
}

type PortRangeInput struct {
	Start graphql.Int `json:"start"`
	End   graphql.Int `json:"end"`
}

func newProtocolsInput(protocols *model.Protocols) *ProtocolsInput {
	if protocols == nil {
		return nil
	}

	return &ProtocolsInput{
		UDP:       newProtocol(protocols.UDP),
		TCP:       newProtocol(protocols.TCP),
		AllowIcmp: graphql.Boolean(protocols.AllowIcmp),
	}
}

func newProtocol(protocol *model.Protocol) *ProtocolInput {
	if protocol == nil {
		return nil
	}

	return &ProtocolInput{
		Ports:  newPorts(protocol.Ports),
		Policy: graphql.String(protocol.Policy),
	}
}

func newPorts(ports []*model.PortRange) []*PortRangeInput {
	return utils.Map[*model.PortRange, *PortRangeInput](ports, func(port *model.PortRange) *PortRangeInput {
		return &PortRangeInput{
			Start: graphql.Int(port.Start),
			End:   graphql.Int(port.End),
		}
	})
}

func (client *Client) CreateResource(ctx context.Context, input *model.Resource) (*model.Resource, error) {
	variables := newVars(
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlIDs(input.Groups, "groupIds"),
		gqlVar(input.Name, "name"),
		gqlVar(input.Address, "address"),
	)
	variables["protocols"] = newProtocolsInput(input.Protocols)

	response := query.CreateResource{}

	err := client.GraphqlClient.NamedMutate(ctx, "createResource", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", resourceResourceName)
	}

	if !response.Ok {
		return nil, NewAPIError(NewMutationError(response.Error), "create", resourceResourceName)
	}

	if response.Entity == nil {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "create", resourceResourceName)
	}

	resource := response.Entity.ToModel()
	resource.Groups = input.Groups
	resource.ServiceAccounts = input.ServiceAccounts
	resource.IsAuthoritative = input.IsAuthoritative

	return resource, nil
}

func (client *Client) ReadResource(ctx context.Context, resourceID string) (*model.Resource, error) {
	if resourceID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", resourceResourceName)
	}

	response := query.ReadResource{}
	variables := newVars(gqlID(resourceID))

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, resourceID)
	}

	err = response.Resource.Groups.FetchPages(ctx, client.readResourceGroupsAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.Resource.ToModel(), nil
}

func (client *Client) readResourceGroupsAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.GroupEdge], error) {
	response := query.ReadResourceGroups{}
	resourceID := variables["id"]
	variables[query.CursorGroups] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, resourceID)
	}

	return &response.Resource.Groups.PaginatedResource, nil
}

func (client *Client) ReadResources(ctx context.Context) ([]*model.Resource, error) {
	response := query.ReadResources{}
	variables := newVars(gqlNullable("", query.CursorResources))

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	err = response.FetchPages(ctx, client.readResourcesAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readResourcesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.ResourceEdge], error) {
	variables[query.CursorResources] = cursor
	response := query.ReadResources{}

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "read", resourceResourceName)
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "read", resourceResourceName)
	}

	return &response.PaginatedResource, nil
}

func (client *Client) UpdateResource(ctx context.Context, input *model.Resource) (*model.Resource, error) {
	variables := newVars(
		gqlID(input.ID),
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlIDs(input.Groups, "groupIds"),
		gqlVar(input.Name, "name"),
		gqlVar(input.Address, "address"),
		gqlVar(newProtocolsInput(input.Protocols), "protocols"),
	)

	response := query.UpdateResource{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)

	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", resourceResourceName, input.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "update", resourceResourceName, input.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", resourceResourceName, input.ID)
	}

	err = response.Entity.Groups.FetchPages(ctx, client.readResourceGroupsAfter, newVars(gqlID(input.ID)))
	if err != nil {
		return nil, err //nolint
	}

	resource := response.Entity.ToModel()
	resource.ServiceAccounts = input.ServiceAccounts
	resource.IsAuthoritative = input.IsAuthoritative

	return resource, nil
}

func (client *Client) DeleteResource(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", resourceResourceName)
	}

	response := query.DeleteResource{}

	variables := newVars(gqlID(resourceID))

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", resourceResourceName, resourceID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), "delete", resourceResourceName, resourceID)
	}

	return nil
}

func (client *Client) UpdateResourceActiveState(ctx context.Context, resource *model.Resource) error {
	variables := newVars(
		gqlID(resource.ID),
		gqlVar(resource.IsActive, "isActive"),
	)

	response := query.UpdateResourceActiveState{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)

	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), "update", resourceResourceName, resource.ID)
	}

	return nil
}

func (client *Client) ReadResourcesByName(ctx context.Context, name string) ([]*model.Resource, error) {
	response := query.ReadResourcesByName{}
	variables := newVars(
		gqlVar(name, "name"),
		gqlNullable("", query.CursorResources),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, "All")
	}

	err = response.FetchPages(ctx, client.readResourcesByNameAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readResourcesByNameAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.ResourceEdge], error) {
	response := query.ReadResourcesByName{}
	variables[query.CursorResources] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, "All")
	}

	return &response.PaginatedResource, nil
}

func (client *Client) DeleteResourceServiceAccounts(ctx context.Context, resourceID string, deleteServiceAccountIDs []string) error {
	if len(deleteServiceAccountIDs) == 0 {
		return nil
	}

	if resourceID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, operationUpdate, resourceResourceName)
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
	if len(resource.Groups) == 0 {
		return nil
	}

	if resource.ID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, operationUpdate, resourceResourceName)
	}

	variables := newVars(
		gqlID(resource.ID),
		gqlIDs(resource.Groups, "groupIds"),
	)

	response := query.AddResourceGroups{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, operationUpdate, resourceResourceName, resource.ID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), operationUpdate, resourceResourceName, resource.ID)
	}

	return nil
}

func (client *Client) DeleteResourceGroups(ctx context.Context, resourceID string, deleteGroupIDs []string) error {
	if len(deleteGroupIDs) == 0 {
		return nil
	}

	if resourceID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, operationUpdate, resourceResourceName)
	}

	response := query.UpdateResourceRemoveGroups{}
	variables := newVars(
		gqlID(resourceID),
		gqlIDs(deleteGroupIDs, "removedGroupIds"),
	)

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, operationUpdate, resourceResourceName, resourceID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), operationUpdate, resourceResourceName, resourceID)
	}

	if response.Entity == nil {
		return NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationUpdate, resourceResourceName, resourceID)
	}

	return nil
}

func (client *Client) ReadResourceServiceAccounts(ctx context.Context, resourceID string) ([]string, error) {
	serviceAccounts, err := client.ReadServiceAccounts(ctx)
	if err != nil {
		return nil, err
	}

	serviceAccountIDs := make(map[string]bool)

	for _, account := range serviceAccounts {
		if utils.Contains(account.Resources, resourceID) {
			serviceAccountIDs[account.ID] = true
		}
	}

	return utils.MapKeys(serviceAccountIDs), nil
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
