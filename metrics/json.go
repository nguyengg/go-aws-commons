package metrics

import (
	"encoding/json"
	"time"
)

func (m *Metrics) MarshalJSON() ([]byte, error) {
	m.init()

	if m.End.IsZero() {
		m.End = time.Now()
	}

	res := map[string]any{
		"startTime": m.Start.UnixMilli(),
		"endTime":   m.End.UTC().Format(time.RFC1123),
		"latency":   FormatDuration(m.End.Sub(m.Start)),
	}

	for k, v := range m.properties {
		res[k] = v.v
	}

	if len(m.counters) != 0 {
		counters := map[string]any{}

		for k, c := range m.counters {
			counters[k] = c.v
		}

		res["counters"] = counters
	}

	if len(m.timings) != 0 {
		timings := map[string]any{}

		for k, t := range m.timings {
			timings[k] = map[string]any{
				"sum": FormatDuration(t.sum),
				"min": FormatDuration(t.min),
				"max": FormatDuration(t.max),
				"n":   t.n,
			}
		}

		res["timings"] = timings
	}

	return json.Marshal(res)
}
