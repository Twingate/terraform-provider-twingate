package twingate

import (
	"strconv"

	"github.com/hasura/go-graphql-client"
)

type IDName struct {
	ID   graphql.ID     `json:"id"`
	Name graphql.String `json:"name"`
}

func (in *IDName) StringID() string {
	return in.ID.(string)
}

func (in *IDName) StringName() string {
	return string(in.Name)
}

type OkError struct {
	Ok    graphql.Boolean `json:"ok"`
	Error graphql.String  `json:"error"`
}

type Edges struct {
	Node *IDName `json:"node"`
}

func newEmptyProtocols() *ProtocolsInput {
	pi := newProtocolsInput()
	pi.AllowIcmp = graphql.Boolean(true)
	pi.UDP.Policy = graphql.String(policyAllowAll)
	pi.TCP.Policy = graphql.String(policyAllowAll)

	return pi
}

type ProtocolsInput struct {
	UDP       *ProtocolInput  `json:"udp"`
	TCP       *ProtocolInput  `json:"tcp"`
	AllowIcmp graphql.Boolean `json:"allowIcmp"`
}

func (pi *ProtocolsInput) flattenProtocols() []interface{} {
	protocols := make(map[string]interface{})
	protocols["allow_icmp"] = pi.AllowIcmp

	if pi.TCP != nil {
		protocols["tcp"] = pi.TCP.flattenPorts()
	}

	if pi.UDP != nil {
		protocols["udp"] = pi.UDP.flattenPorts()
	}

	return []interface{}{protocols}
}

func (pi *ProtocolInput) flattenPorts() []interface{} {
	c := make(map[string]interface{})
	c["ports"], c["policy"] = pi.buildPortsRange()

	return []interface{}{c}
}

func newProtocolsInput() *ProtocolsInput {
	return &ProtocolsInput{
		TCP: &ProtocolInput{Ports: []*PortRangeInput{}},
		UDP: &ProtocolInput{Ports: []*PortRangeInput{}},
	}
}

type ProtocolInput struct {
	Ports  []*PortRangeInput `json:"ports"`
	Policy graphql.String    `json:"policy"`
}

func (pi *ProtocolInput) buildPortsRange() ([]string, string) {
	var ports []string

	for _, port := range pi.Ports {
		if port.Start == port.End {
			ports = append(ports, strconv.Itoa(int(port.Start)))
		} else {
			ports = append(ports, strconv.Itoa(int(port.Start))+"-"+strconv.Itoa(int(port.End)))
		}
	}

	return ports, string(pi.Policy)
}

type PortRangeInput struct {
	Start graphql.Int `json:"start"`
	End   graphql.Int `json:"end"`
}
