package twingate

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	ID              string
	RemoteNetworkID string
	Address         string
	Name            string
	GroupsIds       []string
	Protocols       *Protocols
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

func extractPortsFromResults(ports []*readResourceResponseProtocolsPorts) []string {
	var parsedPorts = make([]string, 0)

	for _, port := range ports {
		if port.Start == port.End {
			parsedPorts = append(parsedPorts, fmt.Sprintf("%d", port.Start))
		} else {
			parsedPorts = append(parsedPorts, fmt.Sprintf("%d-%d", port.Start, port.End))
		}
	}

	return parsedPorts
}

type CreateResourceResponse struct {
	Data CreateResourceResponseData `json:"data"`
}

type CreateResourceResponseData struct {
	ResourceCreate CreateResourceResponseDataResourceCreate `json:"resourceCreate"`
}

type CreateResourceResponseDataResourceCreate struct {
	Ok     bool                                            `json:"ok"`
	Error  string                                          `json:"error"`
	Entity *CreateResourceResponseDataresourceCreateEntity `json:"entity"`
}

type CreateResourceResponseDataresourceCreateEntity struct {
	ID string `json:"id"`
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

	r := CreateResourceResponse{}
	err = client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIError(err, "create", resourceResourceName)
	}

	if !r.Data.ResourceCreate.Ok {
		message := r.Data.ResourceCreate.Error
		return NewAPIError(NewMutationError(message), "create", resourceResourceName)
	}

	resource.ID = r.Data.ResourceCreate.Entity.ID

	return nil
}

type readResourceResponse struct {
	Data *readResourceResponseData `json:"data"`
}

type readResourceResponseData struct {
	Resource *readResourceResponseDataResource `json:"resource"`
}

type readResourceResponseDataResource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"address"`
	RemoteNetwork struct {
		ID string `json:"id"`
	} `json:"remoteNetwork"`
	Groups struct {
		PageInfo struct {
			HasNextPage bool `json:"hasNextPage"`
		} `json:"pageInfo"`
		Edges []struct {
			Node struct {
				ID string `json:"id"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"groups"`
	Protocols struct {
		UDP struct {
			Ports  []*readResourceResponseProtocolsPorts `json:"ports"`
			Policy string                                `json:"policy"`
		} `json:"udp"`
		TCP struct {
			Ports  []*readResourceResponseProtocolsPorts `json:"ports"`
			Policy string                                `json:"policy"`
		} `json:"tcp"`
		AllowIcmp bool `json:"allowIcmp"`
	} `json:"protocols"`
}

type readResourceResponseProtocolsPorts struct {
	End   int `json:"end"`
	Start int `json:"start"`
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
	r := readResourceResponse{}
	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	log.Println(r.Data.Resource)

	if r.Data.Resource == nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	var groups = make([]string, 0)

	if r.Data.Resource.Groups.PageInfo.HasNextPage {
		return nil, NewAPIErrorWithID(ErrTooManyGroupsError, "read", resourceResourceName, resourceID)
	}

	for _, elem := range r.Data.Resource.Groups.Edges {
		log.Println(elem)
		groups = append(groups, elem.Node.ID)
	}

	protocols := &Protocols{}
	protocols.AllowIcmp = r.Data.Resource.Protocols.AllowIcmp
	protocols.TCPPorts = extractPortsFromResults(r.Data.Resource.Protocols.TCP.Ports)
	protocols.UDPPorts = extractPortsFromResults(r.Data.Resource.Protocols.UDP.Ports)

	resource := &Resource{
		ID:              resourceID,
		Name:            r.Data.Resource.Name,
		Address:         r.Data.Resource.Address.Value,
		RemoteNetworkID: r.Data.Resource.RemoteNetwork.ID,
		GroupsIds:       groups,
		Protocols:       protocols,
	}

	return resource, nil
}

type updateResourceResponse struct {
	Data *updateResourceResponseData `json:"data"`
}

type updateResourceResponseData struct {
	ResourceUpdate *updateResourceResponseResourceUpdate `json:"resourceUpdate"`
}

type updateResourceResponseResourceUpdate struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func newUpdateResourceResponse() *updateResourceResponse {
	return &updateResourceResponse{
		Data: &updateResourceResponseData{
			ResourceUpdate: &updateResourceResponseResourceUpdate{},
		},
	}
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
	r := newUpdateResourceResponse()

	err = client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	if !r.Data.ResourceUpdate.Ok {
		message := r.Data.ResourceUpdate.Error
		return NewAPIErrorWithID(NewMutationError(message), "update", resourceResourceName, resource.ID)
	}

	return nil
}

type deleteResourceResponse struct {
	Data *deleteResourceResponseData `json:"data"`
}

type deleteResourceResponseData struct {
	ResourceDelete *deleteResourceResponseDataResourceDelete `json:"resourceDelete"`
}

type deleteResourceResponseDataResourceDelete struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func newDeleteResourceResponse() *deleteResourceResponse {
	return &deleteResourceResponse{
		Data: &deleteResourceResponseData{
			ResourceDelete: &deleteResourceResponseDataResourceDelete{},
		},
	}
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

	r := newDeleteResourceResponse()

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", resourceResourceName, resourceID)
	}

	if !r.Data.ResourceDelete.Ok {
		message := r.Data.ResourceDelete.Error
		return NewAPIErrorWithID(NewMutationError(message), "delete", resourceResourceName, resourceID)
	}

	return nil
}
