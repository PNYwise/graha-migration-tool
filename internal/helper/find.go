package helper

func Find[T any](data []T, criteria func(T) bool) (opt *T) {
	for i := range data {
		if criteria(data[i]) {
			return &data[i]
		}
	}
	return nil
}
