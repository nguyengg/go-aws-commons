package metrics

import (
	"context"
	"log"
	"time"

	"github.com/rs/zerolog"
)

// ZerologMetricsLogger implements Logger using zerolog.Logger.
type ZerologMetricsLogger struct {
	// Logger is the zerolog.Logger instance to use.
	//
	// Default to zerolog.Ctx.
	Logger *zerolog.Logger

	// Level is the log level to use.
	//
	// By default, the log level is dynamic. If the Metrics instance indicates an error state with non-zero fault
	// and/or panicked counter, zerolog.ErrorLevel is used. Otherwise, zerolog.InfoLevel is used.
	Level *zerolog.Level

	// Msg is the message to be used with the log.
	Msg string

	// Dict is the name of the [zerolog.Event.Dict] to place the Metrics content.
	//
	// Equivalent to SlogMetricsLogger.Group.
	Dict string

	// NoCustomFormatter disables specialised formatting for timestamps and durations.
	//
	// By default, to match LogJSON, the start time is logged as epoch millisecond, the end time using RFC1123, and
	// all time.Duration using FormatDuration. To disable this behaviour and rely on slog's own formatters, set
	// NoCustomFormatter to true.
	NoCustomFormatter bool
}

func (l *ZerologMetricsLogger) Log(ctx context.Context, m *Metrics) error {
	logger := zerolog.Ctx(ctx)

	var e *zerolog.Event
	if c := m.counters[CounterKeyFault]; c.isPositive() {
		e = logger.Error()
	} else if c := m.counters[CounterKeyPanicked]; c.isPositive() {
		e = logger.Error()
	} else {
		e = logger.Info()
	}

	if l.Dict != "" {
		e = e.Dict(l.Dict, m.e(zerolog.Dict(), l.NoCustomFormatter))
	} else {
		e = m.e(e, l.NoCustomFormatter)
	}

	e.Msg(l.Msg)

	return nil
}

// e set the Metrics.End (if not set) and adds fields to the given zerolog.Event.
func (m *Metrics) e(e *zerolog.Event, noCustomerFormatter bool) *zerolog.Event {
	m.init()

	if m.End.IsZero() {
		m.End = time.Now()
	}

	if noCustomerFormatter {
		e.
			Time(ReservedKeyStartTime, m.Start).
			Time(ReservedKeyEndTime, m.End).
			Dur(ReservedKeyDuration, m.End.Sub(m.Start))
	} else {
		e.
			Int64(ReservedKeyStartTime, m.Start.UnixMilli()).
			Str(ReservedKeyEndTime, m.End.UTC().Format(time.RFC1123)).
			Str(ReservedKeyDuration, FormatDuration(m.End.Sub(m.Start)))
	}

	for k, p := range m.properties {
		p.e(k, e)
	}

	if len(m.counters) != 0 {
		d := zerolog.Dict()
		for k, c := range m.counters {
			c.e(k, d)
		}

		e.Dict(ReservedKeyCounters, d)
	}

	if len(m.timings) != 0 {
		d := zerolog.Dict()

		if noCustomerFormatter {
			for k, t := range m.timings {
				d.Dict(k, zerolog.Dict().
					Dur("sum", t.sum).
					Dur("min", t.min).
					Dur("max", t.max).
					Dur("avg", t.avg()).
					Int64("n", t.n))
			}
		} else {
			for k, v := range m.timings {
				d.Dict(k, zerolog.Dict().
					Str("sum", FormatDuration(v.sum)).
					Str("min", FormatDuration(v.min)).
					Str("max", FormatDuration(v.max)).
					Str("avg", FormatDuration(v.avg())).
					Int64("n", v.n))
			}
		}

		e.Dict(ReservedKeyTimings, d)
	}

	if len(m.errors) != 0 {
		data, err := m.errors.MarshalJSON()
		if err != nil {
			panic(err)
		}

		log.Printf("raw json")

		e.RawJSON(ReservedKeyErrors, data)
	}

	return e
}

func (p *property) e(key string, e *zerolog.Event) *zerolog.Event {
	switch p.t {
	case stringKind:
		return e.Str(key, p.v.(string))
	case int64Kind:
		return e.Int64(key, p.v.(int64))
	case float64Kind:
		return e.Float64(key, p.v.(float64))
	case anyKind:
		return e.Any(key, p.v.(any))
	default:
		panic("invalid property type")
	}
}

func (p *property) c(key string, e zerolog.Context) zerolog.Context {
	switch p.t {
	case stringKind:
		return e.Str(key, p.v.(string))
	case int64Kind:
		return e.Int64(key, p.v.(int64))
	case float64Kind:
		return e.Float64(key, p.v.(float64))
	case anyKind:
		return e.Any(key, p.v.(any))
	default:
		panic("invalid property type")
	}
}

func (c *counter) e(key string, e *zerolog.Event) *zerolog.Event {
	switch c.t {
	case int64Kind:
		return e.Int64(key, c.v.(int64))
	case float64Kind:
		return e.Float64(key, c.v.(float64))
	default:
		panic("invalid counter type")
	}
}
