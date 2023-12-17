package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
)

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
	var nameFilter StringFilterOperationInput

	switch filter {
	case attr.FilterByRegexp:
		nameFilter.Regexp = &name
	case attr.FilterByContains:
		nameFilter.Contains = &name
	case attr.FilterByExclude:
		nameFilter.Ne = &name
	case attr.FilterByPrefix:
		nameFilter.StartsWith = &name
	case attr.FilterBySuffix:
		nameFilter.EndsWith = &name
	default:
		nameFilter.Eq = &name
	}

	return &ResourceFilterInput{
		Name: &nameFilter,
	}
}
