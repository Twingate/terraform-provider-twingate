package twingate

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
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

type Protocols struct {
	AllowIcmp bool
	UDPPolicy string
	UDPPorts  []string
	TCPPolicy string
	TCPPorts  []string
}

type Resource struct {
	ID               string
	RemoteNetworkID  string
	Address          string
	Name             string
	GroupsIds        []string
	Protocols        *Protocols
	ResourceIDs      []string
	RemoteNetworkIDs []string
}

const resourceResourceName = "resource"

func validatePort(port string) (int64, error) {
	parsed, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("port is not a valid integer: %w", err)
	}

	if parsed < 0 || parsed > 65535 {
		return 0, NewPortNotInRangeError(parsed)
	}

	return parsed, nil
}

func convertPorts(ports []string) (string, error) {
	var converted = make([]string, 0)

	for _, elem := range ports {
		if strings.Contains(elem, "-") {
			split := strings.SplitN(elem, "-", 2)

			start, err := validatePort(split[0])
			if err != nil {
				return "", err
			}

			end, err := validatePort(split[1])
			if err != nil {
				return "", err
			}

			if end < start {
				return "", NewPortRangeNotRisingSequenceError(start, end)
			}

			converted = append(converted, fmt.Sprintf("{start: %s, end: %s}", split[0], split[1]))
		} else {
			_, err := validatePort(elem)
			if err != nil {
				return "", err
			}

			converted = append(converted, fmt.Sprintf("{start: %s, end: %s}", elem, elem))
		}
	}

	if len(converted) > 0 {
		return strings.Join(converted, ","), nil
	}

	return "", nil
}

func convertProtocols(protocols *Protocols) (string, error) {
	if protocols == nil {
		return "", nil
	}

	tcpPorts, err := convertPorts(protocols.TCPPorts)
	if err != nil {
		return "", err
	}

	udpPorts, err := convertPorts(protocols.UDPPorts)
	if err != nil {
		return "", err
	}

	var converted = make([]string, 0)
	converted = append(converted, fmt.Sprintf("tcp: {policy: %s, ports: [%s]}", protocols.TCPPolicy, tcpPorts))
	converted = append(converted, fmt.Sprintf("udp: {policy: %s, ports: [%s]}", protocols.UDPPolicy, udpPorts))
	converted = append(converted, fmt.Sprintf("allowIcmp: %t", protocols.AllowIcmp))
	protocolsQuery := fmt.Sprintf("{%s}", strings.Join(converted, ","))

	return protocolsQuery, nil
}

func convertGroups(groups []string) string {
	var converted = make([]string, 0)
	for _, elem := range groups {
		converted = append(converted, fmt.Sprintf("\"%s\"", elem))
	}

	return fmt.Sprintf("[%s]", strings.Join(converted, ","))
}

func extractPortsFromResults(resourceData *gabs.Container, portPath string) []string {
	var parsedPorts = make([]string, 0)

	if resourceData.ExistsP(portPath) {
		for _, elem := range resourceData.Path(portPath).Children() {
			start := int(elem.Path("start").Data().(float64))
			end := int(elem.Path("end").Data().(float64))

			if start == end {
				parsedPorts = append(parsedPorts, fmt.Sprintf("%d", start))
			} else {
				parsedPorts = append(parsedPorts, fmt.Sprintf("%d-%d", start, end))
			}
		}
	}

	return parsedPorts
}

func extractProtocolsFromResult(resource *Resource, resourceData *gabs.Container) {
	resource.Protocols = &Protocols{
		AllowIcmp: resourceData.Path("protocols.allowIcmp").Data().(bool),
		UDPPolicy: resourceData.Path("protocols.udp.policy").Data().(string),
		TCPPolicy: resourceData.Path("protocols.tcp.policy").Data().(string),
	}
	resource.Protocols.TCPPorts = extractPortsFromResults(resourceData, "protocols.tcp.ports")
	resource.Protocols.UDPPorts = extractPortsFromResults(resourceData, "protocols.udp.ports")
}

