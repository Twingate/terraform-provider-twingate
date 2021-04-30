package twingate

import (
	"fmt"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

type Protocols struct {
	AllowIcmp bool
	UdpPolicy string
	UdpPorts  []string
	TcpPolicy string
	TcpPorts  []string
}

type Resource struct {
	Id              string
	RemoteNetworkId string
	Address         string
	Name            string
	Groups          []string
	Protocols       *Protocols
}

func convertPortsToGraphql(ports []string) string {
	var converted = make([]string, 0)
	for _, elem := range ports {
		if strings.Contains(elem, "-") {
			split := strings.SplitN(elem, "-", 2)
			converted = append(converted, fmt.Sprintf("{start: %s, end: %s}", split[0], split[1]))
		} else {
			converted = append(converted, fmt.Sprintf("{start: %s, end: %s}", elem, elem))
		}
	}
	if len(converted) > 0 {
		return fmt.Sprintf("%s", strings.Join(converted, ","))
	}
	return ""
}
func protocolsToGraphql(protocols *Protocols) string {
	var converted = make([]string, 0)
	converted = append(converted, fmt.Sprintf("allowIcmp: %t", protocols.AllowIcmp))
	if protocols.TcpPolicy != "" {
		converted = append(converted, fmt.Sprintf("tcp: {policy: %s, ports: [%s]}", protocols.TcpPolicy, convertPortsToGraphql(protocols.TcpPorts)))
	}
	if protocols.UdpPolicy != "" {
		converted = append(converted, fmt.Sprintf("udp: {policy: %s, ports: [%s]}", protocols.UdpPolicy, convertPortsToGraphql(protocols.UdpPorts)))
	}
	protocolsQuery := fmt.Sprintf("{%s}", strings.Join(converted, ","))
	return protocolsQuery
}
func convertGroups(groups []string) string {
	var converted = make([]string, 0)
	for _, elem := range groups {
		converted = append(converted, fmt.Sprintf("\"%s\"", elem))
	}
	return fmt.Sprintf("[%s]", strings.Join(converted, ","))
}
func (client *Client) createResource(resource *Resource) error {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
			mutation{
			  resourceCreate(name: "%s", address: "%s", remoteNetworkId: "%s", protocols: %s, groupIds: %s) {
				ok
				error
				entity {
				  id
				}
			  }
		}
        `, resource.Name, resource.Address, resource.RemoteNetworkId, protocolsToGraphql(resource.Protocols), convertGroups(resource.Groups)),
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

func formatPorts(resourceData *gabs.Container, portPath string) []string {
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
func parseProtocols(resource *Resource, resourceData *gabs.Container) {
	resource.Protocols = &Protocols{
		AllowIcmp: resourceData.Path("protocols.allowIcmp").Data().(bool),
		UdpPolicy: resourceData.Path("protocols.udp.policy").Data().(string),
		TcpPolicy: resourceData.Path("protocols.tcp.policy").Data().(string),
	}
	resource.Protocols.TcpPorts = formatPorts(resourceData, "protocols.tcp.ports")
	resource.Protocols.UdpPorts = formatPorts(resourceData, "protocols.udp.ports")
}

func (client *Client) readResource(resourceId string) (*Resource, error) {
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  resource(id: "%s") {
			id
			name
			address {
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
	var groups = make([]string, 0)
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

	parseProtocols(resource, resourceQuery)

	return resource, nil
}

func (client *Client) updateResource(resource *Resource) error {
	//mutation := map[string]string{
	//	"query": fmt.Sprintf(`
	//			mutation {
	//				resourceUpdate(id: "%s", name: "%s"){
	//					ok
	//					error
	//				}
	//			}
	//    `, resourceId, resourceName),
	//}
	//mutationResource, err := client.doGraphqlRequest(mutation)
	//if err != nil {
	//	return fmt.Errorf("can't update remote network : %w", err)
	//}
	//
	//status := mutationResource.Path("data.resourceUpdate.ok").Data().(bool)
	//if !status {
	//	errorMessage := mutationResource.Path("data.resourceUpdate.error").Data().(string)
	//
	//	return APIError("can't update network: %s", errorMessage)
	//}

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
