package metrics

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rotisserie/eris"
)

// Metrics contains metrics about a single request or operation.
//
// Unlike structured logging which may emit several messages with the same key-value pairs, Metrics instances should be
// logged only once by way of Close, after a request or operation has finished. For example, a Metrics instance may
// measure a GET request from start to end time, capturing whether the response's status is a 2xx or 5xx, etc. A Metrics
// instance may also measure processing a long-running task such as handling SQS messages, etc.
//
// The zero-value instance is ready for use as well; the first call to modify the Metrics instance will set Start to
// time.Now if Start is zero value. Metrics is not safe for concurrent use.
//
// The Metrics instance can be logged in several ways. By default, Close will write a JSON entry to os.Stderr. To log
// with slog or zerolog, pass SlogMetricsLogger or ZerologMetricsLogger into the Factory that is used to create the
// Metrics instance.
type Metrics struct {
	// Start is the start time of the Metrics instance.
	//
	// This value is useful for computing the duration of a singular request. If not overridden (zero-value), any call
	// to modify the Metrics instance will set Start to time.Now.
	Start time.Time

	// End is the end time of the Metrics instance.
	//
	// This value is useful for computing the duration of a singular request. If not overridden (zero-value), time.Now
	// will be used when logging the Metrics instance.
	End time.Time

	properties map[string]property
	counters   map[string]counter
	timings    map[string]durationStats
	errors     errorStack

	logger Logger
	once   sync.Once
}

// Logger controls how the Metrics instance is actually logged when Metrics.Close is called.
type Logger interface {
	Log(ctx context.Context, m *Metrics) error
}

// New creates a new Metrics instance with Metrics.Start set to time.Now.
//
// By default, a single JSON entry is logged to os.Stderr (JSONLogger). There's support for SlogMetricsLogger and ZerologMetricsLogger
// out of the box.
//
// New uses DefaultFactory to create the Metrics instance.
func New(optFns ...func(m *Metrics)) *Metrics {
	return DefaultFactory.New(optFns...)
}

// Close will log the Metrics instance to the channel specified at init.
//
// If End is zero value, End will be set to time.Now.
func (m *Metrics) Close() error {
	m.init()

	if m.End.IsZero() {
		m.End = time.Now()
	}

	return m.logger.Log(context.Background(), m)
}

// CloseContext is a variant of Close that accepts a context.
//
// Useful if using slog or zerolog that can retrieve a logger from context.
func (m *Metrics) CloseContext(ctx context.Context) error {
	m.init()

	if m.End.IsZero() {
		m.End = time.Now()
	}

	return m.logger.Log(ctx, m)
}

// Reserved property keys.
const (
	// ReservedKeyStartTime is the top-level property for the start time.
	//
	// Formatted as epoch millisecond for machine parsing; to use native formatter, use Metrics.RawFormatting.
	ReservedKeyStartTime = "startTime"
	// ReservedKeyEndTime is the top-level property for the end time.
	//
	// Formatted as http.TimeFormat for human readability; to use native formatter, use Metrics.RawFormatting.
	ReservedKeyEndTime = "endTime"
	// ReservedKeyDuration is the top-level property measuring duration between Metrics.Start and Metrics.End.
	//
	// Formatted using FormatDuration; to use native formatter, use Metrics.RawFormatting.
	ReservedKeyDuration = "duration"
	// ReservedKeyCounters is the top-level property containing int64-based and float64-based metrics.
	ReservedKeyCounters = "counters"
	// ReservedKeyTimings is the top-level property containing timing-based metrics.
	//
	// Formatted using FormatDuration; to use native formatter, use Metrics.RawFormatting.
	ReservedKeyTimings = "timings"
	// ReservedKeyError is the top-level property containing the error that causes your program to stop, and is often
	// returned by the entry point method.
	//
	// This property can only be set using Metrics.Error, and will automatically set "fault" (CounterKeyFault) counter
	// to 1.
	//
	// Each error will be formatted using eris.ToJSON.
	ReservedKeyError = "error"
	// ReservedKeyErrors is the top-level property containing errors and their stack traces.
	//
	// This property is a stack and can only be modified with Metrics.PushError. Because it's a stack, the log will show
	// them in reverse order: later errors will show up before earlier errors.
	//
	// Each error will be formatted using eris.ToJSON.
	ReservedKeyErrors = "errors"

	// CounterKeyFault is a special counter metrics that indicates the request or operation ends with an error.
	CounterKeyFault = "fault"
	// CounterKeyPanicked is a special counter metrics that indicates handling the request or operation ends with a
	// panic which was recovered.
	CounterKeyPanicked = "panicked"
)

