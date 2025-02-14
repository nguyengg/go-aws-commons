package timestamp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/stretchr/testify/assert"
)

const (
	testEpochMillisecondValueInRFC3339   = "2006-01-02T15:04:05.999Z"
	testEpochMillisecondValueInUnixMilli = "1136214245999"
)

func TestEpochMillisecond_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		e    EpochMillisecond
		want []byte
	}{
		{
			name: "marshal",
			e:    EpochMillisecond(Must(time.Parse(time.RFC3339, testEpochMillisecondValueInRFC3339))),
			want: []byte(testEpochMillisecondValueInUnixMilli),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.MarshalJSON()
			assert.NoError(t, err, "MarshalJSON() error = %v", err)
			assert.Equalf(t, tt.want, got, "MarshalJSON() got = %v, want = %v", got, tt.want)
		})
	}
}

func TestEpochMillisecond_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "unmarshal",
			args: args{data: []byte(testEpochMillisecondValueInUnixMilli)},
			want: Must(time.Parse(time.RFC3339, testEpochMillisecondValueInRFC3339)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EpochMillisecond(time.Now())
			err := got.UnmarshalJSON(tt.args.data)
			assert.NoError(t, err, "UnmarshalJSON() error = %v", err)
			assert.Equalf(t, tt.want, got.ToTime(), "UnmarshalJSON() got = %v, want = %v", got, tt.want)
		})
	}
}

func TestEpochMillisecond_MarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name    string
		e       EpochMillisecond
		want    types.AttributeValue
		wantErr bool
	}{
		{
			name: "marshal ddb",
			e:    EpochMillisecond(Must(time.Parse(time.RFC3339, testEpochMillisecondValueInRFC3339))),
			want: &types.AttributeValueMemberN{Value: testEpochMillisecondValueInUnixMilli},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.MarshalDynamoDBAttributeValue()
			assert.NoError(t, err, "MarshalDynamoDBAttributeValue() error = %v", err)
			assert.Equalf(t, tt.want, got, "MarshalDynamoDBAttributeValue() got = %v, want = %v", got, tt.want)
		})
	}
}

func TestEpochMillisecond_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	type args struct {
		av types.AttributeValue
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "unmarshall ddb",
			args: args{av: &types.AttributeValueMemberN{Value: testEpochMillisecondValueInUnixMilli}},
			want: Must(time.Parse(time.RFC3339, testEpochMillisecondValueInRFC3339)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EpochMillisecond(time.Now())
			err := got.UnmarshalDynamoDBAttributeValue(tt.args.av)
			assert.NoError(t, err, "UnmarshalDynamoDBAttributeValue() error = %v", err)
			assert.Equalf(t, tt.want, got.ToTime(), "UnmarshalDynamoDBAttributeValue() got = %v, want = %v", got, tt.want)
		})
	}
}

func TestEpochMillisecond_TruncateNanosecond(t *testing.T) {
	v, err := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999Z")
	if err != nil {
		t.Error(err)
	}

	data, err := json.Marshal(EpochMillisecond(v))
	assert.NoError(t, err, "Marshal() error = %v", err)

	got := EpochMillisecond(time.Time{})
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err, "Unmarshal() error = %v", err)

	// got's underlying time.time is truncated to 2006-01-02T15:04:05.
	assert.NotEqualf(t, v, got.ToTime(), "shouldn't be equal; got %v, want %v", got, v)

	// if we reset v's nano time, then they are equal.
	v = time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), got.ToTime().Nanosecond(), v.Location())
	assert.Equalf(t, v, got.ToTime(), "got %#v, want %#v", got, v)
}
