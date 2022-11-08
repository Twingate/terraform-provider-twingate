package twingate

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/twingate/go-graphql-client"
)

const readResourceQueryGroupsSize = 50

type Resource struct {
	ID              graphql.ID
	RemoteNetworkID graphql.ID
	Address         graphql.String
	Name            graphql.String
	GroupsIds       []graphql.ID
	Protocols       *ProtocolsInput
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
	Protocols *ProtocolsInput
	IsActive  graphql.Boolean
}

type ResourceEdge struct {
	Node *ResourceNode
}

type Resources struct {
	PaginatedResource[*ResourceEdge]
}

func (r *Resources) toList() []*Resource {
	return toList[*ResourceEdge, *Resource](r.Edges,
		func(edge *ResourceEdge) *Resource {
			res := edge.Node

			return &Resource{
				ID:              res.ID,
				Name:            res.Name,
				Address:         res.Address.Value,
				RemoteNetworkID: res.RemoteNetwork.ID,
				Protocols:       res.Protocols,
			}
		},
	)
}

type gqlResource struct {
	ResourceNode
	Groups Groups `graphql:"groups(first: $groupsPageSize)"`
}

type gqlResourceGroups struct {
	ID     graphql.ID
	Groups Groups `graphql:"groups(first: $groupsPageSize, after: $groupsEndCursor)"`
}

func (r *gqlResource) convertResource() *Resource {
	if r == nil {
		return nil
	}

	groups := make([]graphql.ID, 0, len(r.Groups.Edges))
	for _, elem := range r.Groups.Edges {
		groups = append(groups, elem.Node.ID)
	}

	return &Resource{
		ID:              r.ID,
		Name:            r.Name,
		Address:         r.Address.Value,
		RemoteNetworkID: r.RemoteNetwork.ID,
		GroupsIds:       groups,
		Protocols:       r.Protocols,
		IsActive:        r.IsActive,
	}
}

func (r *Resource) stringGroups() []string {
	var groups []string

	if len(r.GroupsIds) > 0 {
		for _, id := range r.GroupsIds {
			groups = append(groups, fmt.Sprintf("%v", id))
		}
	}

	return groups
}

type StringProtocolsInput struct {
	AllowIcmp bool
	UDPPolicy string
	UDPPorts  []string
	TCPPolicy string
	TCPPorts  []string
}

func (spi *StringProtocolsInput) convertToGraphql() (*ProtocolsInput, error) {
	protocols := newEmptyProtocols()
	protocols.AllowIcmp = graphql.Boolean(spi.AllowIcmp)

	protocols.UDP.Policy = graphql.String(spi.UDPPolicy)
	udp, err := convertPorts(spi.UDPPorts)

	if err != nil {
		return nil, err
	}

	protocols.UDP.Ports = udp

	protocols.TCP.Policy = graphql.String(spi.TCPPolicy)
	tcp, err := convertPorts(spi.TCPPorts)

	if err != nil {
		return nil, err
	}

	protocols.TCP.Ports = tcp

	return protocols, nil
}

const resourceResourceName = "resource"

func validatePort(port string) (graphql.Int, error) {
	parsed, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("port is not a valid integer: %w", err)
	}

	if parsed < 0 || parsed > 65535 {
		return 0, NewPortNotInRangeError(parsed)
	}

	return graphql.Int(parsed), nil
}

func convertPortsToSlice(a []interface{}) []string {
	var res = make([]string, 0)

	for _, elem := range a {
		if elem == nil {
			res = append(res, "")

			continue
		}

		res = append(res, elem.(string))
	}

	return res
}

func convertPorts(ports []string) ([]*PortRangeInput, error) {
	converted := []*PortRangeInput{}

	for _, elem := range ports {
		if strings.Contains(elem, "-") {
			split := strings.SplitN(elem, "-", 2) //nolint:gomnd

			start, err := validatePort(split[0])
			if err != nil {
				return converted, ErrInvalidPortRange(elem, err)
			}

			end, err := validatePort(split[1])
			if err != nil {
				return converted, ErrInvalidPortRange(elem, err)
			}

			if end < start {
				return converted, ErrInvalidPortRange(elem, NewPortRangeNotRisingSequenceError(int64(start), int64(end)))
			}

			c := &PortRangeInput{
				Start: start,
				End:   end,
			}

			converted = append(converted, c)
		} else {
			port, err := validatePort(elem)
			if err != nil {
				return converted, ErrInvalidPortRange(elem, err)
			}

			portRange := &PortRangeInput{
				Start: port,
				End:   port,
			}

			converted = append(converted, portRange)
		}
	}

	return converted, nil
}

