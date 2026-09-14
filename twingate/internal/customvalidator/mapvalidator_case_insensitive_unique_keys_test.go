package customvalidator

import (
	"context"
	"testing"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		name           string
		input          types.Map
		expectedDetail string
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
			name:           "collision among distinct keys - reports the colliding pair",
			input:          buildMap("X-Twingate-User", "x-other", "x-twingate-user"),
			expectedDetail: `Keys "X-Twingate-User" and "x-twingate-user" differ only by case.`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &validator.MapResponse{}

			CaseInsensitiveUniqueKeys().ValidateMap(context.Background(), validator.MapRequest{
				Path:        path.Root("request_header_rewrites"),
				ConfigValue: c.input,
			}, resp)

			if c.expectedDetail == "" {
				assert.False(t, resp.Diagnostics.HasError())

				return
			}

			require.True(t, resp.Diagnostics.HasError())
			require.Len(t, resp.Diagnostics.Errors(), 1)
			assert.Equal(t, "Duplicate key", resp.Diagnostics.Errors()[0].Summary())
			assert.Equal(t, c.expectedDetail, resp.Diagnostics.Errors()[0].Detail())
		})
	}
}
