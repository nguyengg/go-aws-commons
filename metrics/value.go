package metrics

import (
	"time"
)

// kind is the type of values that can be logged with Metrics.
//
// Inspired by slog.Value, only string, interface, int64, or float64 are supported.
type kind int

const (
	stringKind kind = iota
	anyKind
	int64Kind
	float64Kind
)

// property is either string or interface (any).
type property struct {
	t kind
	v any
}

// counter is either int64 or float 64.
type counter struct {
	t kind
	v any
}

func (c *counter) addInt64(v int64) {
	switch c.t {
	case int64Kind:
		c.v = c.v.(int64) + v
	case float64Kind:
		// if type is float64 then retain the float64 type.
		c.v = c.v.(float64) + float64(v)
	default:
		panic("invalid counter type")
	}
}

func (c *counter) addFloat64(v float64) {
	switch c.t {
	case int64Kind:
		// if type was int64 then convert to float64.
		c.v = float64(c.v.(int64)) + v
	case float64Kind:
		c.v = c.v.(float64) + v
	default:
		panic("invalid counter type")
	}
}

func (c *counter) isPositive() bool {
	switch c.t {
	case int64Kind:
		return c.v.(int64) > 0
	case float64Kind:
		return c.v.(float64) > 0
	default:
		panic("invalid counter type")
	}
}

// latencyStats aggregates one or more time.Duration metrics.
type latencyStats struct {
	sum time.Duration
	min time.Duration
	max time.Duration
	n   int64
}

// newDurationStats creates a new latencyStats instance.
func newDurationStats(duration time.Duration) latencyStats {
	return latencyStats{
		sum: duration,
		min: duration,
		max: duration,
		n:   1,
	}
}

// add adds the specified time.Duration to the dataset.
func (s *latencyStats) add(d time.Duration) *latencyStats {
	s.sum += d

	if s.min > d {
		s.min = d
	}

	if s.max < d {
		s.max = d
	}

	s.n++

	return s
}

func (s *latencyStats) avg() time.Duration {
	return s.sum / time.Duration(s.n)
}
