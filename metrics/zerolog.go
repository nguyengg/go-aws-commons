package metrics

import (
	"time"

	"github.com/rs/zerolog"
)

// LogWithZerolog will change Close to log with zerolog instead.
//
// See ZerologOptions for more options.
func LogWithZerolog(optFns ...func(opts *ZerologOptions)) func(*Metrics) {
	opts := &ZerologOptions{}

	for _, fn := range optFns {
		fn(opts)
	}

	return func(m *Metrics) {
		m.logger = opts
	}
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
			Dur(ReservedKeyLatency, m.End.Sub(m.Start))
	} else {
		e.
			Int64(ReservedKeyStartTime, m.Start.UnixMilli()).
			Str(ReservedKeyEndTime, m.End.UTC().Format(time.RFC1123)).
			Str(ReservedKeyLatency, FormatDuration(m.End.Sub(m.Start)))
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
