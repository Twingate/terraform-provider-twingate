package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const (
	resourceResourceName        = "resource"
	readResourceQueryGroupsSize = 50
)

type Resource struct {
	ID              graphql.ID
	RemoteNetworkID graphql.ID
	Address         graphql.String
	Name            graphql.String
	GroupsIds       []*graphql.ID
	Protocols       *Protocols
	IsActive        graphql.Boolean
}

type ResourceNode struct {
	IDName
	Address struct {
		Value graphql.String
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	Protocols *Protocols
	IsActive  graphql.Boolean
}

type ResourceEdge struct {
	Node *ResourceNode
}

type Resources struct {
	PaginatedResource[*ResourceEdge]
}

type gqlResource struct {
	ResourceNode
	Groups Groups `graphql:"groups(first: $groupsPageSize)"`
}

type gqlResourceGroups struct {
	ID     graphql.ID
	Groups Groups `graphql:"groups(first: $groupsPageSize, after: $groupsEndCursor)"`
}

type Protocols struct {
	UDP       *Protocol       `json:"udp"`
	TCP       *Protocol       `json:"tcp"`
	AllowIcmp graphql.Boolean `json:"allowIcmp"`
}

type Protocol struct {
	Ports  []*PortRange   `json:"ports"`
	Policy graphql.String `json:"policy"`
}

type PortRange struct {
	Start graphql.Int `json:"start"`
	End   graphql.Int `json:"end"`
}

type createResourceQuery struct {
	ResourceCreate struct {
		OkError
		Entity *gqlResource
	} `graphql:"resourceCreate(name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

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

func (client *Client) CreateResource(ctx context.Context, input *model.Resource) (*model.Resource, error) {
	variables := newVars(
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlIDs(input.Groups, "groupIds"),
		gqlField(input.Name, "name"),
		gqlField(input.Address, "address"),
		gqlField(readResourceQueryGroupsSize, "groupsPageSize"),
		gqlNullableField("", cursorUsers),
	)
	variables["protocols"] = newProtocolsInput(input.Protocols)

	response := createResourceQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "createResource", &response, variables)

	if err != nil {
		return nil, NewAPIError(err, "create", resourceResourceName)
	}

	if !response.ResourceCreate.Ok {
		return nil, NewAPIError(NewMutationError(response.ResourceCreate.Error), "create", resourceResourceName)
	}

	resource := response.ResourceCreate.Entity.ToModel()
	resource.Groups = input.Groups

	return resource, nil
}

type readResourceQuery struct {
	Resource *gqlResource `graphql:"resource(id: $id)"`
}

func (client *Client) ReadResource(ctx context.Context, resourceID string) (*model.Resource, error) {
	if resourceID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", resourceResourceName)
	}

	response := readResourceQuery{}
	variables := newVars(
		gqlID(resourceID),
		gqlField(readResourceQueryGroupsSize, "groupsPageSize"),
		gqlNullableField("", cursorUsers),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, resourceID)
	}

	err = response.Resource.Groups.fetchPages(ctx, client.readResourceGroupsAfter, variables)
	if err != nil {
		return nil, err
	}

	return response.Resource.ToModel(), nil
}

type readResourceGroupsQuery struct {
	Resource *gqlResourceGroups `graphql:"resource(id: $id)"`
}

func (client *Client) readResourceGroupsAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*GroupEdge], error) {
	response := readResourceGroupsQuery{}
	resourceID := variables["id"]
	variables["groupsEndCursor"] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, resourceID)
	}

	return &response.Resource.Groups.PaginatedResource, nil
}

type readResourcesQuery struct {
	Resources Resources
}

func (client *Client) ReadResources(ctx context.Context) ([]*model.Resource, error) {
	response := readResourcesQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	err = response.Resources.fetchPages(ctx, client.readResourcesAfter, nil)
	if err != nil {
		return nil, err
	}

	return response.Resources.ToModel(), nil
}

type readResourcesAfterQuery struct {
	Resources Resources `graphql:"resources(after: $resourcesEndCursor)"`
}

func (client *Client) readResourcesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*ResourceEdge], error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	variables["resourcesEndCursor"] = cursor
	response := readResourcesAfterQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "read", resourceResourceName)
	}

	if len(response.Resources.Edges) == 0 {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "read", resourceResourceName)
	}

	return &response.Resources.PaginatedResource, nil
}

type updateResourceQuery struct {
	ResourceUpdate *struct {
		OkError
		Entity *gqlResource
	} `graphql:"resourceUpdate(id: $id, name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

func (client *Client) UpdateResource(ctx context.Context, input *model.Resource) (*model.Resource, error) {
	variables := newVars(
		gqlID(input.ID),
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlIDs(input.Groups, "groupIds"),
		gqlField(input.Name, "name"),
		gqlField(input.Address, "address"),
		gqlField(newProtocolsInput(input.Protocols), "protocols"),
		gqlField(readResourceQueryGroupsSize, "groupsPageSize"),
		gqlNullableField("", cursorUsers),
	)

	response := updateResourceQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)

	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", resourceResourceName, input.ID)
	}

	if !response.ResourceUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.ResourceUpdate.Error), "update", resourceResourceName, input.ID)
	}

	if response.ResourceUpdate.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", resourceResourceName, input.ID)
	}

	resource := response.ResourceUpdate.Entity.ToModel()
	resource.Groups = input.Groups

	return resource, nil
}

type deleteResourceQuery struct {
	ResourceDelete *OkError `graphql:"resourceDelete(id: $id)"`
}

func (client *Client) DeleteResource(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", resourceResourceName)
	}

	response := deleteResourceQuery{}

	variables := newVars(gqlID(resourceID))

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", resourceResourceName, resourceID)
	}

	if !response.ResourceDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ResourceDelete.Error), "delete", resourceResourceName, resourceID)
	}

	return nil
}

type updateResourceActiveStateQuery struct {
	ResourceUpdate *OkError `graphql:"resourceUpdate(id: $id, isActive: $isActive)"`
}

func (client *Client) UpdateResourceActiveState(ctx context.Context, resource *model.Resource) error {
	variables := newVars(
		gqlID(resource.ID),
		gqlField(resource.IsActive, "isActive"),
	)

	response := updateResourceActiveStateQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)

	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	if !response.ResourceUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ResourceUpdate.Error), "update", resourceResourceName, resource.ID)
	}

	return nil
}

type readResourcesByNameQuery struct {
	Resources Resources `graphql:"resources(filter: {name: {eq: $name}})"`
}

func (client *Client) ReadResourcesByName(ctx context.Context, name string) ([]*model.Resource, error) {
	response := readResourcesByNameQuery{}
	variables := newVars(
		gqlField(name, "name"),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	if len(response.Resources.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, "All")
	}

	err = response.Resources.fetchPages(ctx, client.readResourcesByNameAfter, variables)
	if err != nil {
		return nil, err
	}

	return response.Resources.ToModel(), nil
}

type readResourcesByNameAfter struct {
	Resources Resources `graphql:"resources(filter: {name: {eq: $name}}, after: $resourcesEndCursor)"`
}

func (client *Client) readResourcesByNameAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*ResourceEdge], error) {
	response := readResourcesByNameAfter{}
	variables["resourcesEndCursor"] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	if len(response.Resources.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, "All")
	}

	return &response.Resources.PaginatedResource, nil
}
