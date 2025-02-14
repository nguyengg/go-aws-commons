package timestamp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Day is a timestamp that is truncated to start of day.
//
// Day is a struct to avoid the truncation pitfalls that other timestamps have.
type Day struct {
	v time.Time
}

// DayLayout is the string layout as well as DynamoDB string value of Day.
const DayLayout = "2006-01-02"

// TruncateToStartOfDay creates a new Day from the specified time.Time.
func TruncateToStartOfDay(t time.Time) Day {
	return Day{
		v: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()),
	}
}

// TodayInLocation creates a new Day with the specified location.
func TodayInLocation(loc *time.Location) Day {
	return TruncateToStartOfDay(time.Now().In(loc))
}

// ParseDay creates an instance of Day from parsing the specified string.
//
// If the string fails to be parsed using layout DayLayout, a zero-value Day is returned.
func ParseDay(value string) (Day, error) {
	t, err := time.Parse(DayLayout, value)
	if err != nil {
		return Day{}, err
	}

	return Day{v: t}, nil
}

// ToTime returns a copy of the underlying time.Time instance.
func (d Day) ToTime() time.Time {
	return d.v
}

// String implements the fmt.Stringer interface.
func (d Day) String() string {
	return d.Format(DayLayout)
}

var _ json.Marshaler = &Day{}
var _ json.Marshaler = (*Day)(nil)
var _ json.Unmarshaler = &Day{}
var _ json.Unmarshaler = (*Day)(nil)
var _ attributevalue.Marshaler = &Day{}
var _ attributevalue.Marshaler = (*Day)(nil)
var _ attributevalue.Unmarshaler = &Day{}
var _ attributevalue.Unmarshaler = (*Day)(nil)

// MarshalJSON must not use receiver pointer to allow both pointer and non-pointer usage.
func (d Day) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(d.Format(DayLayout))
	if err != nil {
		return nil, fmt.Errorf("day marshal JSON error: %w", err)
	}
	return data, nil
}

func (d *Day) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("day unmarshal JSON error: not a string: %w", err)
	} else if t, err := time.Parse(DayLayout, value); err != nil {
		return fmt.Errorf(`day unmarshal JSON error: not a string in time layout "%s": %w`, DayLayout, err)
	} else {
		*d = Day{v: t}
	}
	return nil
}

// MarshalDynamoDBAttributeValue must not use receiver pointer to allow both pointer and non-pointer usage.
func (d Day) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: d.String()}, nil
}

func (d *Day) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	avS, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("day unmarshal DDB AV error: not type S")
	}

	s := avS.Value
	if s == "" {
		return fmt.Errorf("day unmarshal DDB AV error: empty S string")
	}

	t, err := time.Parse(DayLayout, s)
	if err != nil {
		return fmt.Errorf(`day unmarshal DDB AV error: not a string in time layout "%s": %w`, DayLayout, err)
	}

	*d = Day{v: t}
	return nil
}

// ToAttributeValueMap is convenient method to implement [.model.HasCreatedDay] or [.model.HasModifiedDay].
func (d Day) ToAttributeValueMap(key string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{key: &types.AttributeValueMemberS{Value: d.String()}}
}

// After is convenient method to [time.Time.After].
func (d Day) After(other Day) bool {
	return d.v.After(other.v)
}

// Before is convenient method to [time.Time.Before].
func (d Day) Before(other Day) bool {
	return d.v.Before(other.v)
}

// Equal is convenient method to [time.Time.Equal].
func (d Day) Equal(other Day) bool {
	return d.v.Equal(other.v)
}

// Compare is convenient method to [time.Time.Compare].
func (d Day) Compare(other Day) int {
	return d.v.Compare(other.v)
}

// Format is convenient method to [time.Time.Format].
func (d Day) Format(layout string) string {
	return d.v.Format(layout)
}

// IsZero is convenient method to [time.Time.IsZero].
func (d Day) IsZero() bool {
	return d.v.IsZero()
}

// In is convenient method to [time.Time.In].
func (d Day) In(loc *time.Location) Day {
	return Day{v: d.v.In(loc)}
}
