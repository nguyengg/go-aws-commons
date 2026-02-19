package metrics

import (
	"bytes"
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_zerolog(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)
		zerolog.DefaultContextLogger = &logger

		m := New(func(m *Metrics) {
			m.logger = &ZerologMetricsLogger{}
		})
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		assert.NoError(t, m.Close())
		assert.JSONEq(
			t,
			`{
    "level": "info",
    "startTime": 946684800000,
    "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
    "latency": "3s",
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

func TestMetrics_zerologWithDict(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)
		zerolog.DefaultContextLogger = &logger

		m := New(func(m *Metrics) {
			m.logger = &ZerologMetricsLogger{Dict: "metrics"}
		})
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		assert.NoError(t, m.Close())
		assert.JSONEq(
			t,
			`{
    "level": "info",
    "metrics": {
        "startTime": 946684800000,
        "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
        "latency": "3s",
        "status": 418,
        "pi": 3.14,
        "hello": "world",
        "counters": {
            "fault": 0,
            "panicked": 0,
            "userDidSomethingCool": 1
        }
    }
}`,
			buf.String())
	})
}

func TestMetrics_zerologNoCustomFormatter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer
		logger := zerolog.New(&buf)
		zerolog.DefaultContextLogger = &logger

		m := New(func(m *Metrics) {
			m.logger = &ZerologMetricsLogger{NoCustomFormatter: true}
		})
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		assert.NoError(t, m.Close())
		assert.JSONEq(
			t,
			`{
    "level": "info",
    "startTime": "1999-12-31T16:00:00-08:00",
    "endTime": "1999-12-31T16:00:03-08:00",
    "latency": 3000,
    "status": 418,
    "pi": 3.14,
    "hello": "world",
    "counters": {
        "fault": 0,
        "panicked": 0,
        "userDidSomethingCool": 1
    }
}`,
			buf.String())
	})
}
