package gospel

func If[T any](condition bool, value T) T {
	if condition {
		return value
	}
	return *new(T)
}
