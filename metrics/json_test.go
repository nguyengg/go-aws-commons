package metrics

import (
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_MarshalJSON(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := New()
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		data, err := m.MarshalJSON()
		assert.NoError(t, err)
		assert.JSONEq(t,
			`{
    "counters": {
        "fault": 0,
        "panicked": 0,
        "userDidSomethingCool": 1
    },
    "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
    "hello": "world",
    "latency": "3s",
    "pi": 3.14,
    "startTime": 946684800000,
    "status": 418
}`,
			string(data))
	})
}
