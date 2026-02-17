package metrics

import (
	"log/slog"
	"time"
)

// Attrs sets the Metrics.End (if not set) and returns the attributes to be logged with slog.
func (m *Metrics) Attrs() (attrs []slog.Attr) {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.End.IsZero() {
		m.End = time.Now()
	}

	if m.RawFormatting {
		attrs = []slog.Attr{
			slog.Time(ReservedKeyStartTime, m.Start),
			slog.Time(ReservedKeyEndTime, m.End),
			slog.Duration(ReservedKeyLatency, m.End.Sub(m.Start)),
		}
	} else {
		attrs = []slog.Attr{
			slog.Int64(ReservedKeyStartTime, m.Start.UnixMilli()),
			slog.String(ReservedKeyEndTime, m.End.UTC().Format(time.RFC1123)),
			slog.String(ReservedKeyLatency, FormatDuration(m.End.Sub(m.Start))),
		}
	}

	for k, p := range m.properties {
		attrs = append(attrs, p.attr(k))
	}

	if len(m.counters) != 0 {
		counterAttrs := make([]slog.Attr, 0, len(m.counters))
		for k, c := range m.counters {
			counterAttrs = append(counterAttrs, c.attr(k))
		}

		attrs = append(attrs, slog.GroupAttrs(ReservedKeyCounters, counterAttrs...))
	}

	if len(m.timings) != 0 {
		timingAttrs := make([]slog.Attr, 0, len(m.timings))
		if m.RawFormatting {
			for k, t := range m.timings {
				timingAttrs = append(timingAttrs, slog.GroupAttrs(k,
					slog.Duration("sum", t.sum),
					slog.Duration("min", t.min),
					slog.Duration("max", t.max),
					slog.Duration("avg", t.avg()),
					slog.Int64("n", t.n),
				))
			}
		} else {
			for k, v := range m.timings {
				timingAttrs = append(timingAttrs, slog.GroupAttrs(k,
					slog.String("sum", FormatDuration(v.sum)),
					slog.String("min", FormatDuration(v.min)),
					slog.String("max", FormatDuration(v.max)),
					slog.String("avg", FormatDuration(v.avg())),
					slog.Int64("n", v.n),
				))
			}
		}

		attrs = append(attrs, slog.GroupAttrs(ReservedKeyTimings, timingAttrs...))
	}

	return
}

func (m *Metrics) LogValue() slog.Value {
	return slog.GroupValue(m.Attrs()...)
}

func (p *property) attr(key string) slog.Attr {
	switch p.t {
	case stringKind:
		return slog.String(key, p.v.(string))
	case anyKind:
		return slog.Any(key, p.v)
	default:
		panic("invalid property type")
	}
}
func (c *counter) attr(key string) slog.Attr {
	switch c.t {
	case int64Kind:
		return slog.Int64(key, c.v.(int64))
	case float64Kind:
		return slog.Float64(key, c.v.(float64))
	default:
		panic("invalid counter type")
	}
}
