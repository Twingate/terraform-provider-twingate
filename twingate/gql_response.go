package twingate

import "github.com/hasura/go-graphql-client"

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
	pi.AllowIcmp = true
	pi.UDP.Policy = "ALLOW_ALL"
	pi.TCP.Policy = "ALLOW_ALL"

	return pi
}

type ProtocolsInput struct {
	UDP       *ProtocolInput  `json:"udp"`
	TCP       *ProtocolInput  `json:"tcp"`
	AllowIcmp graphql.Boolean `json:"allowIcmp"`
}

func newProtocolsInput() *ProtocolsInput {
	tcpPorts := []*PortRangeInput{}
	udpPorts := []*PortRangeInput{}
	tcp := &ProtocolInput{Ports: tcpPorts}
	udp := &ProtocolInput{Ports: udpPorts}

	return &ProtocolsInput{
		TCP: tcp,
		UDP: udp,
	}
}

type ProtocolInput struct {
	Ports  []*PortRangeInput `json:"ports"`
	Policy graphql.String    `json:"policy"`
}

type PortRangeInput struct {
	Start graphql.Int `json:"start"`
	End   graphql.Int `json:"end"`
}

// func parseErrors(Errors []*queryErrors) []graphql.String {
// 	messages := []graphql.String{}

// 	for _, e := range Errors {
// 		messages = append(messages, e.Message)
// 	}

// 	return messages
// }
