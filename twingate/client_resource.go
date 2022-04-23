package twingate

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/twingate/go-graphql-client"
)

const readResourceQueryGroupsSize = 50

var (
	ErrTooManyGroupsError        = errors.New("provider does not support more than 50 groups per resource")
	ErrGraphqlIDIsEmpty          = errors.New("id is empty")
	ErrGraphqlConnectorIDIsEmpty = errors.New("network id is empty")
	ErrGraphqlNetworkIDIsEmpty   = errors.New("network id is empty")
	ErrGraphqlNetworkNameIsEmpty = errors.New("network name is empty")
)

func ErrInvalidPortRange(portRange string, err error) error {
	return fmt.Errorf(`failed to parse protocols port range "%s": %w`, portRange, err)
}

type PortNotInRangeError struct {
	Port int64
}

func NewPortNotInRangeError(port int64) *PortNotInRangeError {
	return &PortNotInRangeError{
		Port: port,
	}
}

func (e *PortNotInRangeError) Error() string {
	return fmt.Sprintf("port %d not in the range of 0-65535", e.Port)
}

type PortRangeNotRisingSequenceError struct {
	Start int64
	End   int64
}

func NewPortRangeNotRisingSequenceError(start int64, end int64) *PortRangeNotRisingSequenceError {
	return &PortRangeNotRisingSequenceError{
		Start: start,
		End:   end,
	}
}

func (e *PortRangeNotRisingSequenceError) Error() string {
	return fmt.Sprintf("ports %d, %d needs to be in a rising sequence", e.Start, e.End)
}

type Resource struct {
	ID              graphql.ID
	RemoteNetworkID graphql.ID
	Address         graphql.String
	Name            graphql.String
	GroupsIds       []*graphql.ID
	Protocols       *ProtocolsInput
}

func (r *Resource) stringGroups() []string {
	var groups []string

	if len(r.GroupsIds) > 0 {
		for _, id := range r.GroupsIds {
			groups = append(groups, fmt.Sprintf("%v", *id))
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
	parsed, err := strconv.ParseInt(port, 10, 64) //nolint:gomnd
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
		Entity struct {
			ID graphql.ID
		}
	} `graphql:"resourceCreate(name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

func (client *Client) createResource(ctx context.Context, resource *Resource) error {
	variables := map[string]interface{}{
		"name":            resource.Name,
		"address":         resource.Address,
		"remoteNetworkId": resource.RemoteNetworkID,
		"groupIds":        resource.GroupsIds,
		"protocols":       resource.Protocols,
	}

	response := createResourceQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "createResource", &response, variables)

	if err != nil {
		return NewAPIError(err, "create", resourceResourceName)
	}

	if !response.ResourceCreate.Ok {
		return NewAPIError(NewMutationError(response.ResourceCreate.Error), "create", resourceResourceName)
	}

	resource.ID = response.ResourceCreate.Entity.ID

	return nil
}

type readResourceQuery struct {
	Resource *struct {
		IDName
		Address struct {
			Type  graphql.String
			Value graphql.String
		}
		RemoteNetwork struct {
			ID graphql.ID
		}
		Groups struct {
			PageInfo struct {
				HasNextPage graphql.Boolean
			}
			Edges []*Edges
		} `graphql:"groups(first: $first)"`
		Protocols *ProtocolsInput
	} `graphql:"resource(id: $id)"`
}

func (client *Client) readResource(ctx context.Context, resourceID graphql.ID) (*Resource, error) {
	if resourceID.(string) == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", resourceResourceName)
	}

	response := readResourceQuery{}
	variables := map[string]interface{}{
		"id":    resourceID,
		"first": graphql.Int(readResourceQueryGroupsSize),
	}

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	var groups = make([]*graphql.ID, 0)

	if response.Resource.Groups.PageInfo.HasNextPage {
		return nil, NewAPIErrorWithID(ErrTooManyGroupsError, "read", resourceResourceName, resourceID)
	}

	for _, elem := range response.Resource.Groups.Edges {
		groups = append(groups, &elem.Node.ID)
	}

	resource := &Resource{
		ID:              resourceID,
		Name:            response.Resource.Name,
		Address:         response.Resource.Address.Value,
		RemoteNetworkID: response.Resource.RemoteNetwork.ID,
		GroupsIds:       groups,
		Protocols:       response.Resource.Protocols,
	}

	return resource, nil
}

type readResourcesQuery struct { //nolint
	Resources struct {
		Edges []*Edges
	}
}

func (client *Client) readResources(ctx context.Context) ([]*Edges, error) { //nolint
	response := readResourcesQuery{}
	variables := map[string]interface{}{}

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	return response.Resources.Edges, nil
}

type updateResourceQuery struct {
	ResourceUpdate *OkError `graphql:"resourceUpdate(id: $id, name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

func (client *Client) updateResource(ctx context.Context, resource *Resource) error {
	variables := map[string]interface{}{
		"id":              resource.ID,
		"name":            resource.Name,
		"address":         resource.Address,
		"remoteNetworkId": resource.RemoteNetworkID,
		"groupIds":        resource.GroupsIds,
		"protocols":       resource.Protocols,
	}

	response := updateResourceQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)

	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	if !response.ResourceUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ResourceUpdate.Error), "update", resourceResourceName, resource.ID)
	}

	return nil
}

type deleteResourceQuery struct {
	ResourceDelete *OkError `graphql:"resourceDelete(id: $id)"`
}

func (client *Client) deleteResource(ctx context.Context, resourceID graphql.ID) error {
	if resourceID.(string) == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", resourceResourceName)
	}

	response := deleteResourceQuery{}

	variables := map[string]interface{}{
		"id": resourceID,
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
