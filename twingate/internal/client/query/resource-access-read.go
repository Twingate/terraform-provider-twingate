package query

import "github.com/hasura/go-graphql-client"

type ReadResourceAccess struct {
	Resource *gqlResourceAccess `graphql:"resource(id: $id)"`
}

func (q ReadResourceAccess) IsEmpty() bool {
	return q.Resource == nil
}

type gqlResourceAccess struct {
	ID     graphql.ID
	Access Access `graphql:"access(after: $accessEndCursor, first: $pageLimit)"`
}
