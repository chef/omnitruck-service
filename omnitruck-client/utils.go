package omnitruck_client

func FilterList[T any](s []T, filter func(T) bool) []T {
	out := make([]T, len(s))

	counter := 0
	for i := 0; i < len(s); i++ {
		if !filter(s[i]) {
			out[counter] = s[i]
			counter++
		}
	}
	return out[:counter]
}

func FilterProductList[T any](s []T, product string, filter func(string, T) bool) []T {
	out := make([]T, len(s))

	counter := 0
	for i := 0; i < len(s); i++ {
		if !filter(product, s[i]) {
			out[counter] = s[i]
			counter++
		}
	}
	return out[:counter]
}
