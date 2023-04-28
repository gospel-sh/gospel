package gospel

func If[T any](condition bool, value T) T {
	if condition {
		return value
	}
	return *new(T)
}
func DoIf[T any](condition bool, value func() T) T {
	if condition {
		return value()
	}
	return *new(T)
}
