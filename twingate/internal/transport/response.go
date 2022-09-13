package transport

import (
	"github.com/twingate/go-graphql-client"
)

type IDName struct {
	ID   graphql.ID     `json:"id"`
	Name graphql.String `json:"name"`
}

func (in *IDName) StringID() string {
	return idToString(in.ID)
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

func (e Edges) GetName() string {
	return e.Node.StringName()
}

func (e Edges) GetID() string {
	return e.Node.StringID()
}

//func NewEmptyProtocols() *Protocols {
//	pi := NewProtocolsInput()
//	pi.AllowIcmp = graphql.Boolean(true)
//	pi.UDP.Policy = graphql.String(model.PolicyAllowAll)
//	pi.TCP.Policy = graphql.String(model.PolicyAllowAll)
//
//	return pi
//}

//func (pi *Protocols) FlattenProtocols() []interface{} {
//	if pi == nil {
//		return nil
//	}
//
//	protocols := make(map[string]interface{})
//	protocols["allow_icmp"] = pi.AllowIcmp
//
//	if pi.TCP != nil {
//		protocols["tcp"] = pi.TCP.flattenPorts()
//	}
//
//	if pi.UDP != nil {
//		protocols["udp"] = pi.UDP.flattenPorts()
//	}
//
//	return []interface{}{protocols}
//}
//
//func (pi *Protocol) flattenPorts() []interface{} {
//	c := make(map[string]interface{})
//	c["ports"], c["policy"] = pi.BuildPortsRange()
//
//	return []interface{}{c}
//}
//
//func NewProtocolsInput() *Protocols {
//	return &Protocols{
//		TCP: &Protocol{Ports: []*PortRange{}},
//		UDP: &Protocol{Ports: []*PortRange{}},
//	}
//}

//func (pi *Protocol) BuildPortsRange() ([]string, string) {
//	var ports []string
//
//	for _, port := range pi.Ports {
//		if port.Start == port.End {
//			ports = append(ports, strconv.Itoa(int(port.Start)))
//		} else {
//			ports = append(ports, strconv.Itoa(int(port.Start))+"-"+strconv.Itoa(int(port.End)))
//		}
//	}
//
//	return ports, string(pi.Policy)
//}
