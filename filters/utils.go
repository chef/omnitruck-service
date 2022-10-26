package filters

func FilterList[T any](s []T, filter func(T) bool) []T {
	out := make([]T, len(s))

	counter := 0
	for i := 0; i < len(s); i++ {
		if filter(s[i]) {
			out[counter] = s[i]
			counter++
		}
	}
	return out[:counter]
}
