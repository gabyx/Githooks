package common

func CopySlice[T any](s []T) []T {
	res := make([]T, 0, len(s))
	return append(res, s...)
}

func CopySliceC[T any](s []T, capacity int) []T {
	res := make([]T, 0, capacity)
	return append(res, s...)
}
