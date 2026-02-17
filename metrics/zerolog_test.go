package metrics

import (
	"bytes"
	"testing"
	"testing/synctest"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_Log(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := New()
		time.Sleep(3 * time.Second)
		m.AddCounter("userDidSomethingCool", 1)

		var buf bytes.Buffer
		logger := zerolog.New(&buf)

		m.Log(logger.Info().Timestamp()).Send()

		assert.JSONEq(
			t,
			`{
    "time":"1999-12-31T16:00:03-08:00",
    "level": "info",
    "startTime": 946684800000,
    "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
    "latency": "3s",
    "counters": {
        "userDidSomethingCool": 1,
        "fault": 0,
        "panicked": 0
    }
}`,
			buf.String())
	})
}
