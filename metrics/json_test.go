package metrics

import (
	"bytes"
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_JSONLogger(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var buf bytes.Buffer

		f := Factory{Logger: &JSONLogger{Out: &buf}}
		m := f.New()
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		assert.NoError(t, m.Close())
		assert.JSONEq(t,
			`{
    "counters": {
        "fault": 0,
        "panicked": 0,
        "userDidSomethingCool": 1
    },
    "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
    "hello": "world",
    "duration": "3s",
    "pi": 3.14,
    "startTime": 946684800000,
    "status": 418
}`, buf.String())
	})
}
