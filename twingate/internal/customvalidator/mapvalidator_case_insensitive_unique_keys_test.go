package customvalidator

import (
	"context"
	"testing"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestCaseInsensitiveUniqueKeys(t *testing.T) {
	buildMap := func(keys ...string) types.Map {
		elements := make(map[string]tfattr.Value, len(keys))
		for _, key := range keys {
			elements[key] = types.StringValue("value")
		}

		return types.MapValueMust(types.StringType, elements)
	}

	cases := []struct {
		name        string
		input       types.Map
		expectError bool
	}{
		{
			name:  "null map - no error",
			input: types.MapNull(types.StringType),
		},
		{
			name:  "unknown map - no error",
			input: types.MapUnknown(types.StringType),
		},
		{
			name:  "empty map - no error",
			input: types.MapValueMust(types.StringType, map[string]tfattr.Value{}),
		},
		{
			name:  "distinct keys - no error",
			input: buildMap("x-a", "x-b", "x-c"),
		},
		{
			name:        "keys differing only by case - error",
			input:       buildMap("X-Twingate-User", "x-twingate-user"),
			expectError: true,
		},
		{
			name:        "mixed case collision among distinct keys - error",
			input:       buildMap("x-a", "X-A", "x-b"),
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &validator.MapResponse{}

			CaseInsensitiveUniqueKeys().ValidateMap(context.Background(), validator.MapRequest{
				Path:        path.Root("request_header_rewrites"),
				ConfigValue: c.input,
			}, resp)

			assert.Equal(t, c.expectError, resp.Diagnostics.HasError())
		})
	}
}
