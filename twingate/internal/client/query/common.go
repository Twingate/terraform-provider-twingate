package query

import (
	"github.com/hasura/go-graphql-client"
)

type IDName struct {
	ID   graphql.ID `json:"id"`
	Name string     `json:"name"`
}

func (node IDName) GetID() string {
	return string(node.ID)
}

type OkError struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func (ok OkError) OK() bool {
	return ok.Ok
}

func (ok OkError) ErrorStr() string {
	return ok.Error
}
