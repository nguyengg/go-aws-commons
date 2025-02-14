package timestamp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	want := "{\"day\":\"2006-01-02\",\"timestamp\":\"2006-01-02T15:04:05.999Z\",\"epochMillisecond\":1136214245999,\"epochSecond\":1136214245}"

	// non-pointer version.
	got, err := json.Marshal(item)
	assert.NoError(t, err)
	assert.JSONEq(t, want, string(got))

	// pointer version.
	got, err = json.Marshal(&item)
	assert.NoError(t, err)
	assert.JSONEq(t, want, string(got))

	newItem := JSONItem{}
	err = json.Unmarshal([]byte(want), &newItem)
	assert.NoError(t, err)
	assert.Equal(t, item, newItem)
}
