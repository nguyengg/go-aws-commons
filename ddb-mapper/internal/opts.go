package internal

// ApplyOpts applies the given optFns on the given v argument then return it.
func ApplyOpts[T any](v *T, optFns ...func(opts *T)) *T {
	for _, fn := range optFns {
		fn(v)
	}
	return v
}