type createResourceQuery struct {
	ResourceCreate struct {
		OkError
		Entity *gqlResource
	} `graphql:"resourceCreate(name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

func (client *Client) createResource(ctx context.Context, input *Resource) (*Resource, error) {
	variables := map[string]interface{}{
		"name":            input.Name,
		"address":         input.Address,
		"remoteNetworkId": input.RemoteNetworkID,
		"groupIds":        input.GroupsIds,
		"protocols":       input.Protocols,
		"groupsPageSize":  graphql.Int(readResourceQueryGroupsSize),
	}

	response := createResourceQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "createResource", &response, variables)

	if err != nil {
		return nil, NewAPIError(err, "create", resourceResourceName)
	}

	if !response.ResourceCreate.Ok {
		return nil, NewAPIError(NewMutationError(response.ResourceCreate.Error), "create", resourceResourceName)
	}

	if response.ResourceCreate.Entity == nil {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, "create", resourceResourceName)
	}

	resource := response.ResourceCreate.Entity.convertResource()
	resource.GroupsIds = input.GroupsIds

	return resource, nil
}

type readResourceQuery struct {
	Resource *gqlResource `graphql:"resource(id: $id)"`
}

func (client *Client) readResource(ctx context.Context, resourceID string) (*Resource, error) {
	if resourceID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", resourceResourceName)
	}

	response := readResourceQuery{}
	variables := map[string]interface{}{
		"id":             graphql.ID(resourceID),
		"groupsPageSize": graphql.Int(readResourceQueryGroupsSize),
	}

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

	return response.Resource.convertResource(), nil
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

type readResourcesQuery struct { //nolint
	Resources Resources
}

func (client *Client) readResources(ctx context.Context) ([]*Resource, error) { //nolint
	response := readResourcesQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	err = response.Resources.fetchPages(ctx, client.readResourcesAfter, nil)
	if err != nil {
		return nil, err
	}

	return response.Resources.toList(), nil
}

type readResourcesAfterQuery struct { //nolint
	Resources Resources `graphql:"resources(after: $resourcesEndCursor)"`
}

func (client *Client) readResourcesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*ResourceEdge], error) { //nolint
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

func (client *Client) updateResource(ctx context.Context, input *Resource) (*Resource, error) {
	variables := map[string]interface{}{
		"id":              input.ID,
		"name":            input.Name,
		"address":         input.Address,
		"remoteNetworkId": input.RemoteNetworkID,
		"groupIds":        input.GroupsIds,
		"protocols":       input.Protocols,
		"groupsPageSize":  graphql.Int(readResourceQueryGroupsSize),
	}

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

	resource := response.ResourceUpdate.Entity.convertResource()
	resource.GroupsIds = input.GroupsIds

	return resource, nil
}

type deleteResourceQuery struct {
	ResourceDelete *OkError `graphql:"resourceDelete(id: $id)"`
}

func (client *Client) deleteResource(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", resourceResourceName)
	}

	response := deleteResourceQuery{}

	variables := map[string]interface{}{
		"id": graphql.ID(resourceID),
	}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", resourceResourceName, resourceID)
	}

	if !response.ResourceDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ResourceDelete.Error), "delete", resourceResourceName, resourceID)
	}

	return nil
}

type readResourceWithoutGroupsQuery struct {
	Resource *ResourceNode `graphql:"resource(id: $id)"`
}

func (client *Client) readResourceWithoutGroups(ctx context.Context, resourceID string) (*Resource, error) {
	if resourceID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", resourceResourceName)
	}

	response := readResourceWithoutGroupsQuery{}
	variables := map[string]interface{}{
		"id": graphql.ID(resourceID),
	}

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, resourceID)
	}

	resource := &Resource{
		ID:              resourceID,
		Name:            response.Resource.Name,
		Address:         response.Resource.Address.Value,
		RemoteNetworkID: response.Resource.RemoteNetwork.ID,
		Protocols:       response.Resource.Protocols,
	}

	return resource, nil
}

type updateResourceActiveStateQuery struct {
	ResourceUpdate *OkError `graphql:"resourceUpdate(id: $id, isActive: $isActive)"`
}

func (client *Client) updateResourceActiveState(ctx context.Context, resource *Resource) error {
	variables := map[string]interface{}{
		"id":       resource.ID,
		"isActive": resource.IsActive,
	}

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

func (client *Client) readResourcesByName(ctx context.Context, name string) ([]*Resource, error) {
	response := readResourcesByNameQuery{}
	variables := map[string]interface{}{
		"name": graphql.String(name),
	}

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

	return response.Resources.toList(), nil
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
