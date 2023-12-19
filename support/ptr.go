package support

func AsPtr[T any](in T) *T {
	return &in
}
