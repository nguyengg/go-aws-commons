package types

import (
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testUnixTime     int64 = 1136239445
	testUnixTimeTime       = time.Unix(testUnixTime, 0)
	testUnixTimeStr        = strconv.FormatInt(testUnixTime, 10)
)

func TestUnixTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		v    UnixTime
		want string
	}{
		{
			name: "marshal",
			v:    UnixTime(testUnixTimeTime),
			want: testUnixTimeStr,
		},
		{
			name: "marshal zero value",
			v:    UnixTime{},
			want: "null",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.MarshalJSON()
			assert.NoErrorf(t, err, "MarshalJSON(%#v) error", tt.v)
			assert.JSONEqf(t, tt.want, string(got), "MarshalJSON(%#v) got = %s, want = %s", tt.v, got, tt.want)
		})
	}
}

func TestUnixTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want UnixTime
	}{
		{
			name: "unmarshal",
			data: []byte(testUnixTimeStr),
			want: UnixTime(testUnixTimeTime),
		},
		{
			name: "unmarshal null value",
			data: []byte("null"),
			want: UnixTime{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnixTime(time.Now())
			err := got.UnmarshalJSON(tt.data)
			assert.NoErrorf(t, err, "UnmarshalJSON(%#v) error", tt.data)
			assert.Truef(t, tt.want.Equal(got), "UnmarshalJSON(%#v) got = %v, want = %v", tt.data, got, tt.want)
		})
	}
}

func TestUnixTime_MarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name    string
		v       UnixTime
		want    types.AttributeValue
		wantErr bool
	}{
		{
			name: "marshal ddb",
			v:    UnixTime(testUnixTimeTime),
			want: &types.AttributeValueMemberN{Value: testUnixTimeStr},
		},
		{
			name: "marshal zero value",
			v:    UnixTime{},
			want: &types.AttributeValueMemberNULL{Value: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.MarshalDynamoDBAttributeValue()
			assert.NoErrorf(t, err, "MarshalDynamoDBAttributeValue(%#v)", tt.v)
			assert.Equalf(t, tt.want, got, "MarshalDynamoDBAttributeValue(%#v) got = %v, want = %v", tt.v, got, tt.want)
		})
	}
}

func TestUnixTime_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name    string
		av      types.AttributeValue
		want    UnixTime
		wantErr bool
	}{
		{
			name: "unmarshall",
			av:   &types.AttributeValueMemberN{Value: testUnixTimeStr},
			want: UnixTime(testUnixTimeTime),
		},
		{
			name: "unmarshall null",
			av:   &types.AttributeValueMemberNULL{Value: true},
			want: UnixTime{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnixTime(time.Now())
			err := got.UnmarshalDynamoDBAttributeValue(tt.av)
			assert.NoErrorf(t, err, "UnmarshalDynamoDBAttributeValue(%#v)", tt.av)
			assert.Truef(t, tt.want.Equal(got), "UnmarshalDynamoDBAttributeValue(%#v) got = %v, want = %v", tt.av, got, tt.want)
		})
	}
}

func TestUnixTime_Sub(t *testing.T) {
	a := time.Now()
	b := a.Add(1*time.Second + 1*time.Microsecond) // the microsecond is truncated.
	assert.Equal(t, -1*time.Second, UnixTime(a).Sub(UnixTime(b)))
	assert.Equal(t, -1*time.Second, a.Truncate(time.Millisecond).Sub(b.Truncate(time.Millisecond)))
}

func TestUnixTime_Equal(t *testing.T) {
	n := time.Now()
	a := UnixTime(n)
	b := UnixTime(n.Add(1 * time.Microsecond))
	assert.True(t, a.Equal(b))
	assert.NotEqual(t, a, b)                       // b is 1 microsecond after a
	assert.NotEqual(t, time.Time(a), time.Time(b)) // same reason above
}

func TestUnixTime_TaggedUnixTime(t *testing.T) {
	type Item struct {
		Created  UnixTime  `dynamodbav:"created"`
		Modified UnixTime  `dynamodbav:"modified,unixtime"`
		Accessed time.Time `dynamodbav:"accessed,unixtime"`
	}

	n := time.Unix(1136239445, 0)
	item, err := attributevalue.Marshal(Item{
		Created:  UnixTime(n),
		Modified: UnixTime(n),
		Accessed: n,
	})
	require.NoError(t, err)
	require.Equal(t,
		map[string]types.AttributeValue{
			"created":  &types.AttributeValueMemberN{Value: "1136239445"},
			"modified": &types.AttributeValueMemberN{Value: "1136239445"},
			"accessed": &types.AttributeValueMemberN{Value: "1136239445"},
		},
		item.(*types.AttributeValueMemberM).Value)
}

func TestUnixTime_MarshalingQuirk(t *testing.T) {
	var (
		u UnixTime
		n = time.Now()
	)
	av, _ := UnixTime(n).MarshalDynamoDBAttributeValue()
	_ = u.UnmarshalDynamoDBAttributeValue(av)
	assert.NotEqual(t, n, time.Time(u))  // comparing using time.Time will always fail since n has microsecond and nanosecond components.
	assert.True(t, UnixTime(n).Equal(u)) // comparing using UnixTime.Equal will work.
}
