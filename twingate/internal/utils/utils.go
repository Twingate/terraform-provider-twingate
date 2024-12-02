package utils

import (
	"fmt"
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

// Contains - checks if element exists in the slice.
func Contains[T comparable](items []T, element T) bool {
	for _, item := range items {
		if item == element {
			return true
		}
	}

	return false
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
