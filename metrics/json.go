package metrics

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"
)

// JSONLogger logs Metrics instance as JSON.
//
// The zero-value struct is ready for use.
type JSONLogger struct {
	// Out is the io.Writer to write JSON content.
	//
	// Default to os.Stderr.
	Out io.Writer
}

// Log implements Logger.Log.
//
// The method returns a non-nil error if there was error encoding the Metrics instance as JSON.
func (l JSONLogger) Log(ctx context.Context, m *Metrics) error {
	w := l.Out
	if w == nil {
		w = os.Stderr
	}

	err := json.NewEncoder(w).Encode(m)
	if err == nil {
		_, err = w.Write([]byte("\n"))
	}
	return err
}

func (m *Metrics) MarshalJSON() ([]byte, error) {
	m.init()

	if m.End.IsZero() {
		m.End = time.Now()
	}

	res := map[string]any{
		ReservedKeyStartTime: m.Start.UnixMilli(),
		ReservedKeyEndTime:   m.End.UTC().Format(time.RFC1123),
		ReservedKeyDuration:  FormatDuration(m.End.Sub(m.Start)),
	}

	for k, v := range m.properties {
		res[k] = v.v
	}

	if len(m.counters) != 0 {
		counters := map[string]any{}

		for k, c := range m.counters {
			counters[k] = c.v
		}

		res[ReservedKeyCounters] = counters
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

		res[ReservedKeyTimings] = timings
	}

	if len(m.errors) != 0 {
		res[ReservedKeyErrors] = m.errors.toJSON()
	}

	return json.Marshal(res)
}
