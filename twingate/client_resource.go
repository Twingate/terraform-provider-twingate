package twingate

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hasura/go-graphql-client"
)

var ErrTooManyGroupsError = errors.New("provider does not support more than 50 groups per resource")

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

func (spi *StringProtocolsInput) convertToGraphql() *ProtocolsInput {
	pi := newEmptyProtocols()
	pi.AllowIcmp = graphql.Boolean(spi.AllowIcmp)

	pi.UDP.Policy = graphql.String(spi.UDPPolicy)
	udp, err := convertPorts(spi.UDPPorts)
	if err != nil {
		return nil

	}
	pi.UDP.Ports = udp

	pi.TCP.Policy = graphql.String(spi.TCPPolicy)
	tcp, err := convertPorts(spi.TCPPorts)
	if err != nil {
		return nil

	}
	pi.TCP.Ports = tcp

	return pi
}

type Resources struct {
	ID   string
	Name string
}

const resourceResourceName = "resource"

func validatePortGraphql(port string) (graphql.Int, error) {
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
		res = append(res, elem.(string))
	}
	return res
}

func convertPorts(ports []string) ([]*PortRangeInput, error) {
	converted := []*PortRangeInput{}

	for _, elem := range ports {
		if strings.Contains(elem, "-") {
			split := strings.SplitN(elem, "-", 2)

			start, err := validatePortGraphql(split[0])
			if err != nil {
				return converted, err
			}

			end, err := validatePortGraphql(split[1])
			if err != nil {
				return converted, err
			}

			if end < start {
				return converted, NewPortRangeNotRisingSequenceError(int64(start), int64(end))
			}
			c := &PortRangeInput{
				Start: start,
				End:   end,
			}
			converted = append(converted, c)

		} else {
			p, err := validatePortGraphql(elem)
			if err != nil {
				return converted, err
			}

			c := &PortRangeInput{
				Start: p,
				End:   p,
			}

			converted = append(converted, c)

		}
	}

	if len(converted) > 0 {
		return converted, nil
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

func (client *Client) createResource(resource *Resource) error {
	variables := map[string]interface{}{
		"name":            graphql.String(resource.Name),
		"address":         graphql.String(resource.Address),
		"remoteNetworkId": graphql.ID(resource.RemoteNetworkID),
		"groupIds":        resource.GroupsIds,
		"protocols":       resource.Protocols,
	}

	r := createResourceQuery{}

	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)
	if err != nil {
		return NewAPIError(err, "create", resourceResourceName)
	}

	if !r.ResourceCreate.Ok {
		return NewAPIError(NewMutationError(r.ResourceCreate.Error), "create", resourceResourceName)
	}

	resource.ID = r.ResourceCreate.Entity.ID.(string)

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

func (client *Client) readResource(resourceID string) (*Resource, error) { //nolint:funlen
	r := readResourceQuery{}
	variables := map[string]interface{}{
		"id":    graphql.ID(resourceID),
		"first": graphql.Int(50),
	}

	err := client.GraphqlClient.Query(context.Background(), &r, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, resourceID)
	}

	if r.Resource == nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	var groups = make([]*graphql.ID, 0)

	if r.Resource.Groups.PageInfo.HasNextPage {
		return nil, NewAPIErrorWithID(ErrTooManyGroupsError, "read", resourceResourceName, resourceID)
	}

	for _, elem := range r.Resource.Groups.Edges {
		groups = append(groups, &elem.Node.ID)
	}

	resource := &Resource{
		ID:              resourceID,
		Name:            r.Resource.Name,
		Address:         r.Resource.Address.Value,
		RemoteNetworkID: r.Resource.RemoteNetwork.ID,
		GroupsIds:       groups,
		Protocols:       r.Resource.Protocols,
	}

	return resource, nil
}

type readResourcesQuery struct { //nolint
	Resources struct {
		Edges []*Edges
	}
}

func (client *Client) readResources() (map[int]*Resources, error) { //nolint
	r := readResourcesQuery{}
	variables := map[string]interface{}{}

	err := client.GraphqlClient.Query(context.Background(), &r, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	var resources = make(map[int]*Resources)

	for i, elem := range r.Resources.Edges {
		c := &Resources{ID: elem.Node.StringID(), Name: elem.Node.StringName()}
		resources[i] = c
	}

	return resources, nil
}

type updateResourceQuery struct {
	ResourceUpdate *OkError `graphql:"resourceUpdate(id: $id, name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

func (client *Client) updateResource(resource *Resource) error {

	variables := map[string]interface{}{
		"id":              graphql.ID(resource.ID),
		"name":            graphql.String(resource.Name),
		"address":         graphql.String(resource.Address),
		"remoteNetworkId": graphql.ID(resource.RemoteNetworkID),
		"groupIds":        resource.GroupsIds,
		"protocols":       resource.Protocols,
	}

	r := updateResourceQuery{}

	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)

	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	if !r.ResourceUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(r.ResourceUpdate.Error), "update", resourceResourceName, resource.ID)
	}

	return nil
}

type deleteResourceQuery struct {
	ResourceDelete *OkError `graphql:"resourceDelete(id: $id)"`
}

func (client *Client) deleteResource(resourceID string) error {
	r := deleteResourceQuery{}

	variables := map[string]interface{}{
		"id": graphql.ID(resourceID),
	}

	err := client.GraphqlClient.Mutate(context.Background(), &r, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", resourceResourceName, resourceID)
	}

	if !r.ResourceDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(r.ResourceDelete.Error), "delete", resourceResourceName, resourceID)
	}

	return nil
}
