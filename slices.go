package commons

import "slices"

// First returns the first element in the slice.
//
// If the slice is empty, the returned bool value will be false.
func First[H any](s []H) (v H, ok bool) {
	for _, v = range s {
		return v, true
	}

	return
}

// Last returns the last element in the slice.
//
// If the slice is empty, the returned bool value will be false.
func Last[H any](s []H) (v H, ok bool) {
	for _, v = range slices.Backward(s) {
		return v, true
	}

	return
}

// Any returns the first (any) key-value entry from looping over the map.
//
// If the map is empty, the returned bool value will be false.
//
// Useful if the map contains only one entry. If it has more than one, this method may return different entries upon
// subsequent calls.
func Any[H any](m map[string]H) (k string, v H, ok bool) {
	for k, v = range m {
		return k, v, true
	}

	return
}
