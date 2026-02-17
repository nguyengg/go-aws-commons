package metrics

import (
	"log/slog"
	"sync"
	"time"
)

// Metrics contains metrics about a single request that are to be logged with slog.
//
// Accessing the struct fields directly is not concurrency-safe. Use the various Set/Add methods if you need to modify
// the instance from multiple goroutines. The zero-value instance is ready for use by those methods as well.
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

	properties map[string]slog.Value
	counters   map[string]slog.Value
	timings    map[string]TimingStats

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
	// Formatted as epoch millisecond for machine parsing.
	ReservedKeyStartTime = "startTime"
	// ReservedKeyEndTime is the top-level property for the end time.
	//
	// Formatted as http.TimeFormat for human readability.
	ReservedKeyEndTime = "endTime"
	// ReservedKeyLatency is the top-level property measuring duration between Metrics.Start and Metrics.End.
	//
	// If smaller than 1 second, the latency is formatted at millisecond unit (e.g. 153.25ms). Otherwise, it is
	// formatted at second unit (e.g. 325.12s).
	ReservedKeyLatency = "latency"
	// ReservedKeyCounters is the top-level property containing int64-based and float64-based metrics.
	ReservedKeyCounters = "counters"
	// ReservedKeyTimings is the top-level property containing timing-based metrics (TimingStats).
	ReservedKeyTimings = "timings"
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
			m.properties = make(map[string]slog.Value)
		}

		if m.counters == nil {
			m.counters = make(map[string]slog.Value)
		}

		if m.timings == nil {
			m.timings = make(map[string]TimingStats)
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

	m.properties[key] = slog.StringValue(value)

	return m
}

// AnyValue is a variant of String that accepts slog.Value as the value.
//
// Returns self for chaining.
func (m *Metrics) AnyValue(key string, value slog.Value) *Metrics {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()

	if reservedKeys[key] {
		return m
	}

	m.properties[key] = value

	return m
}

// Any is a variant of String that accepts any value wrapped as slog.AnyValue.
//
// Returns self for chaining.
func (m *Metrics) Any(key string, value any) *Metrics {
	return m.AnyValue(key, slog.AnyValue(value))
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

	m.counters[key] = slog.Int64Value(value)

	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = slog.Int64Value(0)
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

	if v, ok := m.counters[key]; ok {
		switch k := v.Kind(); k {
		case slog.KindInt64:
			m.counters[key] = slog.Int64Value(v.Int64() + delta)
		case slog.KindFloat64:
			m.counters[key] = slog.Float64Value(v.Float64() + float64(delta))
		default:
			panic("unexpected counter type " + k.String())
		}
	} else {
		m.counters[key] = slog.Int64Value(delta)
	}

	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = slog.Int64Value(0)
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

	m.counters[key] = slog.Float64Value(value)

	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = slog.Float64Value(0)
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

	if v, ok := m.counters[key]; ok {
		switch k := v.Kind(); k {
		case slog.KindInt64:
			m.counters[key] = slog.Float64Value(float64(v.Int64()) + delta)
		case slog.KindFloat64:
			m.counters[key] = slog.Float64Value(v.Float64() + delta)
		default:
			panic("unexpected counter type " + k.String())
		}
	} else {
		m.counters[key] = slog.Float64Value(delta)
	}

	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = slog.Int64Value(0)
		}
	}

	return m
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
		stats.Add(latency)
	} else {
		m.timings[key] = NewTimingStats(latency)
	}

	return m
}

// Attrs sets the Metrics.End (if not set) and returns the attributes to be logged with slog.
func (m *Metrics) Attrs() (attrs []slog.Attr) {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.End.IsZero() {
		m.End = time.Now()
	}

	attrs = []slog.Attr{
		slog.Any(ReservedKeyStartTime, startTimeValue{m.Start}),
		slog.Any(ReservedKeyEndTime, endTimeValue{m.End}),
		slog.Any(ReservedKeyLatency, durationValue{m.End.Sub(m.Start)}),
	}

	if len(m.properties) != 0 {
		for k, v := range m.properties {
			attrs = append(attrs, slog.Any(k, v))
		}
	}

	if len(m.counters) != 0 {
		counterAttrs := make([]slog.Attr, 0, len(m.counters))
		for k, v := range m.counters {
			counterAttrs = append(counterAttrs, slog.Any(k, v))
		}

		attrs = append(attrs, slog.GroupAttrs(ReservedKeyCounters, counterAttrs...))
	}

	if len(m.timings) != 0 {
		for k, v := range m.timings {
			attrs = append(attrs, slog.Any(k, v))
		}
	}

	return
}

func (m *Metrics) LogValue() slog.Value {
	return slog.GroupValue(m.Attrs()...)
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
