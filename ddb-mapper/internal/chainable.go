package internal

// ChainableFunc can chain functions together using And.
type ChainableFunc func()

// Chainable creates a new chain of functions.
//
// The returned function is never nil.
func Chainable(fns ...func()) ChainableFunc {
	if len(fns) == 0 {
		return func() {}
	}

	return func() {
		for _, fn := range fns {
			fn()
		}
	}
}

// And adds a new function to this chain.
func (fn ChainableFunc) And(f func()) ChainableFunc {
	if fn == nil {
		return f
	}

	if f == nil {
		return fn
	}

	return func() {
		fn()
		f()
	}
}
