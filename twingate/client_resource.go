package twingate

import (
	"errors"
	"fmt"
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

type Resources struct {
	ID   string
	Name string
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

type createResourceResponse struct {
	Data *struct {
		ResourceCreate *struct {
			*OkErrorResponse
			Entity *struct {
				ID string `json:"id"`
			} `json:"entity"`
		} `json:"resourceCreate"`
	} `json:"data"`
}

func (r *createResourceResponse) checkErrors() []*queryResponseErrors {
	return nil
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

	r := createResourceResponse{}

	err = client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIError(err, "create", resourceResourceName)
	}

	if !r.Data.ResourceCreate.Ok {
		return NewAPIError(NewMutationError(r.Data.ResourceCreate.Error), "create", resourceResourceName)
	}

	resource.ID = r.Data.ResourceCreate.Entity.ID

	return nil
}

type readResourceResponse struct {
	Errors []*queryResponseErrors `json:"errors"`
	Data   *struct {
		Resource *readResourceResponseDataResource `json:"resource"`
	} `json:"data"`
}

func (r *readResourceResponse) checkErrors() []*queryResponseErrors {
	return r.Errors
}

type readResourceResponseDataResource struct {
	*IDNameResponse
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
		UDP       *readResourceResponseProtocol `json:"udp"`
		TCP       *readResourceResponseProtocol `json:"tcp"`
		AllowIcmp bool                          `json:"allowIcmp"`
	} `json:"protocols"`
}

type readResourceResponseProtocol struct {
	Ports  []*readResourceResponseProtocolsPorts `json:"ports"`
	Policy string                                `json:"policy"`
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
		return nil, NewAPIErrorWithID(err, "read", remoteNetworkResourceName, resourceID)
	}

	if r.Data.Resource == nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	var groups = make([]string, 0)

	if r.Data.Resource.Groups.PageInfo.HasNextPage {
		return nil, NewAPIErrorWithID(ErrTooManyGroupsError, "read", resourceResourceName, resourceID)
	}

	for _, elem := range r.Data.Resource.Groups.Edges {
		groups = append(groups, elem.Node.ID)
	}

	protocols := &Protocols{
		AllowIcmp: r.Data.Resource.Protocols.AllowIcmp,
		TCPPolicy: r.Data.Resource.Protocols.TCP.Policy,
		UDPPolicy: r.Data.Resource.Protocols.UDP.Policy,
		TCPPorts:  extractPortsFromResults(r.Data.Resource.Protocols.TCP.Ports),
		UDPPorts:  extractPortsFromResults(r.Data.Resource.Protocols.UDP.Ports),
	}

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

type readResourcesResponse struct { //nolint
	Error *struct {
		Errors []*queryResponseErrors `json:"errors"`
	} `json:"error"`
	Data struct {
		Resources struct {
			Edges []*EdgesResponse `json:"edges"`
		} `json:"resources"`
	} `json:"data"`
}

func (r *readResourcesResponse) checkErrors() []*queryResponseErrors {
	if r.Error != nil {
		return r.Error.Errors
	}
	return nil
}

func (client *Client) readResources() (map[int]*Resources, error) { //nolint
	query := map[string]string{
		"query": "{ resources { edges { node { id name } } } }",
	}

	r := readResourcesResponse{}

	err := client.doGraphqlRequest(query, &r)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	var resources = make(map[int]*Resources)

	for i, elem := range r.Data.Resources.Edges {
		c := &Resources{ID: elem.Node.ID, Name: elem.Node.Name}
		resources[i] = c
	}

	return resources, nil
}

type updateResourceResponse struct {
	Data *struct {
		ResourceUpdate *OkErrorResponse `json:"resourceUpdate"`
	} `json:"data"`
}

func (r *updateResourceResponse) checkErrors() []*queryResponseErrors {
	return nil
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

	r := updateResourceResponse{}

	err = client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	if !r.Data.ResourceUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(r.Data.ResourceUpdate.Error), "update", resourceResourceName, resource.ID)
	}

	return nil
}

type deleteResourceResponse struct {
	Data *struct {
		ResourceDelete *OkErrorResponse `json:"resourceDelete"`
	} `json:"data"`
}

func (r *deleteResourceResponse) checkErrors() []*queryResponseErrors {
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

	r := deleteResourceResponse{}

	err := client.doGraphqlRequest(mutation, &r)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", resourceResourceName, resourceID)
	}

	if !r.Data.ResourceDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(r.Data.ResourceDelete.Error), "delete", resourceResourceName, resourceID)
	}

	return nil
}
