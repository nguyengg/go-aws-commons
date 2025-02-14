package timestamp

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nguyengg/golambda/must"
	"reflect"
	"testing"
	"time"
)

const (
	testEpochMillisecondValueInRFC3339   = "2006-01-02T15:04:05.999Z"
	testEpochMillisecondValueInUnixMilli = "1136214245999"
)

func TestEpochMillisecond_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		e       EpochMillisecond
		want    []byte
		wantErr bool
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
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEpochMillisecond_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "unmarshal",
			args: args{data: []byte(testEpochMillisecondValueInUnixMilli)},
			want: Must(time.Parse(time.RFC3339, testEpochMillisecondValueInRFC3339)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := EpochMillisecond(time.Now())
			if err := e.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !e.ToTime().Equal(tt.want) {
				t.Errorf("got %v, want %v", e, tt.want)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalDynamoDBAttributeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalDynamoDBAttributeValue() got = %v, want %v", got, tt.want)
			}
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
			e := EpochMillisecond(time.Now())
			if err := e.UnmarshalDynamoDBAttributeValue(tt.args.av); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalDynamoDBAttributeValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !e.ToTime().Equal(tt.want) {
				t.Errorf("got %v, want %v", e, tt.want)
			}
		})
	}
}

func TestEpochMillisecond_TruncateNanosecond(t *testing.T) {
	v, err := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999Z")
	if err != nil {
		t.Error(err)
	}

	data, err := json.Marshal(EpochMillisecond(v))
	if err != nil {
		t.Error(err)
	}

	got := EpochMillisecond(time.Time{})
	if err := json.Unmarshal(data, &got); err != nil {
		t.Error(err)
	}

	// got's underlying time.time is truncated to 2006-01-02T15:04:05.
	if reflect.DeepEqual(got.ToTime(), v) {
		t.Errorf("shouldn't be equal; got %v, want %v", got, v)
	}

	// if we reset v's nano time, then they are equal.
	v = time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), got.ToTime().Nanosecond(), v.Location())
	if !reflect.DeepEqual(got.ToTime(), v) {
		t.Errorf("got %#v, want %#v", got.ToTime(), v)
	}
}