var reservedKeys = map[string]bool{
	ReservedKeyStartTime: true,
	ReservedKeyEndTime:   true,
	ReservedKeyDuration:  true,
	ReservedKeyCounters:  true,
	ReservedKeyTimings:   true,
	ReservedKeyError:     true,
	ReservedKeyErrors:    true,
}

func (m *Metrics) init() {
	m.once.Do(func() {
		if m.Start.IsZero() {
			m.Start = time.Now()
		}

		if m.properties == nil {
			m.properties = map[string]property{}
		}

		if m.counters == nil {
			m.counters = map[string]counter{
				CounterKeyFault:    {int64Kind, int64(0)},
				CounterKeyPanicked: {int64Kind, int64(0)},
			}
		} else {
			if _, ok := m.counters[CounterKeyFault]; !ok {
				m.counters[CounterKeyFault] = counter{int64Kind, int64(0)}
			}
			if _, ok := m.counters[CounterKeyPanicked]; !ok {
				m.counters[CounterKeyPanicked] = counter{int64Kind, int64(0)}
			}
		}

		if m.timings == nil {
			m.timings = map[string]durationStats{}
		}

		if m.errors == nil {
			m.errors = errorStack{}
		}

		if m.logger == nil {
			m.logger = &JSONLogger{os.Stderr}
		}
	})
}

// String sets a string property.
//
// Properties are top-level fields in the JSON log message. If the property's key is reserved, the method no-ops. If
// called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) String(key, value string) *Metrics {
	m.init()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = property{stringKind, value}

	return m
}

// Int64 sets a property whose value is type int64.
//
// Properties are top-level fields in the JSON log message. If the property's key is reserved, the method no-ops. If
// called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) Int64(key string, value int64) *Metrics {
	m.init()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = property{int64Kind, value}

	return m
}

// Float64 sets a property whose value is type float64.
//
// Properties are top-level fields in the JSON log message. If the property's key is reserved, the method no-ops. If
// called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) Float64(key string, value float64) *Metrics {
	m.init()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = property{float64Kind, value}

	return m
}

// Error sets the property "error" (ReservedKeyError).
//
// The error will be wrapped (eris.Wrap) and formatted as JSON (eris.ToJSON with trace capture) at logging. If you don't
// want this behaviour, use Any. You can also mimic this behaviour but for other properties with:
//
//	m.Any("myError", eris.ToJSON(eris.Wrap(myError, "message"), true))
//
// Calling Error will automatically set "fault" (CounterKeyFault) counter to 1. This should be given the error that is
// returned by your entry point, as opposed to PushError which can be given any errors you feel like should be logged.
//
// Properties are top-level fields in the JSON log message. If the property's key is reserved, the method no-ops. If
// called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) Error(value error) *Metrics {
	m.init()

	m.properties[ReservedKeyError] = property{errorKind, eris.Wrap(value, value.Error())}
	m.counters[CounterKeyFault] = counter{int64Kind, int64(1)}

	return m
}

// HasError returns true only if Error has been called before.
//
// Some framework can automatically call Error, but to avoid overriding a custom error that user has set, use this.
func (m *Metrics) HasError() bool {
	m.init()

	_, ok := m.properties[ReservedKeyError]
	return ok
}

// AnError is a variant of Error that allows arbitrary key (except for reserved keys).
//
// Properties are top-level fields in the JSON log message. If the property's key is reserved, the method no-ops. If
// called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) AnError(key string, value error) *Metrics {
	m.init()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = property{errorKind, eris.Wrap(value, value.Error())}

	return m
}

// Any sets a property whose value can have any type.
//
// The value should implement json.Marshaler if logging using JSON; [slog.Valuer] if logging with slog.
//
// Properties are top-level fields in the JSON log message. If the property's key is reserved, the method no-ops. If
// called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) Any(key string, value any) *Metrics {
	m.init()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = property{anyKind, value}

	return m
}

