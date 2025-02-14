package metrics

import "time"

type TimingStats struct {
	Sum time.Duration
	Min time.Duration
	Max time.Duration
	N   int64
}

func NewTimingStats(duration time.Duration) TimingStats {
	return TimingStats{
		Sum: duration,
		Min: duration,
		Max: duration,
		N:   1,
	}
}

func (s *TimingStats) Add(duration time.Duration) *TimingStats {
	s.Sum += duration
	if s.Min > duration {
		s.Min = duration
	}
	if s.Max < duration {
		s.Max = duration
	}
	s.N++
	return s
}

func (s *TimingStats) Avg() time.Duration {
	return s.Sum / time.Duration(s.N)
}
