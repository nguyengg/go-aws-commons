package metrics

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestSimpleMetrics(t *testing.T) {
	start, _ := time.Parse(time.StampMilli, time.StampMilli)
	m := NewWithStartTime(start)
	m.AddTiming("someLongMethod", 21*time.Second)
	m.AddCount("success", 1)
	m.SetProperty("user", "henry")
	m.SetStatusCode(http.StatusTeapot)

	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)
	end := start.Add(1 * time.Hour)
	m.LogWithEndTime(&logger, end)

	got := string(buf.Bytes())
	want := "{" +
		"\"startTime\":-6795364578871," +
		"\"endTime\":\"Mon, 01 Jan 0001 01:00:00 GMT\"," +
		"\"time\":\"3600000.000 ms\"," +
		"\"user\":\"henry\"," +
		"\"statusCode\":418," +
		"\"counters\":{" +
		"\"2xx\":0," +
		"\"4xx\":1," +
		"\"5xx\":0," +
		"\"fault\":0," +
		"\"panicked\":0," +
		"\"success\":1}," +
		"\"timings\":{" +
		"\"someLongMethod\":{" +
		"\"sum\":\"21000.000 ms\"," +
		"\"min\":\"21000.000 ms\"," +
		"\"max\":\"21000.000 ms\"," +
		"\"n\":1," +
		"\"avg\":\"21000.000 ms\"}}}\n"
	assert.JSONEqf(t, want, got, "LogWithEndTime() got = %v, want = %v", got, want)
}
