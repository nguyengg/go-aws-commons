package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// SimpleMetrics is thread-safe by use of mutex.
type SimpleMetrics struct {
	properties map[string]pvType
	counters   map[string]int64
	floaters   map[string]float64
	timings    map[string]TimingStats
	startTime  time.Time
	mu         sync.Mutex
}

var _ Metrics = &SimpleMetrics{}
var _ Metrics = (*SimpleMetrics)(nil)

// New creates an empty SimpleMetrics instance. The current UTC time will be the startTime property.
func New() Metrics {
	return NewWithStartTime(time.Now())
}

// NewWithStartTime is a variant of New that allows caller to override the startTime property.
func NewWithStartTime(startTime time.Time) Metrics {
	return &SimpleMetrics{
		properties: map[string]pvType{},
		counters: map[string]int64{
			CounterKeyFault:    0,
			CounterKeyPanicked: 0,
		},
		floaters:  map[string]float64{},
		timings:   map[string]TimingStats{},
		startTime: startTime.UTC(),
	}
}

func (m *SimpleMetrics) SetProperty(key, value string) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if reservedKeys[key] {
		return m
	}

	if m.properties == nil {
		m.properties = map[string]pvType{key: strPv{v: value}}
		return m
	}

	m.properties[key] = strPv{v: value}
	return m
}

func (m *SimpleMetrics) SetInt64Property(key string, value int64) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if reservedKeys[key] {
		return m
	}

	if m.properties == nil {
		m.properties = map[string]pvType{key: intPv{v: value}}
		return m
	}

	m.properties[key] = intPv{v: value}
	return m
}

func (m *SimpleMetrics) SetFloat64Property(key string, value float64) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if reservedKeys[key] {
		return m
	}

	if m.properties == nil {
		m.properties = map[string]pvType{key: floatPv{v: value}}
		return m
	}

	m.properties[key] = floatPv{v: value}
	return m
}

func (m *SimpleMetrics) SetJSONProperty(key string, value interface{}) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if reservedKeys[key] {
		return m
	}

	if m.properties == nil {
		m.properties = map[string]pvType{key: interPv{v: value}}
		return m
	}

	m.properties[key] = interPv{v: value}
	return m
}

func (m *SimpleMetrics) SetCount(key string, value int64, ensureExist ...string) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.counters == nil {
		m.counters = make(map[string]int64, 3+len(ensureExist))
		m.counters[CounterKeyFault] = 0
		m.counters[CounterKeyPanicked] = 0
	}

	m.counters[key] = value
	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = 0
		}
	}

	return m
}

func (m *SimpleMetrics) AddCount(key string, delta int64, ensureExist ...string) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.counters == nil {
		m.counters = make(map[string]int64, 3+len(ensureExist))
		m.counters[CounterKeyFault] = 0
		m.counters[CounterKeyPanicked] = 0
	}

	m.counters[key] += delta
	for _, k := range ensureExist {
		if _, ok := m.counters[k]; !ok {
			m.counters[k] = 0
		}
	}

	return m
}

func (m *SimpleMetrics) IncrementCount(key string) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.counters == nil {
		m.counters = map[string]int64{
			// key: 1,
			// cannot do above because the ordering is not guaranteed, i.e. if key == CounterKeyFault, the order of
			// evaluation is not specified (https://go.dev/ref/spec#Order_of_evaluation).
			CounterKeyFault:    0,
			CounterKeyPanicked: 0,
		}
	}

	m.counters[key]++
	return m
}

func (m *SimpleMetrics) Faulted() Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.counters == nil {
		m.counters = map[string]int64{
			CounterKeyFault:    1,
			CounterKeyPanicked: 0,
		}
		return m
	}

	m.counters[CounterKeyFault]++
	return m
}

func (m *SimpleMetrics) Panicked() Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.counters == nil {
		m.counters = map[string]int64{
			CounterKeyFault:    0,
			CounterKeyPanicked: 1,
		}
		return m
	}

	m.counters[CounterKeyPanicked]++
	return m
}

func (m *SimpleMetrics) SetFloat(key string, value float64, ensureExist ...string) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.floaters == nil {
		m.floaters = make(map[string]float64, 1+len(ensureExist))
	}

	m.floaters[key] = value
	for _, k := range ensureExist {
		if _, ok := m.floaters[k]; !ok {
			m.floaters[k] = 0.0
		}
	}

	return m
}

