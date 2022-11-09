package utils

func Map[T, R any](items []T, f func(item T) R) []R {
	result := make([]R, 0, len(items))

	for _, item := range items {
		result = append(result, f(item))
	}

	return result
}
