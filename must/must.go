package must

// Must panics if the second argument is a non-nil error.
//
// Save you from having to type out something like:
//
//	if a, err := someFunction(); err != nil {
//	    panic(err)
//	}
//
// Which can be replaced with:
//
//	import . "github.com/nguyengg/go-aws-commons/must"
//	a := Must(someFunction())
func Must[H any](value H, err error) H {
	if err != nil {
		panic(err)
	}

	return value
}

// Must0 panics if the single argument is a non-nil error.
//
// Save you from having to type out something like:
//
//	if err := someFunction(); err != nil {
//	    panic(err)
//	}
//
// Which can be replaced with:
//
//	import . "github.com/nguyengg/go-aws-commons/must"
//	Must0(someFunction())
func Must0(err error) {
	if err != nil {
		panic(err)
	}
}

// Must2 panics if the last argument is a non-nil error.
//
// Save you from having to type out something like:
//
//	if a, b, err := someFunction(); err != nil {
//	    panic(err)
//	}
//
// Which can be replaced with:
//
//	import . "github.com/nguyengg/go-aws-commons/must"
//	a, b := Must2(someFunction())
func Must2[A any, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}

	return a, b
}

// Must3 panics if the last argument is a non-nil error.
//
// Save you from having to type out something like:
//
//	if a, b, c, err := someFunction(); err != nil {
//	    panic(err)
//	}
//
// Which can be replaced with:
//
//	import . "github.com/nguyengg/go-aws-commons/must"
//	a, b, c := Must3(someFunction())
func Must3[A any, B any, C any](a A, b B, c C, err error) (A, B, C) {
	if err != nil {
		panic(err)
	}

	return a, b, c
}
