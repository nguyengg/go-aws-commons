package lambda

import (
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestStreamToDynamoDBItem_success(t *testing.T) {
	type args struct {
		item map[string]events.DynamoDBAttributeValue
	}
	tests := []struct {
		name string
		args args
		want map[string]dynamodbtypes.AttributeValue
	}{
		{
			name: "Basic test",
			args: args{
				item: map[string]events.DynamoDBAttributeValue{
					"version":   events.NewNumberAttribute("123"),
					"hello":     events.NewStringAttribute("world"),
					"numberSet": events.NewNumberSetAttribute([]string{"45", "67"}),
					"stringSet": events.NewStringSetAttribute([]string{"hello", "world"}),
					"list": events.NewListAttribute([]events.DynamoDBAttributeValue{
						events.NewNumberAttribute("12"),
						events.NewStringAttribute("34"),
					}),
				},
			},
			want: map[string]dynamodbtypes.AttributeValue{
				"version":   &dynamodbtypes.AttributeValueMemberN{Value: "123"},
				"hello":     &dynamodbtypes.AttributeValueMemberS{Value: "world"},
				"numberSet": &dynamodbtypes.AttributeValueMemberNS{Value: []string{"45", "67"}},
				"stringSet": &dynamodbtypes.AttributeValueMemberSS{Value: []string{"hello", "world"}},
				"list": &dynamodbtypes.AttributeValueMemberL{Value: []dynamodbtypes.AttributeValue{
					&dynamodbtypes.AttributeValueMemberN{Value: "12"},
					&dynamodbtypes.AttributeValueMemberS{Value: "34"},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StreamToDynamoDBItem(tt.args.item); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StreamToDynamoDBItem() = %v, want %v", got, tt.want)
			}
		})
	}
}
