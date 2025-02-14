package timestamp

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/nguyengg/golambda/must"
)

type AttributeValueItem struct {
	Day              Day              `dynamodbav:"day"`
	Timestamp        Timestamp        `dynamodbav:"timestamp"`
	EpochMillisecond EpochMillisecond `dynamodbav:"epochMillisecond"`
	EpochSecond      EpochSecond      `dynamodbav:"epochSecond"`
}

// TestAttributeValue_structUsage tests using all the timestamps in a struct.
func TestAttributeValue_structUsage(t *testing.T) {
	millisecond := Must(time.Parse(time.RFC3339, "2006-01-02T15:04:05.999Z"))
	second := Must(time.Parse(time.RFC3339, "2006-01-02T15:04:05Z"))

	item := AttributeValueItem{
		Day:              TruncateToStartOfDay(millisecond),
		Timestamp:        Timestamp(millisecond),
		EpochMillisecond: EpochMillisecond(millisecond),
		EpochSecond:      EpochSecond(second),
	}

	want := map[string]dynamodbtypes.AttributeValue{
		"day":              &dynamodbtypes.AttributeValueMemberS{Value: "2006-01-02"},
		"timestamp":        &dynamodbtypes.AttributeValueMemberS{Value: "2006-01-02T15:04:05.999Z"},
		"epochMillisecond": &dynamodbtypes.AttributeValueMemberN{Value: "1136214245999"},
		"epochSecond":      &dynamodbtypes.AttributeValueMemberN{Value: "1136214245"},
	}

	// non-pointer version.
	if got := Must(attributevalue.MarshalMap(item)); !reflect.DeepEqual(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
	// pointer version.
	if got := Must(attributevalue.MarshalMap(&item)); !reflect.DeepEqual(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}

	got := AttributeValueItem{}
	if Must0(attributevalue.UnmarshalMap(want, &got)); !reflect.DeepEqual(got, item) {
		t.Errorf("got %#v, want %#v", got, item)
	}
}
