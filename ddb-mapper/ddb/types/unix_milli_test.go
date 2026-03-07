package types

import (
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/ddb-local-test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testUnixMilli     int64 = 1136214245999
	testUnixMilliTime       = time.UnixMilli(testUnixMilli)
	testUnixMilliStr        = strconv.FormatInt(testUnixMilli, 10)
)

func TestUnixMilli_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		v    UnixMilli
		want string
	}{
		{
			name: "marshal",
			v:    UnixMilli(testUnixMilliTime),
			want: testUnixMilliStr,
		},
		{
			name: "marshal zero value",
			v:    UnixMilli{},
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

func TestUnixMilli_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want UnixMilli
	}{
		{
			name: "unmarshal",
			data: []byte(testUnixMilliStr),
			want: UnixMilli(testUnixMilliTime),
		},
		{
			name: "unmarshal null value",
			data: []byte("null"),
			want: UnixMilli{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnixMilli(time.Now())
			err := got.UnmarshalJSON(tt.data)
			assert.NoErrorf(t, err, "UnmarshalJSON(%#v) error", tt.data)
			assert.Truef(t, tt.want.Equal(got), "UnmarshalJSON(%#v) got = %v, want = %v", tt.data, got, tt.want)
		})
	}
}

func TestUnixMilli_MarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name    string
		v       UnixMilli
		want    types.AttributeValue
		wantErr bool
	}{
		{
			name: "marshal",
			v:    UnixMilli(testUnixMilliTime),
			want: &types.AttributeValueMemberN{Value: testUnixMilliStr},
		},
		{
			name: "marshal zero value",
			v:    UnixMilli{},
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

func TestUnixMilli_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name    string
		av      types.AttributeValue
		want    UnixMilli
		wantErr bool
	}{
		{
			name: "unmarshall",
			av:   &types.AttributeValueMemberN{Value: testUnixMilliStr},
			want: UnixMilli(testUnixMilliTime),
		},
		{
			name: "unmarshall null",
			av:   &types.AttributeValueMemberNULL{Value: true},
			want: UnixMilli{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnixMilli(time.Now())
			err := got.UnmarshalDynamoDBAttributeValue(tt.av)
			assert.NoErrorf(t, err, "UnmarshalDynamoDBAttributeValue(%#v)", tt.av)
			assert.Truef(t, tt.want.Equal(got), "UnmarshalDynamoDBAttributeValue(%#v) got = %v, want = %v", tt.av, got, tt.want)
		})
	}
}

func TestUnixMilli_Sub(t *testing.T) {
	a := time.Now()
	b := a.Add(1*time.Second + 1*time.Microsecond) // the microsecond is truncated.
	assert.Equal(t, -1*time.Second, UnixMilli(a).Sub(UnixMilli(b)))
	assert.Equal(t, -1*time.Second, a.Truncate(time.Millisecond).Sub(b.Truncate(time.Millisecond)))
}

func TestUnixMilli_Equal(t *testing.T) {
	n := time.Now()
	a := UnixMilli(n)
	b := UnixMilli(n.Add(1 * time.Microsecond))
	assert.True(t, a.Equal(b))
	assert.NotEqual(t, a, b)                       // b is 1 microsecond after a
	assert.NotEqual(t, time.Time(a), time.Time(b)) // same reason above
}

func TestUnixMilli_TaggedUnixTime(t *testing.T) {
	type Item struct {
		Created  UnixMilli `dynamodbav:"created"`
		Modified UnixMilli `dynamodbav:"modified,unixtime"`
		Accessed time.Time `dynamodbav:"accessed,unixtime"`
	}

	n := time.UnixMilli(1136239445012)
	item, err := attributevalue.Marshal(Item{
		Created:  UnixMilli(n),
		Modified: UnixMilli(n),
		Accessed: n,
	})
	require.NoError(t, err)
	RequireHasAttributes(t, map[string]any{
		"created":  1136239445012,
		"modified": 1136239445012,
		"accessed": 1136239445,
	}, item.(*types.AttributeValueMemberM).Value)
}

func TestUnixMilli_MarshalingQuirk(t *testing.T) {
	var (
		u UnixMilli
		n = time.Now()
	)
	av, _ := UnixMilli(n).MarshalDynamoDBAttributeValue()
	_ = u.UnmarshalDynamoDBAttributeValue(av)
	assert.NotEqual(t, n, time.Time(u))   // comparing using time.Time will always fail since n has microsecond and nanosecond components.
	assert.True(t, UnixMilli(n).Equal(u)) // comparing using UnixTime.Equal will work.
}
