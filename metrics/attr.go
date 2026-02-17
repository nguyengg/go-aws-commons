package metrics

import (
	"fmt"
	"log/slog"
	"time"
)

type startTimeValue struct {
	time.Time
}

func (t startTimeValue) LogValue() slog.Value {
	return slog.Int64Value(t.UnixMilli())
}

type endTimeValue struct {
	time.Time
}

func (t endTimeValue) LogValue() slog.Value {
	return slog.StringValue(t.Format(time.RFC1123))
}

type durationValue struct {
	time.Duration
}

func (d durationValue) LogValue() slog.Value {
	if v := d.Duration; v >= 1*time.Second {
		return slog.StringValue(fmt.Sprintf("%.3fs", d.Seconds()))
	} else {
		return slog.StringValue(fmt.Sprintf("%.3fms", float64(v)/float64(time.Millisecond)))
	}
}
