package metrics

import (
	"strconv"
	"sync"
	"time"
)

// Metrics contains metrics about a single request or operation.
//
// Accessing the struct fields directly is not concurrency-safe. Use the various Set/Add methods if you need to modify
// the instance from multiple goroutines. The zero-value instance is ready for use by those methods as well.
//
// The Metrics instance can be logged in several ways.
//
//	// this will log the metrics as top-level JSON fields.
//	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
//	slog.LogAttrs(context.Background(), slog.LevelInfo, "request done", m.Attrs()...)
//
//	// zerolog can also be used.
//	m.Zerolog(zerolog.New(os.Stderr).Info().Timestamp()).Send()
//
//	// finally, the instance can also be marshalled as JSON.
type Metrics struct {
	// Start is the start time of the Metrics instance.
	//
	// This value is useful for computing the latency ("time") of a singular request. If not overridden (zero-value),
	// any call to modify the Metrics instance will set Start to time.Now.
	Start time.Time

	// End is the end time of the Metrics instance.
	//
	// This value is useful for computing the latency ("time") of a singular request. If not overridden (zero-value),
	// time.Now will be used when logging the Metrics instance.
	End time.Time

	// RawFormatting, if true, will not apply special formatting to Start, End, and latency metrics.
	RawFormatting bool

	properties map[string]property
	counters   map[string]counter
	timings    map[string]latencyStats

	mu   sync.Mutex
	once sync.Once
}

// New creates a new Metrics instance with Metrics.Start set to time.Now and all struct fields populated.
func New() *Metrics {
	m := &Metrics{}
	m.init()
	return m
}

// NewWithStart is a variant of New with the specified Metrics.Start.
func NewWithStart(start time.Time) *Metrics {
	m := &Metrics{Start: start}
	m.init()
	return m
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
	// ReservedKeyLatency is the top-level property measuring duration between Metrics.Start and Metrics.End.
	//
	// Formatted using FormatDuration; to use native formatter, use Metrics.RawFormatting.
	ReservedKeyLatency = "latency"
	// ReservedKeyCounters is the top-level property containing int64-based and float64-based metrics.
	ReservedKeyCounters = "counters"
	// ReservedKeyTimings is the top-level property containing timing-based metrics.
	//
	// Formatted using FormatDuration; to use native formatter, use Metrics.RawFormatting.
	ReservedKeyTimings = "timings"

	// CounterKeyFault is a special counter metrics that indicates the request or operation ends with an error.
	CounterKeyFault = "fault"
	// CounterKeyPanicked is a special counter metrics that indicates handling the request or operation ends with a
	// panic which was recovered.
	CounterKeyPanicked = "panicked"
)

var reservedKeys = map[string]bool{
	ReservedKeyStartTime: true,
	ReservedKeyEndTime:   true,
	ReservedKeyLatency:   true,
	ReservedKeyCounters:  true,
	ReservedKeyTimings:   true,
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
		}

		if m.timings == nil {
			m.timings = map[string]latencyStats{}
		}
	})
}

// String creates or modifies a string key-value property pair.
//
// Properties are top-level fields in the JSON log message. If the property is reserved, the method no-ops.
//
// If called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) String(key, value string) *Metrics {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = property{stringKind, value}

	return m
}

// Int64 creates or modifies an int64 key-value property pair.
//
// Properties are top-level fields in the JSON log message. If the property is reserved, the method no-ops.
//
// If called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) Int64(key string, value int64) *Metrics {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = property{int64Kind, value}

	return m
}

// Float64 creates or modifies a flaot64 key-value property pair.
//
// Properties are top-level fields in the JSON log message. If the property is reserved, the method no-ops.
//
// If called multiples on the same key, the last one wins.
//
// Returns self for chaining.
func (m *Metrics) Float64(key string, value float64) *Metrics {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = property{float64Kind, value}

	return m
}

// Any is a variant of String that accepts any value instead.
//
// Returns self for chaining.
func (m *Metrics) Any(key string, value any) *Metrics {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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

// Faulted is a convenient method to increase the CounterKeyFault counter by 1.
//
// Returns self for chaining.
func (m *Metrics) Faulted() *Metrics {
	return m.AddCounter(CounterKeyFault, 1)
}

// Panicked is a convenient method to increase the CounterKeyPanicked counter by 1.
//
// Returns self for chaining.
func (m *Metrics) Panicked() *Metrics {
	return m.AddCounter(CounterKeyPanicked, 1)
}

// AddTiming adds the latency time.Duration to aggregated dataset.
//
// The statistics will show up under ReservedKeyTimings top-level property.
//
// Returns self for chaining.
func (m *Metrics) AddTiming(key string, latency time.Duration) *Metrics {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()

	if stats, ok := m.timings[key]; ok {
		stats.add(latency)
	} else {
		m.timings[key] = newDurationStats(latency)
	}

	return m
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

// FormatDuration formats the given time.Duration as seconds or milliseconds, truncating it to the next thousandth unit
// (retaining at most 3 decimal points).
func FormatDuration(d time.Duration) string {
	if d >= 1*time.Second {
		return strconv.FormatFloat(d.Truncate(time.Millisecond).Seconds(), 'f', -1, 64) + "s"
	}

	return strconv.FormatFloat(float64(d.Truncate(time.Microsecond))/float64(time.Millisecond), 'f', -1, 64) + "ms"
}
