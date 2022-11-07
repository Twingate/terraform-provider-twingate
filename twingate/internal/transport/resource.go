package transport

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

type gqlResource struct {
	IDName
	Groups  Groups `graphql:"groups(first: $groupsPageSize)"`
	Address struct {
		Value graphql.String
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	Protocols *Protocols
	IsActive  graphql.Boolean
}

type PageInfo struct {
	EndCursor   graphql.String
	HasNextPage graphql.Boolean
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
	// TODO: convert model to gql resource
	//protocols := &ProtocolsInput{
	//	AllowIcmp: true,
	//	TCP: &ProtocolInput{
	//		Policy: model.PolicyAllowAll,
	//		Ports:  []*PortRangeInput{
	//			//		{Start: 80, End: 83},
	//			//		{Start: 85, End: 85},
	//		},
	//	},
	//	UDP: &ProtocolInput{
	//		Policy: model.PolicyAllowAll,
	//		Ports:  []*PortRangeInput{},
	//	},
	//}

	variables := newVars(
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlIDs(input.Groups, "groupIds"),
		gqlField(input.Name, "name"),
		gqlField(input.Address, "address"),
		gqlField(readResourceQueryGroupsSize, "groupsPageSize"),
		//gqlField(newProtocolsInput(resource.Protocols), "protocols"),
	)
	//variables["protocols"] = protocols
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
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, resourceID)
	}

	resource, err := client.readAllResourceGroups(ctx, response.Resource)
	if err != nil {
		return nil, err
	}

	return resource.ToModel(), nil
}

func (client *Client) readAllResourceGroups(ctx context.Context, resource *gqlResource) (*gqlResource, error) {
	page := resource.Groups.PageInfo
	for page.HasNextPage {
		resp, err := client.readResourceGroupsAfter(ctx, resource.ID, page.EndCursor)
		if err != nil {
			return nil, err
		}

		resource.Groups.Edges = append(resource.Groups.Edges, resp.Resource.Groups.Edges...)
		page = resp.Resource.Groups.PageInfo
	}

	return resource, nil
}

type readResourceGroupsQuery struct {
	Resource *gqlResourceGroups `graphql:"resource(id: $id)"`
}

func (client *Client) readResourceGroupsAfter(ctx context.Context, resourceID graphql.ID, cursor graphql.String) (*readResourceGroupsQuery, error) {
	response := readResourceGroupsQuery{}
	variables := newVars(
		gqlID(resourceID),
		gqlField(readResourceQueryGroupsSize, "groupsPageSize"),
		gqlField(cursor, "groupsEndCursor"),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, resourceID)
	}

	return &response, nil
}

type readResourcesQuery struct { //nolint
	Resources Resources
}

func (client *Client) ReadResources(ctx context.Context) ([]*model.Resource, error) { //nolint
	response := readResourcesQuery{}
	variables := newVars(
		gqlField(readResourceQueryGroupsSize, "groupsPageSize"),
	)
	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	resource, err := client.readAllResources(ctx, &response.Resources)
	if err != nil {
		return nil, err
	}

	return resource.ToModel(), nil
}

func (client *Client) readAllResources(ctx context.Context, resource *Resources) (*Resources, error) { //nolint
	page := resource.PageInfo

	for page.HasNextPage {
		resp, err := client.readResourcesAfter(ctx, page.EndCursor)
		if err != nil {
			return nil, err
		}

		resource.Edges = append(resource.Edges, resp.Resources.Edges...)
		page = resp.Resources.PageInfo
	}

	return resource, nil
}

type readResourcesAfterQuery struct { //nolint
	Resources Resources `graphql:"resources(after: $resourcesEndCursor)"`
}

func (client *Client) readResourcesAfter(ctx context.Context, cursor graphql.String) (*readResourcesAfterQuery, error) { //nolint
	response := readResourcesAfterQuery{}
	variables := newVars(
		gqlField(readResourceQueryGroupsSize, "groupsPageSize"),
		gqlField(cursor, "resourcesEndCursor"),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "read", resourceResourceName)
	}

	if len(response.Resources.Edges) == 0 {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "read", resourceResourceName)
	}

	return &response, nil
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

type ResourceEdge struct {
	Node *gqlResource
}

type Resources struct {
	PageInfo PageInfo
	Edges    []*ResourceEdge
}

type readResourcesByNameQuery struct {
	Resources Resources `graphql:"resources(filter: {name: {eq: $name}})"`
}

func (client *Client) ReadResourcesByName(ctx context.Context, name string) ([]*model.Resource, error) {
	response := readResourcesByNameQuery{}
	variables := newVars(
		gqlField(name, "name"),
		gqlField(readResourceQueryGroupsSize, "groupsPageSize"),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	if len(response.Resources.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, "All")
	}

	resources, err := client.readAllResourcesByName(ctx, &response.Resources, variables)
	if err != nil {
		return nil, err
	}

	return resources.ToModel(), nil
}

func (client *Client) readAllResourcesByName(ctx context.Context, resources *Resources, variables map[string]interface{}) (*Resources, error) {
	page := resources.PageInfo
	for page.HasNextPage {
		resp, err := client.readResourcesByNameAfter(ctx, page.EndCursor, variables)
		if err != nil {
			return nil, err
		}

		resources.Edges = append(resources.Edges, resp.Edges...)
		page = resp.PageInfo
	}

	return resources, nil
}

type readResourcesByNameAfter struct {
	Resources Resources `graphql:"resources(filter: {name: {eq: $name}}, after: $resourcesEndCursor)"`
}

func (client *Client) readResourcesByNameAfter(ctx context.Context, cursor graphql.String, variables map[string]interface{}) (*Resources, error) {
	response := readResourcesByNameAfter{}
	variables["resourcesEndCursor"] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	if len(response.Resources.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, "All")
	}

	return &response.Resources, nil
}