func (client *Client) createResource(resource *Resource) error {
	protocols, err := convertProtocols(resource.Protocols)
	if err != nil {
		return NewAPIError(err, "create", resourceResourceName)
	}

	mutation := map[string]string{
		"query": fmt.Sprintf(`
			mutation{
			  resourceCreate(name: "%s", address: "%s", remoteNetworkId: "%s", groupIds: %s, protocols: %s) {
				ok
				error
				entity {
				  id
				}
			  }
		}
        `, resource.Name, resource.Address, resource.RemoteNetworkID, convertGroups(resource.GroupsIds), protocols),
	}

	mutationResource, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return NewAPIError(err, "create", resourceResourceName)
	}

	status := mutationResource.Path("data.resourceCreate.ok").Data().(bool)
	if !status {
		message := mutationResource.Path("data.resourceCreate.error").Data().(string)

		return NewAPIError(NewMutationError(message), "create", resourceResourceName)
	}

	resource.ID = mutationResource.Path("data.resourceCreate.entity.id").Data().(string)

	return nil
}

func (client *Client) readAllResources() (*Resource, error) {
	query := map[string]string{
		"query": "{ resources { edges { node { id } } } }",
	}
	queryResource, err := client.doGraphqlRequest(query)
	if err != nil {
		return nil, fmt.Errorf("error getting resources %s", resourceResourceName)
	}

	var resources = make([]string, 0)

	for _, elem := range queryResource.Path("data.resources.edges").Children() {
		nodeID := elem.Path("node.id").Data().(string)
		resources = append(resources, nodeID)
	}

	resource := &Resource{
		ResourceIDs: resources,
	}

	return resource, nil
}

func (client *Client) readResource(resourceID string) (*Resource, error) { //nolint:funlen
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  resource(id: "%s") {
			id
			name
			address {
			  type
			  value
			}
			remoteNetwork {
			  id
			}
			groups (first: 50){
			  pageInfo{
				hasNextPage
			  }
			  edges{
				node {
				  id
				}
			  }
			}
			protocols {
			  udp {
				ports {
				  end
				  start
				}
				policy
			  }
			  tcp {
				ports {
				  end
				  start
				}
				policy
			  }
			  allowIcmp
			}
		  }
		}
        `, resourceID),
	}

	queryResource, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	resourceQuery := queryResource.Path("data.resource")
	if resourceQuery.Data() == nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	var groups = make([]string, 0)

	hasNextPage := resourceQuery.Path("groups.pageInfo.hasNextPage").Data().(bool)
	if hasNextPage {
		return nil, NewAPIErrorWithID(ErrTooManyGroupsError, "read", resourceResourceName, resourceID)
	}

	for _, elem := range resourceQuery.Path("groups.edges").Children() {
		nodeID := elem.Path("node.id").Data().(string)
		groups = append(groups, nodeID)
	}

	resource := &Resource{
		ID:        resourceID,
		Name:      resourceQuery.Path("name").Data().(string),
		Address:   resourceQuery.Path("address.value").Data().(string),
		GroupsIds: groups,
	}

	if resourceQuery.ExistsP("remoteNetwork.id") {
		resource.RemoteNetworkID = resourceQuery.Path("remoteNetwork.id").Data().(string)
	}

	extractProtocolsFromResult(resource, resourceQuery)

	return resource, nil
}

func (client *Client) updateResource(resource *Resource) error {
	protocols, err := convertProtocols(resource.Protocols)
	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	mutation := map[string]string{
		"query": fmt.Sprintf(`
			mutation{
			  resourceUpdate(id: "%s", name: "%s", address: "%s", remoteNetworkId: "%s", groupIds: %s, protocols: %s) {
				ok
				error
			  }
		}
        `, resource.ID, resource.Name, resource.Address, resource.RemoteNetworkID, convertGroups(resource.GroupsIds), protocols),
	}

	mutationResource, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	status := mutationResource.Path("data.resourceUpdate.ok").Data().(bool)
	if !status {
		message := mutationResource.Path("data.resourceUpdate.error").Data().(string)

		return NewAPIErrorWithID(NewMutationError(message), "update", resourceResourceName, resource.ID)
	}

	return nil
}

func (client *Client) deleteResource(resourceID string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  resourceDelete(id: "%s"){
			ok
			error
		  }
		}
		`, resourceID),
	}
	deleteResource, err := client.doGraphqlRequest(mutation)

	if err != nil {
		return NewAPIErrorWithID(err, "delete", resourceResourceName, resourceID)
	}

	status := deleteResource.Path("data.resourceDelete.ok").Data().(bool)
	if !status {
		message := deleteResource.Path("data.resourceDelete.error").Data().(string)

		return NewAPIErrorWithID(NewMutationError(message), "delete", resourceResourceName, resourceID)
	}

	return nil
}
