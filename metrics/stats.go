package metrics

import (
	"log/slog"
	"time"
)

// TimingStats aggregates one or more time.Duration metrics.
type TimingStats struct {
	sum time.Duration
	min time.Duration
	max time.Duration
	n   int64
}

// NewTimingStats creates a new TimingStats instance.
func NewTimingStats(duration time.Duration) TimingStats {
	return TimingStats{
		sum: duration,
		min: duration,
		max: duration,
		n:   1,
	}
}

// Add adds the specified time.Duration to the dataset.
func (s *TimingStats) Add(duration time.Duration) *TimingStats {
	s.sum += duration
	if s.min > duration {
		s.min = duration
	}
	if s.max < duration {
		s.max = duration
	}
	s.n++
	return s
}

// Sum returns the total of all time.Duration.
func (s *TimingStats) Sum() time.Duration {
	return s.sum
}

// Min returns the lowest time.Duration value.
func (s *TimingStats) Min() time.Duration {
	return s.min
}

// Max returns the highest time.Duration value.
func (s *TimingStats) Max() time.Duration {
	return s.max
}

// Avg returns the average time.Duration value.
func (s *TimingStats) Avg() time.Duration {
	return s.sum / time.Duration(s.n)
}

func (s *TimingStats) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("sum", durationValue{s.sum}),
		slog.Any("min", durationValue{s.min}),
		slog.Any("max", durationValue{s.max}),
		slog.Any("avg", durationValue{s.Avg()}),
		slog.Int64("n", s.n),
	)
}
