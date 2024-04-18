package helper

func Find[T any](data []T, criteria func(T) bool) (opt T) {
	for _, element := range data {
		if criteria(element) {
			return element
		}
	}
	return
}
