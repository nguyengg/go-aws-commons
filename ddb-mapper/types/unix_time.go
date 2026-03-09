package types

import (
	"cmp"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// UnixTime is [time.Time] encoded as Unix time in seconds using DynamoDB N data type.
//
// UnixTime is a good candidate for a [time-to-live] value though be careful of how the default encoder handles (or
// doesn't handle) zero-value [time.Time]. It would be better to use *UnixTime in that case.
//
// UnixTime's methods always work at second precision. However, converting it to a [time.Time] may cause precision
// issues. For example:
//
//	now := time.Now()
//	a := types.UnixTime(now)
//	b := types.UnixTime(now.Add(1 * time.Microsecond))
//	fmt.Printf("a.Equal(b): %t\n", a.Equal(b))                                             // prints True
//	fmt.Printf("a == b: %t\n", a == b)                                                     // prints False because b is 1 microsecond after a
//	fmt.Printf("time.Time(a).Equal(time.Time(b)): %t\n", time.Time(a).Equal(time.Time(b))) // prints False for same reason above
//
// In a struct tag, these three fields are equivalent in terms of encoding/decoding:
//
//	type Item struct {
//		Created  UnixTime  `dynamodbav:"created"`
//		Modified UnixTime  `dynamodbav:"modified,unixtime"`
//		Accessed time.Time `dynamodbav:"accessed,unixtime"`
//	}
//
// Note that because UnixTime truncates its underlying [time.Time] when marshaling to a DynamoDB [types.AttributeValue],
// unmarshalling that same [types.AttributeValue] may not produce an equal [time.Time], but [UnixTime.Equal] works:
//
//	var (
//		u UnixTime
//		n = time.Now()
//	)
//	av, _ := UnixTime(n).MarshalDynamoDBAttributeValue()
//	_ = u.UnmarshalDynamoDBAttributeValue(av)
//	assert.NotEqual(t, n, time.Time(u))  // comparing using time.Time will always fail since n has microsecond and nanosecond components.
//	assert.True(t, UnixTime(n).Equal(u)) // comparing using UnixTime.Equal will work.
//
// [time-to-live]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/TTL.html
type UnixTime time.Time

// String returns the epoch second string presentation.
func (t UnixTime) String() string {
	return strconv.FormatInt(t.Unix(), 10)
}

// After implements [time.Time.After] at second precision.
func (t UnixTime) After(u UnixTime) bool {
	return t.Unix() > u.Unix()
}

// Before implements [time.Time.Before] at second precision.
func (t UnixTime) Before(u UnixTime) bool {
	return t.Unix() < u.Unix()
}

// UnixTime delegates to [time.Unix].
func (t UnixTime) Unix() int64 {
	return time.Time(t).Unix()
}

// Sub implements [time.Time.Sub] at second precision.
func (t UnixTime) Sub(u UnixTime) time.Duration {
	return time.Duration(t.Unix()-u.Unix()) * time.Second
}

// Equal implements [time.Time.Equal] at second precision.
func (t UnixTime) Equal(u UnixTime) bool {
	return t.Unix() == u.Unix()
}

// Compare implements [time.Time.Compare] at second precision.
func (t UnixTime) Compare(u UnixTime) int {
	return cmp.Compare(t.Unix(), u.Unix())
}

// Format implements [time.Time.Format] at second precision.
func (t UnixTime) Format(layout string) string {
	return time.Unix(t.Unix(), 0).Format(layout)
}

// IsZero delegates to [time.Time.IsZero].
func (t UnixTime) IsZero() bool {
	return time.Time(t).IsZero()
}

var _ attributevalue.Marshaler = (*UnixTime)(nil)
var _ attributevalue.Unmarshaler = (*UnixTime)(nil)

func (t UnixTime) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	if t.IsZero() {
		return &types.AttributeValueMemberNULL{Value: true}, nil
	}

	return &types.AttributeValueMemberN{Value: t.String()}, nil
}

func (t *UnixTime) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	switch av := value.(type) {
	case *types.AttributeValueMemberNULL:
		*t = UnixTime{}

	case *types.AttributeValueMemberN:
		v, err := strconv.ParseInt(av.Value, 10, 64)
		if err != nil {
			return err
		}

		*t = UnixTime(time.Unix(v, 0))

	default:
		return fmt.Errorf("cannot unmarshal attribute value type %T as UnixTime", av)
	}

	return nil
}

var _ json.Marshaler = (*UnixTime)(nil)
var _ json.Unmarshaler = (*UnixTime)(nil)

func (t UnixTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return strconv.AppendInt(nil, t.Unix(), 10), nil
}

func (t *UnixTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*t = UnixTime{}
		return nil
	}

	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}

	*t = UnixTime(time.Unix(n, 0))
	return nil
}
