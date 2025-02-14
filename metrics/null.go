package metrics

import (
	"time"

	"github.com/rs/zerolog"
)

// NullMetrics no-ops on all implementations.
type NullMetrics struct {
}

var _ Metrics = &NullMetrics{}
var _ Metrics = (*NullMetrics)(nil)

func (m *NullMetrics) SetProperty(string, string) Metrics {
	return m
}

func (m *NullMetrics) SetInt64Property(string, int64) Metrics {
	return m
}

func (m *NullMetrics) SetFloat64Property(string, float64) Metrics {
	return m
}

func (m *NullMetrics) SetJSONProperty(string, interface{}) Metrics {
	return m
}

func (m *NullMetrics) SetCount(string, int64, ...string) Metrics {
	return m
}

func (m *NullMetrics) AddCount(string, int64, ...string) Metrics {
	return m
}

func (m *NullMetrics) IncrementCount(string) Metrics {
	return m
}

func (m *NullMetrics) Faulted() Metrics {
	return m
}

func (m *NullMetrics) Panicked() Metrics {
	return m
}

func (m *NullMetrics) SetFloat(string, float64, ...string) Metrics {
	return m
}

func (m *NullMetrics) AddFloat(string, float64, ...string) Metrics {
	return m
}

func (m *NullMetrics) SetTiming(string, time.Duration) Metrics {
	return m
}

func (m *NullMetrics) AddTiming(string, time.Duration) Metrics {
	return m
}

func (m *NullMetrics) SetStatusCode(int) Metrics {
	return m
}

func (m *NullMetrics) SetStatusCodeWithFlag(int, int) Metrics {
	return m
}

func (m *NullMetrics) Log(logger *zerolog.Logger) {
	logger.Log().Int("nullMetrics", 1).Send()
}

func (m *NullMetrics) LogWithEndTime(logger *zerolog.Logger, _ time.Time) {
	logger.Log().Int("nullMetrics", 1).Send()
}