// SetCounter sets the Metrics.Counters mapping with the specified key to the given value.
//
// Additional names can be given to ensure they exist with the initial value (0) unless they've already been set. The
// values will show up under ReservedKeyCounters top-level property.
//
// Returns self for chaining.
func (m *Metrics) SetCounter(key string, value int64, ensureExist ...string) *Metrics {
	m.init()

	m.counters[key] = counter{int64Kind, value}

	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = counter{int64Kind, int64(0)}
		}
	}

	return m
}

// AddCounter increases the Metrics.Counters mapping with the specified key by the given delta.
//
// Additional names can be given to ensure they exist with the initial value (0) unless they've already been set. The
// values will show up under ReservedKeyCounters top-level property.
//
// Returns self for chaining.
func (m *Metrics) AddCounter(key string, delta int64, ensureExist ...string) *Metrics {
	m.init()

	if c, ok := m.counters[key]; ok {
		c.addInt64(delta)
	} else {
		m.counters[key] = counter{int64Kind, delta}
	}

	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = counter{int64Kind, int64(0)}
		}
	}

	return m
}

// SetFloater sets the Metrics.Floaters mapping with the specified key to the given value.
//
// Additional names can be given to ensure they exist with the initial value (0) unless they've already been set. The
// values will show up under ReservedKeyCounters top-level property.
//
// Returns self for chaining.
func (m *Metrics) SetFloater(key string, value float64, ensureExist ...string) *Metrics {
	m.init()

	m.counters[key] = counter{float64Kind, value}

	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = counter{float64Kind, int64(0)}
		}
	}

	return m
}

// AddFloater increases the Metrics.Floaters mapping with the specified key by the given delta.
//
// Additional names can be given to ensure they exist with the initial value (0) unless they've already been set. The
// values will show up under ReservedKeyCounters top-level property.
//
// Returns self for chaining.
func (m *Metrics) AddFloater(key string, delta float64, ensureExist ...string) *Metrics {
	m.init()

	if c, ok := m.counters[key]; ok {
		c.addFloat64(delta)
	} else {
		m.counters[key] = counter{float64Kind, delta}
	}

	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = counter{float64Kind, int64(0)}
		}
	}

	return m
}

// Faulted is a convenient method to set the CounterKeyFault counter to 1.
//
// Returns self for chaining.
func (m *Metrics) Faulted() *Metrics {
	m.init()

	m.counters[CounterKeyFault] = counter{int64Kind, int64(1)}

	return m
}

// Panicked is a convenient method to set the CounterKeyPanicked counter .
//
// Returns self for chaining.
func (m *Metrics) Panicked() *Metrics {
	m.init()

	m.counters[CounterKeyPanicked] = counter{int64Kind, int64(1)}

	return m
}

// AddTiming adds the latency time.Duration to aggregated dataset.
//
// The statistics will show up under ReservedKeyTimings top-level property.
//
// Returns self for chaining.
func (m *Metrics) AddTiming(key string, latency time.Duration) *Metrics {
	m.init()

	if stats, ok := m.timings[key]; ok {
		stats.add(latency)
	} else {
		m.timings[key] = newDurationStats(latency)
	}

	return m
}

// PushError pushes the given error to the "errors" (ReservedKeyErrors) error stack.
//
// PushError does not automatically set "fault" (CounterKeyFault) counter. It can be used to log arbitrary errors that
// your program encounter but were able to handle. If there was an error that caused your program to stop, use Error
// instead, or in addition to PushError.
//
// Returns self for chaining.
func (m *Metrics) PushError(err error, withTrace bool) *Metrics {
	m.init()

	m.errors.push(err, withTrace)

	return m
}

// FormatDuration formats the given time.Duration as seconds or milliseconds, truncating it to the next thousandth unit
// (retaining at most 3 decimal points).
func FormatDuration(d time.Duration) string {
	if d >= 1*time.Second {
		return strconv.FormatFloat(d.Truncate(time.Millisecond).Seconds(), 'f', -1, 64) + "s"
	}

	return strconv.FormatFloat(float64(d.Truncate(time.Microsecond))/float64(time.Millisecond), 'f', -1, 64) + "ms"
}
