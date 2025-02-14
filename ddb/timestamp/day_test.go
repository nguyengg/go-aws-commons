package timestamp

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nguyengg/golambda/must"
	"reflect"
	"testing"
	"time"
)

const (
	testDayValueInRFC3339   = "2006-01-02T15:04:05.999Z"
	testDayValueInDayFormat = "2006-01-02"
)

func TestDay_MarshalJSON(t *testing.T) {
	type fields struct {
		v time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name:   "marshall success",
			fields: fields{v: Must(time.Parse(time.RFC3339, testDayValueInRFC3339))},
			want:   []byte("\"" + testDayValueInDayFormat + "\""),
		},
		{
			name:   "marshall zero-value",
			fields: fields{v: time.Time{}},
			want:   []byte("\"0001-01-01\""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Day{
				v: tt.fields.v,
			}
			got, err := d.MarshalJSON()
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

func TestDay_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Day
		wantErr bool
	}{
		{
			name: "unmarshall success",
			args: args{data: []byte("\"" + testDayValueInDayFormat + "\"")},
			want: TruncateToStartOfDay(Must(time.Parse(time.RFC3339, testDayValueInRFC3339))),
		},
		{
			name: "unmarshall zero-value",
			args: args{data: []byte("\"0001-01-01\"")},
			want: TruncateToStartOfDay(time.Time{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Day{}
			if err := got.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDay_MarshalDynamoDBAttributeValue(t *testing.T) {
	type fields struct {
		v time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		want    types.AttributeValue
		wantErr bool
	}{
		{
			name:   "marshall success",
			fields: fields{v: Must(time.Parse(time.RFC3339, testDayValueInRFC3339))},
			want:   &types.AttributeValueMemberS{Value: testDayValueInDayFormat},
		},
		{
			name:   "marshall zero-value",
			fields: fields{v: time.Time{}},
			want:   &types.AttributeValueMemberS{Value: "0001-01-01"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Day{
				v: tt.fields.v,
			}
			got, err := d.MarshalDynamoDBAttributeValue()
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

func TestDay_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	type args struct {
		av types.AttributeValue
	}
	tests := []struct {
		name    string
		args    args
		want    Day
		wantErr bool
	}{
		{
			name: "unmarshall success",
			args: args{av: &types.AttributeValueMemberS{Value: testDayValueInDayFormat}},
			want: TruncateToStartOfDay(Must(time.Parse(time.RFC3339, testDayValueInRFC3339))),
		},
		{
			name: "unmarshall zero-value",
			args: args{av: &types.AttributeValueMemberS{Value: "0001-01-01"}},
			want: TruncateToStartOfDay(time.Time{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Day{}
			if err := got.UnmarshalDynamoDBAttributeValue(tt.args.av); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalDynamoDBAttributeValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
