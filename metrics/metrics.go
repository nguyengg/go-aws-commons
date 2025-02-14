package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// Metrics contains metrics that can be logged as structured JSON using zerolog.
type Metrics interface {
	// SetProperty creates or modifies a string key-value property pair.
	//
	// Properties are top-level fields in the JSON log message. A few properties are generated by the metrics instance
	// and cannot be overridden:
	//   - startTime: epoch millisecond for machine parsing
	//   - endTime: http.TimeFormat for human readability
	//   - time: latency from metrics' creation to when it's logged - in milliseconds
	//   - timings: generic latencies - in milliseconds
	//   - counters (int64) and floaters (float64)
	//
	// If called multiples on the same key, the last one wins.
	//
	// Returns self for chaining.
	SetProperty(key, value string) Metrics
	// SetInt64Property is a variant of SetProperty for int64 values.
	SetInt64Property(key string, value int64) Metrics
	// SetFloat64Property is a variant of SetProperty for float64 values.
	SetFloat64Property(key string, value float64) Metrics
	// SetJSONProperty is a variant of SetProperty for values that don't fit the other variants.
	// If the value fails to be serialised, the error message will be used in its stead.
	// See zerolog.Event's Interface method.
	SetJSONProperty(key string, v interface{}) Metrics
	// SetCount sets the integer counter of the specified key to a value.
	// Additional names can be given to ensure they exist with the initial value (0) unless they've already been set.
	// Returns self for chaining.
	SetCount(key string, value int64, ensureExist ...string) Metrics

	// AddCount increases the integer counter of the specified key by a delta.
	// Additional names can be given to ensure they exist with the initial value (0) unless they've already been set.
	// Use IncrementCount if you need to increase by 1.
	// Returns self for chaining.
	AddCount(key string, delta int64, ensureExist ...string) Metrics
	// IncrementCount is a convenient method to increase the integer counter of the specified key by 1.
	// Returns self for chaining.
	IncrementCount(key string) Metrics
	// Faulted is a convenient method to increase the integer counter named CounterKeyFault by 1.
	// Returns self for chaining.
	Faulted() Metrics
	// Panicked is a convenient method to increase the integer counter named CounterKeyPanicked by 1.
	// Returns self for chaining.
	Panicked() Metrics

	// SetFloat sets the float counter of the specified key to a value.
	// Additional names can be given to ensure they exist with the initial value (0) unless they've already been set.
	// Returns self for chaining.
	SetFloat(key string, value float64, ensureExist ...string) Metrics
	// AddFloat increases the float counter of the specified key by a delta.
	// Additional names can be given to ensure they exist with the initial value (0) unless they've already been set.
	// Returns self for chaining.
	AddFloat(key string, delta float64, ensureExist ...string) Metrics

	// SetTiming sets a timing metric of the specified key to a duration.
	// Returns self for chaining.
	SetTiming(key string, duration time.Duration) Metrics
	// AddTiming increases a timing metric of the specified key by a delta.
	// Returns self for chaining.
	AddTiming(key string, delta time.Duration) Metrics

	// SetStatusCode sets a "statusCode" property and start additional status code counters (StatusCodeCommon).
	// Use SetStatusCodeWithFlags to customise which status code counter is not printed.
	// Returns self for chaining.
	SetStatusCode(statusCode int) Metrics
	// SetStatusCodeWithFlag emits a "statusCode" property and additional counters controlled by flag.
	// If flag is 0, only the yxx counter matching the given status code is set to 1. Otherwise, whichever
	// status code flag is specified (StatusCode1xx, StatusCode3xx, StatusCodeCommon, StatusCodeAll, etc.) get a
	// 0-value metric.
	// Returns self for chaining.
	SetStatusCodeWithFlag(statusCode int, flag int) Metrics

	// Log uses the given logger to write the metrics contents.
	Log(*zerolog.Logger)
	// LogWithEndTime is a variant of Log that receives an explicit end time.
	LogWithEndTime(*zerolog.Logger, time.Time)
}

// WithClientSideMetrics counter metrics that are always emitted.
const (
	CounterKeyFault    = "fault"
	CounterKeyPanicked = "panicked"
)

// Reserved property keys.
const (
	ReservedKeyStartTime = "startTime"
	ReservedKeyEndTime   = "endTime"
	ReservedKeyTime      = "time"
	ReservedKeyCounters  = "counters"
	ReservedKeyFloaters  = "floaters"
	ReservedKeyTimings   = "timings"
)

var reservedKeys = map[string]bool{
	ReservedKeyStartTime: true,
	ReservedKeyEndTime:   true,
	ReservedKeyTime:      true,
	ReservedKeyCounters:  true,
	ReservedKeyFloaters:  true,
	ReservedKeyTimings:   true,
}

const (
	StatusCode1xx = 1 << iota
	StatusCode2xx
	StatusCode3xx
	StatusCode4xx
	StatusCode5xx
	StatusCodeCommon = StatusCode2xx | StatusCode4xx | StatusCode5xx
	StatusCodeAll    = StatusCode1xx | StatusCode2xx | StatusCode3xx | StatusCode4xx | StatusCode5xx
)

type metricsKey struct{}

// WithContext returns a new child context with the given metrics attached.
//
// The metrics can be retrieved using Ctx(child) or TryCtx(child).
func WithContext(ctx context.Context, m Metrics) context.Context {
	return context.WithValue(ctx, metricsKey{}, m)
}

// Ctx returns the Metrics instance from the specified context if available.
//
// If not, a NullMetrics instance will be used.
func Ctx(ctx context.Context) Metrics {
	if m, ok := ctx.Value(metricsKey{}).(Metrics); ok && m != nil {
		return m
	}

	return &NullMetrics{}
}

// TryCtx is a variant of Ctx that does not return NullMetrics.
//
// Use this if you absolutely need an existing Metrics instance to exist.
func TryCtx(ctx context.Context) (Metrics, bool) {
	m, ok := ctx.Value(metricsKey{}).(Metrics)
	return m, ok
}

// FormatDuration formats the duration in layout 12.345ms.
func FormatDuration(duration time.Duration) string {
	return strconv.FormatFloat(float64(duration.Nanoseconds())/float64(time.Millisecond), 'f', 3, 64) + " ms"
}
