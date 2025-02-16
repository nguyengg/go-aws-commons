package getenv

import "context"

// err implements the Variable interface for an error.
type errVar[T any] struct {
	err error
}

func (e errVar[T]) Get() (v T, err error) {
	err = e.err
	return
}

func (e errVar[T]) GetWithContext(_ context.Context) (v T, err error) {
	err = e.err
	return
}

func (e errVar[T]) MustGet() T {
	panic(e.err)
}

func (e errVar[T]) MustGetWithContext(_ context.Context) T {
	panic(e.err)
}

var _ Variable[any] = &errVar[any]{}
var _ Variable[any] = (*errVar[any])(nil)
