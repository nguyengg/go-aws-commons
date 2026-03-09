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

// UnixMilli is [time.Time] encoded as Unix time in milliseconds using DynamoDB N data type.
//
// UnixMilli's methods always work at millisecond precision. However, converting it to a [time.Time] may cause precision
// issues. For example:
//
//	now := time.Now()
//	a := types.UnixMilli(now)
//	b := types.UnixMilli(now.Add(1 * time.Microsecond))
//	fmt.Printf("a.Equal(b): %t\n", a.Equal(b))                                             // prints True
//	fmt.Printf("a == b: %t\n", a == b)                                                     // prints False because b is 1 microsecond after a
//	fmt.Printf("time.Time(a).Equal(time.Time(b)): %t\n", time.Time(a).Equal(time.Time(b))) // prints False for same reason above
//
// If you accidentally tag this field with "unixtime", [attributevalue.Encoder] and [attributevalue.Decoder] should
// still correctly respect the [attributevalue.Marshaler] and [attributevalue.Unmarshaler] implementations. For example,
// the "unixtime" tag on Modified field has no effect:
//
//	type Item struct {
//		Created  UnixTime  `dynamodbav:"created"`
//		Modified UnixTime  `dynamodbav:"created,unixtime"`
//	}
//
// Note that because UnixTime truncates its underlying [time.Time] when marshaling to a DynamoDB [types.AttributeValue],
// unmarshalling that same [types.AttributeValue] may not produce an equal [time.Time], but [UnixMilli.Equal] works:
//
//	var (
//		u UnixMilli
//		n = time.Now()
//	)
//	av, _ := UnixMilli(n).MarshalDynamoDBAttributeValue()
//	_ = u.UnmarshalDynamoDBAttributeValue(av)
//	assert.NotEqual(t, n, time.Time(u))   // comparing using time.Time will always fail since n has microsecond and nanosecond components.
//	assert.True(t, UnixMilli(n).Equal(u)) // comparing using UnixTime.Equal will work.
type UnixMilli time.Time

// String returns the epoch millisecond string presentation.
func (t UnixMilli) String() string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}

// After implements [time.Time.After] at millisecond precision.
func (t UnixMilli) After(u UnixMilli) bool {
	return t.UnixMilli() > u.UnixMilli()
}

// Before implements [time.Time.Before] at millisecond precision.
func (t UnixMilli) Before(u UnixMilli) bool {
	return t.UnixMilli() < u.UnixMilli()
}

// UnixMilli delegates to [time.UnixMilli].
func (t UnixMilli) UnixMilli() int64 {
	return time.Time(t).UnixMilli()
}

// Sub implements [time.Time.Sub] at millisecond precision.
func (t UnixMilli) Sub(u UnixMilli) time.Duration {
	return time.Duration(t.UnixMilli()-u.UnixMilli()) * time.Millisecond
}

// Equal implements [time.Time.Equal] at millisecond precision.
func (t UnixMilli) Equal(u UnixMilli) bool {
	return t.UnixMilli() == u.UnixMilli()
}

// Compare implements [time.Time.Compare] at millisecond precision.
func (t UnixMilli) Compare(u UnixMilli) int {
	return cmp.Compare(t.UnixMilli(), u.UnixMilli())
}

// Format implements [time.Time.Format] at millisecond precision.
func (t UnixMilli) Format(layout string) string {
	return time.UnixMilli(t.UnixMilli()).Format(layout)
}

// IsZero delegates to [time.Time.IsZero].
func (t UnixMilli) IsZero() bool {
	return time.Time(t).IsZero()
}

var _ attributevalue.Marshaler = (*UnixMilli)(nil)
var _ attributevalue.Unmarshaler = (*UnixMilli)(nil)

func (t UnixMilli) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	if t.IsZero() {
		return &types.AttributeValueMemberNULL{Value: true}, nil
	}

	return &types.AttributeValueMemberN{Value: t.String()}, nil
}

func (t *UnixMilli) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	switch av := value.(type) {
	case *types.AttributeValueMemberNULL:
		*t = UnixMilli{}

	case *types.AttributeValueMemberN:
		v, err := strconv.ParseInt(av.Value, 10, 64)
		if err != nil {
			return err
		}

		*t = UnixMilli(time.UnixMilli(v))

	default:
		return fmt.Errorf("cannot unmarshal attribute value type %T as UnixMilli", av)
	}

	return nil
}

var _ json.Marshaler = (*UnixMilli)(nil)
var _ json.Unmarshaler = (*UnixMilli)(nil)

func (t UnixMilli) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return strconv.AppendInt(nil, t.UnixMilli(), 10), nil
}

func (t *UnixMilli) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*t = UnixMilli{}
		return nil
	}

	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}

	*t = UnixMilli(time.UnixMilli(n))
	return nil
}
