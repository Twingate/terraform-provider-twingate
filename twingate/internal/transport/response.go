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
