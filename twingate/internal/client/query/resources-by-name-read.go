package query

type ReadResourcesByName struct {
	Resources `graphql:"resources(filter: $filter, after: $resourcesEndCursor, first: $pageLimit)"`
}

func (q ReadResourcesByName) IsEmpty() bool {
	return len(q.Edges) == 0
}

type ResourceFilterInput struct {
	Name *StringFilterOperationInput `json:"name"`
}

func NewResourceFilterInput(name, filter string) *ResourceFilterInput {
	return &ResourceFilterInput{
		Name: NewStringFilterOperationInput(name, filter),
	}
}
