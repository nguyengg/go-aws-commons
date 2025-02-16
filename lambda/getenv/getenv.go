package getenv

import (
	"context"
	"os"
)

// Variable defines ways to retrieve a variable.
type Variable[T any] interface {
	Get() (T, error)
	GetWithContext(ctx context.Context) (T, error)
	MustGet() T
	MustGetWithContext(ctx context.Context) T
}

// Env calls os.Getenv and returns that value in subsequent calls.
//
// See Getenv if you need something that calls os.Getenv on every invocation.
func Env(key string) Variable[string] {
	v := os.Getenv(key)
	return getter[string](func(ctx context.Context) (string, error) {
		return v, nil
	})
}

// EnvAs calls os.Getenv, decodes with m, then returns that value in subsequent calls.
//
// See GetenvAs if you need something that calls os.Getenv on every invocation.
func EnvAs[T any](key string, m func(string) (T, error)) Variable[T] {
	v, err := m(os.Getenv(key))
	return getter[T](func(ctx context.Context) (T, error) {
		return v, err
	})
}

// Getenv calls os.Getenv on every invocation and returns its value.
//
// Most of the time, Env suffices because environment variables are not updated that often. Use Getenv if you have a use
// case where the environment variables might be updated by some other processes.
func Getenv(key string) Variable[string] {
	return getter[string](func(ctx context.Context) (string, error) {
		return os.Getenv(key), nil
	})
}

// GetenvAs calls os.Getenv on every invocation, decodes with m, and returns its value.
//
// Most of the time, EnvAs suffices because environment variables are not updated that often. Use GetenvAs if you have
// a use case where the environment variables might be updated by some other processes.
func GetenvAs[T any](key string, m func(string) (T, error)) Variable[T] {
	return getter[T](func(ctx context.Context) (T, error) {
		return m(os.Getenv(key))
	})
}

// getter implements the Variable interface for a function.
type getter[T any] func(ctx context.Context) (T, error)

func (g getter[T]) Get() (T, error) {
	return g.GetWithContext(context.Background())
}

func (g getter[T]) GetWithContext(ctx context.Context) (T, error) {
	return g(ctx)
}

func (g getter[T]) MustGet() T {
	return g.MustGetWithContext(context.Background())
}

func (g getter[T]) MustGetWithContext(ctx context.Context) T {
	v, err := g(ctx)
	if err != nil {
		panic(err)
	}
	return v
}
