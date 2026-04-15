package utils

import (
	"fmt"
	"maps"
	"strings"

	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Map - transform giving slice of items by applying the func.
func Map[T, R any](items []T, f func(item T) R) []R {
	result := make([]R, 0, len(items))

	for _, item := range items {
		result = append(result, f(item))
	}

	return result
}

func MapWithError[T, R any](items []T, f func(item T) (R, error)) ([]R, error) {
	result := make([]R, 0, len(items))

	for _, item := range items {
		val, err := f(item)
		if err != nil {
			return nil, err
		}

		result = append(result, val)
	}

	return result, nil
}

// Filter - filter down the elements from the given array that pass the test implemented by the provided function.
func Filter[T any](items []T, ok func(item T) bool) []T {
	result := make([]T, 0, len(items))

	for _, item := range items {
		if ok(item) {
			result = append(result, item)
		}
	}

	return result
}

// FilterMap - filter down the elements from the given array that pass the test implemented by the provided function and then transform by applying the transform func.
func FilterMap[T, R any](items []T, ok func(item T) bool, f func(item T) R) []R {
	result := make([]R, 0)

	for _, item := range items {
		if ok(item) {
			result = append(result, f(item))
		}
	}

	return result
}

// MapKeys - collects map keys to slice.
func MapKeys[T comparable](lookup map[T]bool) []T {
	result := make([]T, 0, len(lookup))

	for item := range lookup {
		result = append(result, item)
	}

	return result
}

// MakeLookupMap - creates lookup map from slice.
func MakeLookupMap[T comparable](items []T) map[T]bool {
	lookup := make(map[T]bool, len(items))

	for _, item := range items {
		lookup[item] = true
	}

	return lookup
}

func DocList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	default:
		n := len(items)
		last := items[n-1]

		return fmt.Sprintf("%s or %s", strings.Join(items[:n-1], ", "), last)
	}
}

func MakeStringSet(values []string) types.Set {
	return types.SetValueMust(types.StringType, Map(values, func(value string) tfattr.Value {
		return types.StringValue(value)
	}))
}

// MapUnion - for given two maps A and B,
// If A = {'a': 1, 'b': 2} and B = {'a': 3, 'c': 4}, then the union of A and B is {'a': 3, 'b': 2, 'c': 4}.
func MapUnion(mapA, mapB map[string]string) map[string]string {
	if len(mapA) == 0 {
		return mapB
	}

	if len(mapB) == 0 {
		return mapA
	}

	result := make(map[string]string, max(len(mapA), len(mapB)))
	maps.Copy(result, mapA)
	maps.Copy(result, mapB)

	return result
}

func ConvertMap(raw types.Map) map[string]string {
	if raw.IsNull() || raw.IsUnknown() || len(raw.Elements()) == 0 {
		return nil
	}

	result := make(map[string]string, len(raw.Elements()))

	for key, val := range raw.Elements() {
		result[key] = val.(types.String).ValueString()
	}

	return result
}

func ConvertMapValue(input map[string]string) types.Map {
	if len(input) == 0 {
		return types.MapNull(types.StringType)
	}

	raw := make(map[string]tfattr.Value, len(input))

	for key, val := range input {
		raw[key] = types.StringValue(val)
	}

	return types.MapValueMust(types.StringType, raw)
}

// MapDifference returns a map with all keys from mapA that are NOT present in mapB.
// Returns nil when the result would be empty.
func MapDifference(mapA, mapB map[string]string) map[string]string {
	if len(mapA) == 0 {
		return nil
	}

	result := make(map[string]string)

	for k, v := range mapA {
		if _, exists := mapB[k]; !exists {
			result[k] = v
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}
