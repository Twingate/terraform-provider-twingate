package query

type ReadShallowResourcesWithType struct {
	ShallowResourcesWithType `graphql:"resources(after: $resourcesEndCursor, first: $pageLimit)"`
}

func (q ReadShallowResourcesWithType) IsEmpty() bool {
	return len(q.Edges) == 0
}

type ShallowResourcesWithType struct {
	PaginatedResource[*ShallowResourceEdge]
}

type ShallowResourceEdge struct {
	Node *gqlShallowResource
}

type gqlShallowResource struct {
	Type string `graphql:"__typename"`
	IDName
}
