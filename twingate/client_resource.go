package twingate

import (
	"fmt"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

type Protocols struct {
	AllowIcmp bool
	UDPPolicy string
	UDPPorts  []string
	TCPPolicy string
	TCPPorts  []string
}

type Resource struct {
	Id              string
	RemoteNetworkId string
	Address         string
	Name            string
	Groups          []string
	Protocols       *Protocols
}

func convertPorts(ports []string) string {
	var converted []string
	for _, elem := range ports {
		if strings.Contains(elem, "-") {
			split := strings.SplitN(elem, "-", 2)
			converted = append(converted, fmt.Sprintf("{start: %s, end: %s}", split[0], split[1]))
		} else {
			converted = append(converted, fmt.Sprintf("{start: %s, end: %s}", elem, elem))
		}
	}
	if len(converted) > 0 {
		return strings.Join(converted, ",")
	}

	return ""
}

func convertProtocols(protocols *Protocols) string {
	var converted []string
	if protocols == nil {
		return ""
	}

	converted = append(converted, fmt.Sprintf("tcp: {policy: %s, ports: [%s]}", protocols.TCPPolicy, convertPorts(protocols.TCPPorts)))
	converted = append(converted, fmt.Sprintf("udp: {policy: %s, ports: [%s]}", protocols.UDPPolicy, convertPorts(protocols.UDPPorts)))
	converted = append(converted, fmt.Sprintf("allowIcmp: %t", protocols.AllowIcmp))
	protocolsQuery := fmt.Sprintf("{%s}", strings.Join(converted, ","))

	return protocolsQuery
}
func convertGroups(groups []string) string {
	var converted []string
	for _, elem := range groups {
		converted = append(converted, fmt.Sprintf("\"%s\"", elem))
	}

	return fmt.Sprintf("[%s]", strings.Join(converted, ","))
}
func (client *Client) createResource(resource *Resource) error {
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
        `, resource.Name, resource.Address, resource.RemoteNetworkId, convertGroups(resource.Groups), convertProtocols(resource.Protocols)),
	}
	mutationResource, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return fmt.Errorf("can't create resource : %w", err)
	}

	status := mutationResource.Path("data.resourceCreate.ok").Data().(bool)
	if !status {
		errorMessage := mutationResource.Path("data.resourceCreate.error").Data().(string)

		return APIError("can't create resource name %s, error: %s", resource.Name, errorMessage)
	}
	resource.Id = mutationResource.Path("data.resourceCreate.entity.id").Data().(string)

	return nil
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

func (client *Client) readResource(resourceId string) (*Resource, error) { //nolint:funlen
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
			groups {
			  edges {
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
        `, resourceId),
	}
	queryResource, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return nil, fmt.Errorf("can't read resource : %w", err)
	}

	resourceQuery := queryResource.Path("data.resource")

	if resourceQuery.Data() == nil {
		return nil, APIError("can't read resource: %s", resourceId)
	}
	var groups []string
	for _, elem := range resourceQuery.Path("groups.edges").Children() {
		nodeId := elem.Path("node.id").Data().(string)
		groups = append(groups, nodeId)
	}

	resource := &Resource{
		Id:      resourceId,
		Name:    resourceQuery.Path("name").Data().(string),
		Address: resourceQuery.Path("address.value").Data().(string),
		Groups:  groups,
	}

	if resourceQuery.ExistsP("remoteNetwork.id") {
		resource.RemoteNetworkId = resourceQuery.Path("remoteNetwork.id").Data().(string)
	}

	extractProtocolsFromResult(resource, resourceQuery)

	return resource, nil
}

func (client *Client) updateResource(resource *Resource) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
			mutation{
			  resourceUpdate(id: "%s", name: "%s", address: "%s", remoteNetworkId: "%s", groupIds: %s, protocols: %s) {
				ok
				error
			  }
		}
        `, resource.Id, resource.Name, resource.Address, resource.RemoteNetworkId, convertGroups(resource.Groups), convertProtocols(resource.Protocols)),
	}
	mutationResource, err := client.doGraphqlRequest(mutation)
	if err != nil {
		return fmt.Errorf("can't update resource : %w", err)
	}

	status := mutationResource.Path("data.resourceUpdate.ok").Data().(bool)
	if !status {
		errorMessage := mutationResource.Path("data.resourceUpdate.error").Data().(string)

		return APIError("can't update resource: %s", errorMessage)
	}

	return nil
}

func (client *Client) deleteResource(resourceId string) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  resourceDelete(id: "%s"){
			ok
			error
		  }
		}
		`, resourceId),
	}
	deleteResource, err := client.doGraphqlRequest(mutation)

	if err != nil {
		return fmt.Errorf("can't delete resource : %w", err)
	}

	status := deleteResource.Path("data.resourceDelete.ok").Data().(bool)
	if !status {
		errorMessage := deleteResource.Path("data.resourceDelete.error").Data().(string)

		return APIError("unable to delete resource Id %s, error: %s", resourceId, errorMessage)
	}

	return nil
}
