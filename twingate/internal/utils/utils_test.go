package utils

import (
	"fmt"
	"testing"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
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

func TestMapDifference(t *testing.T) {
	cases := []struct {
		mapA     map[string]string
		mapB     map[string]string
		expected map[string]string
	}{
		{
			// Keys in A that don't appear in B survive.
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     map[string]string{"c": "3"},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			// Same key and value in B is removed; non-overlapping key survives.
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     map[string]string{"b": "2"},
			expected: map[string]string{"a": "1"},
		},
		{
			// Same key but different value in B survives: the entry is not B's.
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     map[string]string{"b": "99"},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			// Full overlap of identical entries → nil.
			mapA:     map[string]string{"a": "1"},
			mapB:     map[string]string{"a": "1"},
			expected: nil,
		},
		{
			// Nil A → nil.
			mapA:     nil,
			mapB:     map[string]string{"a": "1"},
			expected: nil,
		},
		{
			// Nil B → all of A survives.
			mapA:     map[string]string{"a": "1"},
			mapB:     nil,
			expected: map[string]string{"a": "1"},
		},
		{
			// Both nil → nil.
			mapA:     nil,
			mapB:     nil,
			expected: nil,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := MapDifference(c.mapA, c.mapB)

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestMapUnion(t *testing.T) {
	cases := []struct {
		mapA     map[string]string
		mapB     map[string]string
		expected map[string]string
	}{
		{
			mapA:     map[string]string{"a": "1", "b": "2", "c": "3"},
			mapB:     map[string]string{"d": "4", "e": "5", "f": "6"},
			expected: map[string]string{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5", "f": "6"},
		},
		{
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     map[string]string{"b": "3", "c": "4"},
			expected: map[string]string{"a": "1", "b": "3", "c": "4"}, // b key is overwritten by mapB's value
		},
		{
			mapA:     nil,
			mapB:     map[string]string{"a": "1", "b": "2"},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     nil,
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			mapA:     nil,
			mapB:     nil,
			expected: nil,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := MapUnion(c.mapA, c.mapB)

			assert.Equal(t, c.expected, actual)
		})
	}
}

// The API stores nothing for both a null and an empty map, so the response alone
// cannot tell them apart. State has to mirror what was declared or the attribute
// drifts on every plan.
func TestConvertMapValueWithReference(t *testing.T) {
	cases := []struct {
		name      string
		input     map[string]string
		reference types.Map
		expected  types.Map
	}{
		{
			name:      "populated response - converted regardless of state",
			input:     map[string]string{"x-a": "1"},
			reference: types.MapNull(types.StringType),
			expected:  stringMap(map[string]string{"x-a": "1"}),
		},
		{
			name:      "populated response overrides an empty state",
			input:     map[string]string{"x-a": "1"},
			reference: emptyStringMap(),
			expected:  stringMap(map[string]string{"x-a": "1"}),
		},
		{
			name:      "empty response, attribute omitted - stays null",
			input:     nil,
			reference: types.MapNull(types.StringType),
			expected:  types.MapNull(types.StringType),
		},
		{
			name:      "empty response, attribute declared empty - stays empty",
			input:     nil,
			reference: emptyStringMap(),
			expected:  emptyStringMap(),
		},
		{
			name:      "empty response, state populated - drift surfaces as null",
			input:     nil,
			reference: stringMap(map[string]string{"x-a": "1"}),
			expected:  types.MapNull(types.StringType),
		},
		{
			name:      "empty response, state unknown - resolves to null",
			input:     nil,
			reference: types.MapUnknown(types.StringType),
			expected:  types.MapNull(types.StringType),
		},
		{
			name:      "empty map response is treated the same as nil",
			input:     map[string]string{},
			reference: emptyStringMap(),
			expected:  emptyStringMap(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, ConvertMapValueWithReference(c.input, c.reference))
		})
	}
}
