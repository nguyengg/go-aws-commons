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

func TestMetrics_Log(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := New()
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		var buf bytes.Buffer
		logger := zerolog.New(&buf)

		m.Log(logger.Info().Timestamp()).Send()

		assert.JSONEq(
			t,
			`{
    "level": "info",
    "time": "1999-12-31T16:00:03-08:00",
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

		buf.Reset()

		m.RawFormatting = true

		m.Log(logger.Info().Timestamp()).Send()

		assert.JSONEq(
			t,
			`{
    "level": "info",
    "time": "1999-12-31T16:00:03-08:00",
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
