package timestamp

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func Must[H any](value H, err error) H {
	if err != nil {
		panic(err)
	}

	return value
}

type AttributeValueItem struct {
	Day              Day              `dynamodbav:"day"`
	Timestamp        Timestamp        `dynamodbav:"timestamp"`
	EpochMillisecond EpochMillisecond `dynamodbav:"epochMillisecond"`
	EpochSecond      EpochSecond      `dynamodbav:"epochSecond"`
}

// TestAttributeValue_structUsage tests using all the timestamps in a struct.
func TestAttributeValue_structUsage(t *testing.T) {
	millisecond, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05.999Z")
	assert.NoError(t, err)

	second, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	assert.NoError(t, err)

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
	got, err := attributevalue.MarshalMap(item)
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	// pointer version.
	got, err = attributevalue.MarshalMap(&item)
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	newItem := AttributeValueItem{}
	err = attributevalue.UnmarshalMap(want, &newItem)
	assert.NoError(t, err)
	assert.Equal(t, item, newItem)
}
