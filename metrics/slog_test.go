package metrics

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_Attrs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := New()
		m.String("hello", "world")
		m.Int64("status", http.StatusTeapot)
		m.Float64("pi", 3.14)
		m.AddCounter("userDidSomethingCool", 1)
		time.Sleep(3 * time.Second) // to create latency metrics.

		var buf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
		slog.LogAttrs(context.Background(), slog.LevelInfo, "request done", m.Attrs()...)

		assert.JSONEq(t,
			`{
    "time": "1999-12-31T16:00:03-08:00",
    "level": "INFO",
    "msg": "request done",
    "startTime": 946684800000,
    "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
    "latency": "3s",
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

		buf.Reset()

		slog.LogAttrs(context.Background(), slog.LevelInfo, "request done", slog.Any("metrics", m))

		assert.JSONEq(t,
			`{
    "time": "1999-12-31T16:00:03-08:00",
    "level": "INFO",
    "msg": "request done",
    "metrics": {
        "startTime": 946684800000,
        "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
        "latency": "3s",
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

		buf.Reset()

		m.RawFormatting = true

		slog.LogAttrs(context.Background(), slog.LevelInfo, "request done", slog.Any("metrics", m))

		assert.JSONEq(t,
			`{
    "time": "1999-12-31T16:00:03-08:00",
    "level": "INFO",
    "msg": "request done",
    "metrics": {
        "startTime": "1999-12-31T16:00:00-08:00",
        "endTime": "1999-12-31T16:00:03-08:00",
        "latency": 3000000000,
        "hello": "world",
        "status": 418,
        "pi": 3.14,
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
