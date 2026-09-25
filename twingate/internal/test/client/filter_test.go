package client

import (
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client/query"
	"github.com/stretchr/testify/assert"
)

func TestStringFilterToQuery(t *testing.T) {
	cases := []struct {
		name     string
		filter   *client.StringFilter
		expected *query.StringFilterOperationInput
	}{
		{
			name:     "nil filter",
			filter:   nil,
			expected: nil,
		},
		{
			name:     "empty filter",
			filter:   &client.StringFilter{},
			expected: nil,
		},
		{
			name:     "exact match",
			filter:   &client.StringFilter{Name: "test"},
			expected: &query.StringFilterOperationInput{Eq: optionalString("test")},
		},
		{
			name:     "prefix match",
			filter:   &client.StringFilter{Name: "test", Filter: attr.FilterByPrefix},
			expected: &query.StringFilterOperationInput{StartsWith: optionalString("test")},
		},
		{
			name:     "list match",
			filter:   &client.StringFilter{Filter: attr.FilterByIn, Values: []string{"a", "b"}},
			expected: &query.StringFilterOperationInput{In: []string{"a", "b"}},
		},
		{
			name:     "list match wins over the single value",
			filter:   &client.StringFilter{Name: "test", Filter: attr.FilterByIn, Values: []string{"a"}},
			expected: &query.StringFilterOperationInput{In: []string{"a"}},
		},
		{
			name:     "list filter without values",
			filter:   &client.StringFilter{Filter: attr.FilterByIn},
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.filter.ToQuery())
			assert.Equal(t, c.expected == nil, c.filter.IsEmpty())
		})
	}
}
