package metrics

import "context"

type metricsKey struct{}

// WithContext returns a new child context with the given metrics attached.
//
// The metrics can be retrieved using Get(child) or Get(child).
func WithContext(ctx context.Context, m *Metrics) context.Context {
	return context.WithValue(ctx, metricsKey{}, m)
}

// NewWithContext combines both New and WithContext in one call.
func NewWithContext(ctx context.Context, optFns ...func(m *Metrics)) (context.Context, *Metrics) {
	m := &Metrics{}
	m.init()

	for _, fn := range optFns {
		fn(m)
	}

	return context.WithValue(ctx, metricsKey{}, m), m
}

// Get returns the Metrics instance from the specified context if available.
//
// If the context doesn't contain an instance, a new one will be created which most likely is not the right expectation
// since whoever creates the Metrics instance is often responsible for closing it (in order to log it). Prefer MustGet
// which will panic if the context does not contain an existing Metrics instance.
func Get(ctx context.Context) *Metrics {
	if m, ok := ctx.Value(metricsKey{}).(*Metrics); ok && m != nil {
		return m
	}

	return New()
}

// MustGet is a variant of Get/TryGet that panics if no Metrics instance is found from context.
func MustGet(ctx context.Context) *Metrics {
	if m, ok := ctx.Value(metricsKey{}).(*Metrics); ok {
		return m
	}

	panic("metrics not found in context")
}

// TryGet is a variant of Get that does not return a new Metrics instance.
//
// Use this if you absolutely need an existing Metrics instance to exist.
func TryGet(ctx context.Context) (*Metrics, bool) {
	m, ok := ctx.Value(metricsKey{}).(*Metrics)
	return m, ok
}
