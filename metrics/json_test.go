package metrics

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_MarshalJSON(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := New()
		time.Sleep(3 * time.Second)
		m.AddCounter("userDidSomethingCool", 1)

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
    "latency": "3s",
    "startTime": 946684800000
}`,
			string(data))
	})
}
