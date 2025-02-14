package timestamp

import (
	"encoding/json"
	. "github.com/nguyengg/golambda/must"
	"reflect"
	"testing"
	"time"
)

type JSONItem struct {
	Day              Day              `json:"day"`
	Timestamp        Timestamp        `json:"timestamp"`
	EpochMillisecond EpochMillisecond `json:"epochMillisecond"`
	EpochSecond      EpochSecond      `json:"epochSecond"`
}

// TestJSON_structUsage tests using all the timestamps in a struct.
func TestJSON_structUsage(t *testing.T) {
	millisecond := Must(time.Parse(time.RFC3339, "2006-01-02T15:04:05.999Z"))
	second := Must(time.Parse(time.RFC3339, "2006-01-02T15:04:05Z"))

	item := JSONItem{
		Day:              TruncateToStartOfDay(millisecond),
		Timestamp:        Timestamp(millisecond),
		EpochMillisecond: EpochMillisecond(millisecond),
		EpochSecond:      EpochSecond(second),
	}

	want := []byte("{\"day\":\"2006-01-02\",\"timestamp\":\"2006-01-02T15:04:05.999Z\",\"epochMillisecond\":1136214245999,\"epochSecond\":1136214245}")

	// non-pointer version.
	if got := Must(json.Marshal(item)); !reflect.DeepEqual(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
	// pointer version.
	if got := Must(json.Marshal(&item)); !reflect.DeepEqual(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}

	got := JSONItem{}
	if Must0(json.Unmarshal(want, &got)); !reflect.DeepEqual(got, item) {
		t.Errorf("got %#v, want %#v", got, item)
	}
}
