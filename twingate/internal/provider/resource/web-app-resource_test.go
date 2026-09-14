package resource

import (
	"context"
	"testing"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func emptyStringMap() types.Map {
	return types.MapValueMust(types.StringType, map[string]tfattr.Value{})
}

func stringMap(pairs map[string]string) types.Map {
	elements := make(map[string]tfattr.Value, len(pairs))
	for key, value := range pairs {
		elements[key] = types.StringValue(value)
	}

	return types.MapValueMust(types.StringType, elements)
}

// The API stores nothing for both a null and an empty map, so the response alone
// cannot tell them apart. State has to mirror what was declared or the attribute
// drifts on every plan.
func TestConvertHeaderRewrites(t *testing.T) {
	cases := []struct {
		name     string
		rewrites map[string]string
		state    types.Map
		expected types.Map
	}{
		{
			name:     "populated response - converted regardless of state",
			rewrites: map[string]string{"x-a": "1"},
			state:    types.MapNull(types.StringType),
			expected: stringMap(map[string]string{"x-a": "1"}),
		},
		{
			name:     "populated response overrides an empty state",
			rewrites: map[string]string{"x-a": "1"},
			state:    emptyStringMap(),
			expected: stringMap(map[string]string{"x-a": "1"}),
		},
		{
			name:     "empty response, attribute omitted - stays null",
			rewrites: nil,
			state:    types.MapNull(types.StringType),
			expected: types.MapNull(types.StringType),
		},
		{
			name:     "empty response, attribute declared empty - stays empty",
			rewrites: nil,
			state:    emptyStringMap(),
			expected: emptyStringMap(),
		},
		{
			name:     "empty response, state populated - drift surfaces as null",
			rewrites: nil,
			state:    stringMap(map[string]string{"x-a": "1"}),
			expected: types.MapNull(types.StringType),
		},
		{
			name:     "empty response, state unknown - resolves to null",
			rewrites: nil,
			state:    types.MapUnknown(types.StringType),
			expected: types.MapNull(types.StringType),
		},
		{
			name:     "empty map response is treated the same as nil",
			rewrites: map[string]string{},
			state:    emptyStringMap(),
			expected: emptyStringMap(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, convertHeaderRewrites(c.rewrites, c.state))
		})
	}
}

// getHeaderRewrites feeds the client, which sends an empty list for a nil map.
// Both null and empty must reach the API as "clear the stored rewrites".
func TestGetHeaderRewrites(t *testing.T) {
	cases := []struct {
		name     string
		input    types.Map
		expected map[string]string
	}{
		{
			name:     "null map - nil",
			input:    types.MapNull(types.StringType),
			expected: nil,
		},
		{
			name:     "unknown map - nil",
			input:    types.MapUnknown(types.StringType),
			expected: nil,
		},
		{
			name:     "empty map - nil",
			input:    emptyStringMap(),
			expected: nil,
		},
		{
			name:     "populated map - converted",
			input:    stringMap(map[string]string{"x-a": "1", "x-b": "2"}),
			expected: map[string]string{"x-a": "1", "x-b": "2"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, getHeaderRewrites(c.input))
		})
	}
}

func TestWebAppUpstreamRoundTrip(t *testing.T) {
	ctx := context.Background()

	obj, diags := webAppUpstreamObject(ctx, 8080)
	require.False(t, diags.HasError())

	upstream, diags := webAppUpstreamValue(ctx, obj)
	require.False(t, diags.HasError())

	assert.Equal(t, int64(8080), upstream.Port.ValueInt64())
}

func TestWebAppDownstreamRoundTrip(t *testing.T) {
	ctx := context.Background()

	obj, diags := webAppDownstreamObject(ctx, 80)
	require.False(t, diags.HasError())

	downstream, diags := webAppDownstreamValue(ctx, obj)
	require.False(t, diags.HasError())

	assert.Equal(t, int64(80), downstream.Port.ValueInt64())
}
