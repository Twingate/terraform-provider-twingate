package query

import "github.com/hasura/go-graphql-client"

type ReadResourceServiceAccounts struct {
	Resource *gqlResourceServiceAccounts `graphql:"resource(id: $id)"`
}

func (q ReadResourceServiceAccounts) IsEmpty() bool {
	return q.Resource == nil
}

type gqlResourceServiceAccounts struct {
	ID              graphql.ID
	ServiceAccounts ServiceAccounts `graphql:"serviceAccounts(after: $servicesEndCursor, first: $pageLimit)"`
}
