package query

type ReadResourcesByName struct {
	Resources `graphql:"resources(filter: {name: {eq: $name}}, after: $resourcesEndCursor, first: $pageLimit)"`
}

func (q ReadResourcesByName) IsEmpty() bool {
	return len(q.Edges) == 0
}
