package timestamp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// EpochMillisecond is epoch millisecond in UTC, formatted and marshalled as a positive integer (e.g. 1136214245000).
//
// Because EpochMillisecond wraps around time.Time and truncates its serialisation, deserialisation of EpochMillisecond
// values will not result in identical time.Time values. For example:
//
//	func TestEpochMillisecond_TruncateNanosecond(t *testing.T) {
//		v, err := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999Z")
//		if err != nil {
//			t.Error(err)
//		}
//
//		data, err := json.Marshal(EpochMillisecond(v))
//		if err != nil {
//			t.Error(err)
//		}
//
//		got := EpochMillisecond(time.Time{})
//		if err := json.Unmarshal(data, &got); err != nil {
//			t.Error(err)
//		}
//
//		// got's underlying time.time is truncated to 2006-01-02T15:04:05.
//		if reflect.DeepEqual(got.ToTime(), v) {
//			t.Errorf("shouldn't be equal; got %v, want %v", got, v)
//		}
//
//		// if we reset v's nano time, then they are equal.
//		v = time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), got.ToTime().Nanosecond(), v.Location())
//		if !reflect.DeepEqual(got.ToTime(), v) {
//			t.Errorf("got %#v, want %#v", got.ToTime(), v)
//		}
//	}
type EpochMillisecond time.Time

// ToTime returns the underlying time.Time instance.
func (e *EpochMillisecond) ToTime() time.Time {
	if e == nil {
		return time.Time{}
	}

	return time.Time(*e)
}

// String implements the fmt.Stringer interface.
func (e EpochMillisecond) String() string {
	return strconv.FormatInt(e.ToTime().UnixMilli(), 10)
}

var _ json.Marshaler = &EpochMillisecond{}
var _ json.Marshaler = (*EpochMillisecond)(nil)
var _ json.Unmarshaler = &EpochMillisecond{}
var _ json.Unmarshaler = (*EpochMillisecond)(nil)
var _ attributevalue.Marshaler = &EpochMillisecond{}
var _ attributevalue.Marshaler = (*EpochMillisecond)(nil)
var _ attributevalue.Unmarshaler = &EpochMillisecond{}
var _ attributevalue.Unmarshaler = (*EpochMillisecond)(nil)

// MarshalJSON must not use receiver pointer to allow both pointer and non-pointer usage.
func (e EpochMillisecond) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(e.ToTime().UnixMilli())
	if err != nil {
		return nil, fmt.Errorf("epoch milli marshal JSON error: %w", err)
	}
	return data, nil
}

func (e *EpochMillisecond) UnmarshalJSON(data []byte) error {
	var number json.Number
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if err := d.Decode(&number); err != nil {
		return fmt.Errorf("epoch milli unmarshal JSON error: not a number: %w", err)
	}
	v, err := number.Int64()
	if err != nil {
		return fmt.Errorf("epoch milli unmarshal JSON error: not an int64: %w", err)
	}

	*e = EpochMillisecond(time.UnixMilli(v).UTC())
	return nil

}

// MarshalDynamoDBAttributeValue must not use receiver pointer to allow both pointer and non-pointer usage.
func (e EpochMillisecond) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberN{Value: e.String()}, nil
}

func (e *EpochMillisecond) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	avN, ok := av.(*types.AttributeValueMemberN)
	if !ok {
		return fmt.Errorf("epoch milli unmarshal DDB AV error: not type N")
	}

	n := avN.Value
	if n == "" {
		return fmt.Errorf("epoch milli unmarshal DDB AV error: empty N value")
	}

	v, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return fmt.Errorf("epoch milli unmarshal DDB AV error: not an int64: %w", err)
	}

	*e = EpochMillisecond(time.UnixMilli(v).UTC())
	return nil
}

// ToAttributeValueMap is convenient method to implement [.model.HasCreatedTimestamp] or [.model.HasModifiedTimestamp].
func (e EpochMillisecond) ToAttributeValueMap(key string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{key: &types.AttributeValueMemberN{Value: e.String()}}
}

// After is convenient method to [time.Time.After].
func (e EpochMillisecond) After(other EpochMillisecond) bool {
	return time.Time(e).After(other.ToTime())
}

// Before is convenient method to [time.Time.Before].
func (e EpochMillisecond) Before(other EpochMillisecond) bool {
	return time.Time(e).Before(other.ToTime())
}

// Equal is convenient method to [time.Time.Equal].
func (e EpochMillisecond) Equal(other EpochMillisecond) bool {
	return time.Time(e).Equal(other.ToTime())
}

// Compare is convenient method to [time.Time.Compare].
func (e EpochMillisecond) Compare(other EpochMillisecond) int {
	return time.Time(e).Compare(other.ToTime())
}

// Format is convenient method to [time.Time.Format].
func (e EpochMillisecond) Format(layout string) string {
	return time.Time(e).Format(layout)
}

// IsZero is convenient method to [time.Time.IsZero].
func (e EpochMillisecond) IsZero() bool {
	return time.Time(e).IsZero()
}
