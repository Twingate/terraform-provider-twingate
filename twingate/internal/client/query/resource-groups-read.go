package query

import "github.com/twingate/go-graphql-client"

type ReadResourceGroups struct {
	Resource *gqlResourceGroups `graphql:"resource(id: $id)"`
}

type gqlResourceGroups struct {
	ID     graphql.ID
	Groups Groups `graphql:"groups(after: $groupsEndCursor)"`
}