func (m *SimpleMetrics) AddFloat(key string, delta float64, ensureExist ...string) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.floaters == nil {
		m.floaters = make(map[string]float64, 1+len(ensureExist))
	}

	m.floaters[key] += delta
	for _, k := range ensureExist {
		if _, ok := m.floaters[k]; !ok {
			m.floaters[k] = 0.0
		}
	}

	return m
}

func (m *SimpleMetrics) SetTiming(key string, duration time.Duration) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.timings == nil {
		m.timings = map[string]TimingStats{key: NewTimingStats(duration)}
		return m
	}

	m.timings[key] = NewTimingStats(duration)
	return m
}

func (m *SimpleMetrics) AddTiming(key string, delta time.Duration) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.timings == nil {
		m.timings = map[string]TimingStats{key: NewTimingStats(delta)}
	}

	stats, ok := m.timings[key]
	if !ok {
		stats = NewTimingStats(delta)
		m.timings[key] = stats
		return m
	}

	stats.Add(delta)
	return m
}

var statusCodeFlags = []int{StatusCode1xx, StatusCode2xx, StatusCode3xx, StatusCode4xx, StatusCode5xx}
var statusCodeCounters = []string{"1xx", "2xx", "3xx", "4xx", "5xx"}

func (m *SimpleMetrics) SetStatusCode(statusCode int) Metrics {
	return m.SetStatusCodeWithFlag(statusCode, StatusCodeCommon)
}

func (m *SimpleMetrics) SetStatusCodeWithFlag(statusCode int, flag int) Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.properties == nil {
		m.properties = map[string]pvType{"statusCode": intPv{v: int64(statusCode)}}
	} else {
		m.properties["statusCode"] = intPv{v: int64(statusCode)}
	}

	for i, c := range statusCodeCounters {
		if statusCode/100 == i+1 {
			m.counters[c] = 1
		} else if flag&statusCodeFlags[i] != 0 {
			m.counters[c] = 0
		}
	}

	return m
}

func (m *SimpleMetrics) Log(logger *zerolog.Logger) {
	m.LogWithEndTime(logger, time.Now())
}

func (m *SimpleMetrics) LogWithEndTime(logger *zerolog.Logger, endTime time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	e := logger.Log().
		Int64(ReservedKeyStartTime, m.startTime.UnixNano()/int64(time.Millisecond)).
		Str(ReservedKeyEndTime, endTime.Format(http.TimeFormat)).
		Str(ReservedKeyTime, FormatDuration(endTime.Sub(m.startTime)))

	if len(m.properties) != 0 {
		for k, v := range m.properties {
			v.Log(k, e)
		}
	}

	if len(m.counters) != 0 {
		c := zerolog.Dict()
		for k, v := range m.counters {
			c.Int64(k, v)
		}
		e.Dict(ReservedKeyCounters, c)
	}

	if len(m.floaters) != 0 {
		c := zerolog.Dict()
		for k, v := range m.floaters {
			c.Float64(k, v)
		}
		e.Dict(ReservedKeyFloaters, c)
	}

	if len(m.timings) != 0 {
		c := zerolog.Dict()
		for k, v := range m.timings {
			c.Dict(
				k,
				zerolog.Dict().
					Str("sum", FormatDuration(v.Sum)).
					Str("min", FormatDuration(v.Min)).
					Str("max", FormatDuration(v.Max)).
					Int64("n", v.N).
					Str("avg", FormatDuration(v.Avg())))
		}
		e.Dict(ReservedKeyTimings, c)
	}

	e.Send()
}

type pvType interface {
	Log(string, *zerolog.Event)
}

type strPv struct {
	v string
}

func (t strPv) Log(key string, e *zerolog.Event) {
	e.Str(key, t.v)
}

type intPv struct {
	v int64
}

func (t intPv) Log(key string, e *zerolog.Event) {
	e.Int64(key, t.v)
}

type floatPv struct {
	v float64
}

func (t floatPv) Log(key string, e *zerolog.Event) {
	e.Float64(key, t.v)
}

type interPv struct {
	v interface{}
}

func (t interPv) Log(key string, e *zerolog.Event) {
	e.Interface(key, t.v)
}
