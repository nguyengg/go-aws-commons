package timestamp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Timestamp is a UTC timestamp formatted and marshalled as a string using FractionalSecondLayout ("2006-01-02T15:04:05.000Z") layout.
//
// Because Timestamp wraps around time.Time and truncates its serialisation, deserialisation of Timestamp values will
// not result in identical time.Time values. For example:
//
//	func TestTimestamp_TruncateNanosecond(t *testing.T) {
//		v, err := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999Z")
//		if err != nil {
//			t.Error(err)
//		}
//
//		data, err := json.Marshal(Timestamp(v))
//		if err != nil {
//			t.Error(err)
//		}
//
//		got := Timestamp(time.Time{})
//		if err := json.Unmarshal(data, &got); err != nil {
//			t.Error(err)
//		}
//
//		// got's underlying time.time is truncated to 2006-01-02T15:04:05.999.
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
type Timestamp time.Time

// FractionalSecondLayout is the string layout as well as DynamoDB string value of Timestamp.
const FractionalSecondLayout = "2006-01-02T15:04:05.000Z"

// ParseTimestamp creates an instance of Timestamp from parsing the specified string.
//
// If the string fails to be parsed using layout FractionalSecondLayout, a zero-value Timestamp is returned.
func ParseTimestamp(value string) (Timestamp, error) {
	t, err := time.Parse(FractionalSecondLayout, value)
	if err != nil {
		return Timestamp(time.Time{}), err
	}

	return Timestamp(t), nil
}

// ToTime returns the underlying time.Time instance.
func (t *Timestamp) ToTime() time.Time {
	return time.Time(*t)
}

// String implements the fmt.Stringer interface.
func (t Timestamp) String() string {
	return t.Format(FractionalSecondLayout)
}

var _ json.Marshaler = &Timestamp{}
var _ json.Marshaler = (*Timestamp)(nil)
var _ json.Unmarshaler = &Timestamp{}
var _ json.Unmarshaler = (*Timestamp)(nil)
var _ attributevalue.Marshaler = &Timestamp{}
var _ attributevalue.Marshaler = (*Timestamp)(nil)
var _ attributevalue.Unmarshaler = &Timestamp{}
var _ attributevalue.Unmarshaler = (*Timestamp)(nil)

// MarshalJSON must not use receiver pointer to allow both pointer and non-pointer usage.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(t.Format(FractionalSecondLayout))
	if err != nil {
		return nil, fmt.Errorf("timestamp marshal JSON error: %w", err)
	}
	return data, nil
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("timestamp unmarshal JSON error: not a string: %w", err)
	} else if v, err := time.Parse(FractionalSecondLayout, value); err != nil {
		return fmt.Errorf(`timestamp unmarshal JSON error: not a string in time layout "%s": %w`, FractionalSecondLayout, err)
	} else {
		*t = Timestamp(v)
	}
	return nil
}

// MarshalDynamoDBAttributeValue must not use receiver pointer to allow both pointer and non-pointer usage.
func (t Timestamp) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: t.String()}, nil
}

func (t *Timestamp) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	avS, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("timestamp unmarshal DDB AV error: not type S")
	}

	s := avS.Value
	if s == "" {
		return fmt.Errorf("timestamp unmarshal DDB AV error: empty S value")
	}

	v, err := time.Parse(FractionalSecondLayout, s)
	if err != nil {
		return fmt.Errorf(`timestamp unmarshal DDB AV error: not a string in time layout "%s": %w`, FractionalSecondLayout, err)
	}

	*t = Timestamp(v)
	return nil
}

// ToAttributeValueMap is convenient method to implement [.model.HasCreatedTimestamp] or [.model.HasModifiedTimestamp].
func (t Timestamp) ToAttributeValueMap(key string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{key: &types.AttributeValueMemberS{Value: t.String()}}
}

// After is convenient method to [time.Time.After].
func (t Timestamp) After(other Timestamp) bool {
	return time.Time(t).After(other.ToTime())
}

// Before is convenient method to [time.Time.Before].
func (t Timestamp) Before(other Timestamp) bool {
	return time.Time(t).Before(other.ToTime())
}

// Equal is convenient method to [time.Time.Equal].
func (t Timestamp) Equal(other Timestamp) bool {
	return time.Time(t).Equal(other.ToTime())
}

// Compare is convenient method to [time.Time.Compare].
func (t Timestamp) Compare(other Timestamp) int {
	return time.Time(t).Compare(other.ToTime())
}

// Format is convenient method to [time.Time.Format].
func (t Timestamp) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// IsZero is convenient method to [time.Time.IsZero].
func (t Timestamp) IsZero() bool {
	return time.Time(t).IsZero()
}
