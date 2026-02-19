package metrics

import (
	"context"
	"log/slog"
	"sync"

	"github.com/rs/zerolog"
)

// Factory is used to create Metrics instances with prepopulated metrics (properties and counters).
//
// Factory is useful when you have properties (e.g. hostname, thread name, process ID, etc.) that you want to exist in
// all Metrics instances created by the Factory. It also provides useful methods to create slog.Logger and
// zerolog.Logger instances with the same prepopulated attributes.
//
// The zero-value DefaultFactory is ready for use. New uses DefaultFactory.
type Factory struct {
	// Logger controls how the Metrics instance is actually logged when Metrics.Close is called.
	Logger Logger
	// ParentSlogLogger is the slog.Logger that is used by NewSlogLogger to create a child logger.
	//
	// Default to slog.Default.
	ParentSlogLogger *slog.Logger
	// ParentZerologLogger is the zerolog.Logger that is used by NewZerologLogger to create a child logger.
	//
	// Default to zerolog.DefaultContextLogger.
	ParentZerologLogger *zerolog.Logger

	properties map[string]property
	counters   map[string]counter

	once sync.Once
}

// New creates a new Metrics instance and sets its start time to time.Now.
func (f *Factory) New(optFns ...func(m *Metrics)) *Metrics {
	m := &Metrics{logger: f.Logger}
	m.init()

	for k, v := range f.properties {
		m.properties[k] = property{v.t, v.v}
	}

	if len(f.counters) != 0 {
		for k, c := range f.counters {
			m.counters[k] = counter{c.t, c.v}
		}
	}

	for _, fn := range optFns {
		fn(m)
	}

	return m
}

// String creates or modifies a string key-value property pair.
//
// See [Metrics.String].
func (f *Factory) String(key, value string) *Factory {
	f.init()

	if reservedKeys[key] {
		return f
	}

	f.properties[key] = property{stringKind, value}

	return f
}

// Int64 creates or modifies an int64 key-value property pair.
//
// See [Metrics.Int64].
func (f *Factory) Int64(key string, value int64) *Factory {
	f.init()

	if reservedKeys[key] {
		return f
	}

	f.properties[key] = property{int64Kind, value}

	return f
}

// Float64 creates or modifies a float64 key-value property pair.
//
// See [Metrics.Float64].
func (f *Factory) Float64(key string, value float64) *Factory {
	f.init()

	if reservedKeys[key] {
		return f
	}

	f.properties[key] = property{float64Kind, value}

	return f
}

// Any is a variant of String that accepts any value instead.
//
// See [Metrics.Any].
func (f *Factory) Any(key string, value any) *Factory {
	f.init()

	if reservedKeys[key] {
		return f
	}

	f.properties[key] = property{anyKind, value}

	return f
}

// SetCounter sets the Factory.Counters mapping with the specified key to the given value.
//
// See [Metrics.SetCounter].
func (f *Factory) SetCounter(key string, value int64, ensureExist ...string) *Factory {
	f.init()

	f.counters[key] = counter{int64Kind, value}

	for _, k := range ensureExist {
		if _, ok := f.counters[k]; !ok {
			f.counters[k] = counter{int64Kind, int64(0)}
		}
	}

	return f
}

// SetFloater sets the Metrics.Floaters mapping with the specified key to the given value.
//
// See [Metrics.SetFloater].
func (f *Factory) SetFloater(key string, value float64, ensureExist ...string) *Factory {
	f.init()

	f.counters[key] = counter{float64Kind, value}

	for _, k := range ensureExist {
		if _, ok := f.counters[k]; !ok {
			f.counters[k] = counter{float64Kind, int64(0)}
		}
	}

	return f
}

// DefaultFactory is the globally accessible factory for creating Metrics instances.
//
// New uses DefaultFactory.
var DefaultFactory = &Factory{}

func (f *Factory) init() {
	f.once.Do(func() {
		f.properties = make(map[string]property)
		f.counters = make(map[string]counter)
	})
}

// NewSlogLogger creates a new child slog.Logger with attributes filled out from the factory.
//
// The first given logger will be used as the parent logger. If none is given, Factory.ParentSlogLogger is used, which
// itself defaults to slog.Default.
func (f *Factory) NewSlogLogger(logger ...*slog.Logger) *slog.Logger {
	var p *slog.Logger
	if len(logger) != 0 {
		p = logger[0]
	} else if p = f.ParentSlogLogger; p == nil {
		p = slog.Default()
	}

	attrs := make([]any, 0)
	for k, v := range f.properties {
		attrs = append(attrs, v.attr(k))
	}
	if len(f.counters) != 0 {
		counterAttrs := make([]slog.Attr, 0)
		for k, c := range f.counters {
			counterAttrs = append(counterAttrs, c.attr(k))
		}
		attrs = append(attrs, slog.GroupAttrs(ReservedKeyCounters, counterAttrs...))
	}

	return p.With(attrs...)
}

// NewSlogLogger creates a new child slog.Logger using [DefaultFactory.NewSlogLogger].
func NewSlogLogger(logger ...*slog.Logger) *slog.Logger {
	return DefaultFactory.NewSlogLogger(logger...)
}

// NewZerologLogger creates a new child zerolog.Logger with attributes filled out from the factory.
//
// The first given logger will be used as the parent logger. If none is given, Factory.ParentZerologLogger is used which
// itself defaults to zerolog.DefaultContextLogger.
func (f *Factory) NewZerologLogger(logger ...*zerolog.Logger) *zerolog.Logger {
	var p *zerolog.Logger
	if len(logger) != 0 {
		p = logger[0]
	} else if p = f.ParentZerologLogger; p == nil {
		p = zerolog.Ctx(context.Background())
	}

	l := p.With().Logger()
	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		for k, v := range f.properties {
			c = v.c(k, c)
		}

		if len(f.counters) != 0 {
			d := zerolog.Dict()
			for k, c := range f.counters {
				c.e(k, d)
			}
			c = c.Dict(ReservedKeyCounters, d)
		}

		return c
	})
	return &l
}

// NewZerologLogger creates a new child zerolog.Logger using [DefaultFactory.NewZerologLogger].
func NewZerologLogger(logger ...*zerolog.Logger) *zerolog.Logger {
	return DefaultFactory.NewZerologLogger(logger...)
}
