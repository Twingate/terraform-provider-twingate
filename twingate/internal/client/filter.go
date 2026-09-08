package client

import (
	attrs "github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client/query"
)

// StringFilter describes a filter on a single string field: either a single
// value matched with one of the operations from attr (exact, prefix, regexp, ...)
// or, for attr.FilterByIn, a list of values matched exactly.
type StringFilter struct {
	Name   string
	Filter string
	Values []string
}

func (f *StringFilter) IsEmpty() bool {
	return f == nil || f.Name == "" && len(f.Values) == 0
}

func (f *StringFilter) ToQuery() *query.StringFilterOperationInput {
	if f.IsEmpty() {
		return nil
	}

	if f.Filter == attrs.FilterByIn {
		return query.NewStringFilterInOperationInput(f.Values)
	}

	return query.NewStringFilterOperationInput(f.Name, f.Filter)
}
