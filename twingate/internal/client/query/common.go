package query

import (
	"github.com/hasura/go-graphql-client"
)

type IDName struct {
	ID   graphql.ID `json:"id"`
	Name string     `json:"name"`
}

type OkError struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}
