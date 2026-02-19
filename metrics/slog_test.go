package metrics

import (
	"bytes"
	"log/slog"
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_slog(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

		m := New(func(m *Metrics) {
			m.logger = &SlogMetricsLogger{}
		})
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		assert.NoError(t, m.Close())
		assert.JSONEq(t,
			`{
    "time": "1999-12-31T16:00:03-08:00",
    "level": "INFO",
    "msg": "",
    "startTime": 946684800000,
    "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
    "duration": "3s",
    "hello": "world",
    "status": 418,
    "pi": 3.14,
    "counters": {
        "panicked": 0,
        "userDidSomethingCool": 1,
        "fault": 0
    }
}`,
			buf.String())
	})
}

func TestMetrics_slogWithGroup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

		m := New(func(m *Metrics) {
			m.logger = &SlogMetricsLogger{Group: "metrics"}
		})
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		assert.NoError(t, m.Close())
		assert.JSONEq(t,
			`{
    "time": "1999-12-31T16:00:03-08:00",
    "level": "INFO",
    "msg": "",
    "metrics": {
        "startTime": 946684800000,
        "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
        "duration": "3s",
        "hello": "world",
        "status": 418,
        "pi": 3.14,
        "counters": {
            "userDidSomethingCool": 1,
            "fault": 0,
            "panicked": 0
        }
    }
}`,
			buf.String())

	})
}

func TestMetrics_slogNoCustomFormatter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

		m := New(func(m *Metrics) {
			m.logger = &SlogMetricsLogger{NoCustomFormatter: true}
		})
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		assert.NoError(t, m.Close())
		assert.JSONEq(t,
			`{
    "time": "1999-12-31T16:00:03-08:00",
    "level": "INFO",
    "msg": "",
    "startTime": "1999-12-31T16:00:00-08:00",
    "endTime": "1999-12-31T16:00:03-08:00",
    "duration": 3000000000,
    "hello": "world",
    "status": 418,
    "pi": 3.14,
    "counters": {
        "fault": 0,
        "panicked": 0,
        "userDidSomethingCool": 1
    }
}`,
			buf.String())
	})
}
