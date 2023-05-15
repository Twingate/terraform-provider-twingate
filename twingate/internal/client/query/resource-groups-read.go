package query

import "github.com/hasura/go-graphql-client"

type ReadResourceGroups struct {
	Resource *gqlResourceGroups `graphql:"resource(id: $id)"`
}

func (q ReadResourceGroups) IsEmpty() bool {
	return q.Resource == nil
}

type gqlResourceGroups struct {
	ID     graphql.ID
	Groups Groups `graphql:"groups(after: $groupsEndCursor, first: $pageLimit)"`
}
