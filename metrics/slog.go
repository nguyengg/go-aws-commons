package metrics

import (
	"context"
	"log/slog"
	"time"
)

// SlogMetricsLogger implements Logger using slog.Logger.
type SlogMetricsLogger struct {
	// Logger is the slog.Logger instance to use.
	//
	// Default to slog.Default.
	Logger *slog.Logger

	// Level is the log level to use.
	//
	// By default, the log level is dynamic. If the Metrics instance indicates an error state with non-zero fault
	// and/or panicked counter, slog.LevelError is used. Otherwise, slog.LevelInfo is used.
	Level *slog.Level

	// Msg is the message to be used with the log.
	Msg string

	// Group is the name of the slog.Group to place the Metrics content.
	//
	// By default, the Metrics instance is logged as top-level fields (no group).
	//
	// Equivalent to ZerologMetricsLogger.Dict.
	Group string

	// NoCustomFormatter disables specialised formatting for timestamps and durations.
	//
	// By default, to match LogJSON, the start time is logged as epoch millisecond, the end time using RFC1123, and
	// all time.Duration using FormatDuration. To disable this behaviour and rely on slog's own formatters, set
	// NoCustomFormatter to true.
	NoCustomFormatter bool
}

func (l *SlogMetricsLogger) Log(ctx context.Context, m *Metrics) error {
	attrs := m.attrs(l.NoCustomFormatter)

	var logLevel = slog.LevelInfo
	if l.Level != nil {
		logLevel = *l.Level
	} else if c := m.counters[CounterKeyFault]; c.isPositive() {
		logLevel = slog.LevelError
	} else if c := m.counters[CounterKeyPanicked]; c.isPositive() {
		logLevel = slog.LevelError
	}

	logger := l.Logger
	if logger == nil {
		logger = slog.Default()
	}

	if l.Group != "" {
		logger.LogAttrs(ctx, logLevel, l.Msg, slog.GroupAttrs(l.Group, attrs...))
	} else {
		logger.LogAttrs(ctx, logLevel, l.Msg, attrs...)
	}

	return nil
}

func (m *Metrics) attrs(noCustomFormatter bool) (attrs []slog.Attr) {
	m.init()

	if m.End.IsZero() {
		m.End = time.Now()
	}

	if noCustomFormatter {
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

		if noCustomFormatter {
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
	return slog.GroupValue(m.attrs(true)...)
}

func (p *property) attr(key string) slog.Attr {
	switch p.t {
	case stringKind:
		return slog.String(key, p.v.(string))
	case int64Kind:
		return slog.Int64(key, p.v.(int64))
	case float64Kind:
		return slog.Float64(key, p.v.(float64))
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
