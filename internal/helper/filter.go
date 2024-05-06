package helper

func Filter[T any](slice []T, criteria func(T) bool) []T {
	var filtered []T
	for _, item := range slice {
		if criteria(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
