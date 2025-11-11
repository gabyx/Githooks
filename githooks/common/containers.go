package common

import "slices"

// Any returns `true` if one of the strings in the slice
// satisfies the predicate `f`.
func Any(vs []any, f func(any) bool) bool {
	return slices.ContainsFunc(vs, f)
}

// All returns `true` if all of the strings in the slice
// satisfy the predicate `f`.
func All(vs []any, f func(any) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}

	return true
}

// Filter returns a new slice containing all strings in the
// slice that satisfy the predicate `f`.
func Filter(vs []any, f func(any) bool) []any {
	var vsf []any
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}

	return vsf
}

// Map returns a new slice containing the results of applying
// the function `f` to each string in the original slice.
func Map(vs []any, f func(any) any) []any {
	vsm := make([]any, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}

	return vsm
}
