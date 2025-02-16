package getenv

import (
	"context"
	"encoding/base64"
)

// Map provides a way to transform the original variable into another type.
//
// Useful if you're getting base64-encoded strings from either environment variables or AWS ParameterAs Store and Secrets
// Manager Lambda extension.
func Map[In any, Out any](v Variable[In], m func(In) (Out, error)) Variable[Out] {
	return mapper[In, Out]{v, m}
}

// WithBase64Encoding can be used to automatically base64-decode the variable.
func WithBase64Encoding(v Variable[string], encoding *base64.Encoding) Variable[[]byte] {
	return Map[string, []byte](v, encoding.DecodeString)
}

// mapper implements the Variable interface with a mapping function.
type mapper[In any, Out any] struct {
	v Variable[In]
	m func(In) (Out, error)
}

func (m mapper[In, Out]) Get() (Out, error) {
	return m.GetWithContext(context.Background())
}

func (m mapper[In, Out]) GetWithContext(ctx context.Context) (v Out, err error) {
	i, err := m.v.GetWithContext(ctx)
	if err != nil {
		return v, err
	}

	return m.m(i)
}

func (m mapper[In, Out]) MustGet() Out {
	return m.MustGetWithContext(context.Background())
}

func (m mapper[In, Out]) MustGetWithContext(ctx context.Context) (v Out) {
	i, err := m.v.GetWithContext(ctx)
	if err != nil {
		panic(err)
	}

	v, err = m.m(i)
	if err != nil {
		panic(err)
	}

	return
}

var _ Variable[any] = &mapper[any, any]{}
var _ Variable[any] = (*mapper[any, any])(nil)
