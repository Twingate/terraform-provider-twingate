package utils

// Map - transform giving slice of items by applying the func
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

// Contains - checks if element exists in the slice
func Contains[T comparable](items []T, element T) bool {
	lookup := MakeLookupMap[T](items)

	return lookup[element]
}

// MapKeys - collects map keys to slice
func MapKeys[T comparable](lookup map[T]bool) []T {
	result := make([]T, 0, len(lookup))

	for item := range lookup {
		result = append(result, item)
	}

	return result
}

func MakeLookupMap[T comparable](items []T) map[T]bool {
	lookup := make(map[T]bool, len(items))

	for _, item := range items {
		lookup[item] = true
	}

	return lookup
}
