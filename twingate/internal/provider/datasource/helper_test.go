package datasource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSetValues(t *testing.T) {
	cases := []struct {
		name     string
		input    types.Set
		expected []string
	}{
		{
			name:     "null set",
			input:    types.SetNull(types.StringType),
			expected: nil,
		},
		{
			name:     "unknown set",
			input:    types.SetUnknown(types.StringType),
			expected: nil,
		},
		{
			name:     "empty set",
			input:    types.SetValueMust(types.StringType, []attr.Value{}),
			expected: nil,
		},
		{
			name: "values are sorted",
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("b"),
				types.StringValue("a"),
			}),
			expected: []string{"a", "b"},
		},
		{
			name: "null, unknown and empty elements are skipped",
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue(""),
			}),
			expected: []string{"a"},
		},
		{
			name: "only null, unknown and empty elements",
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue(""),
			}),
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, SetValues(c.input))
		})
	}
}
