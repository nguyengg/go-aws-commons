package lambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/metrics"
)

// StartDynamodbEventHandler starts the Lambda loop for handling DynamoDB stream events in a batch.
//
// See https://docs.aws.amazon.com/lambda/latest/dg/services-ddb-batchfailurereporting.html.
//
// The handler works on one record at a time and sequentially. If the handler returns a non-nil error, the wrapper will
// automatically add an events.DynamoDBBatchItemFailure so that only failed records get retried later.
func StartDynamodbEventHandler(handler func(context.Context, events.DynamoDBEventRecord) error, options ...lambda.Option) {
	StartHandlerFunc(func(ctx context.Context, req events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
		m := metrics.Get(ctx)

		res := events.DynamoDBEventResponse{
			BatchItemFailures: make([]events.DynamoDBBatchItemFailure, 0),
		}

		m.AddCounter("recordCount", int64(len(req.Records)), "failureCount")

		for _, record := range req.Records {
			if err := handler(ctx, record); err != nil {
				m.AddCounter("failureCount", 1)
				res.BatchItemFailures = append(res.BatchItemFailures, events.DynamoDBBatchItemFailure{ItemIdentifier: record.Change.SequenceNumber})
			}
		}

		// very important that nil error is returned here.
		return res, nil
	}, options...)
}

// StreamToDynamoDBAttributeValue converts a DynamoDB Stream event attribute value (from
// https://pkg.go.dev/github.com/aws/aws-lambda-go/events) to an equivalent DynamoDB attribute value (from
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb/types).
//
// See StreamToDynamoDBItem for usage.
func StreamToDynamoDBAttributeValue(av events.DynamoDBAttributeValue) dynamodbtypes.AttributeValue {
	// TODO as an exercise, remove recursion.

	switch av.DataType() {
	case events.DataTypeBinary:
		return &dynamodbtypes.AttributeValueMemberB{Value: av.Binary()}
	case events.DataTypeBoolean:
		return &dynamodbtypes.AttributeValueMemberBOOL{Value: av.Boolean()}
	case events.DataTypeBinarySet:
		return &dynamodbtypes.AttributeValueMemberBS{Value: av.BinarySet()}
	case events.DataTypeList:
		l := av.List()
		value := make([]dynamodbtypes.AttributeValue, len(l))
		for i, v := range l {
			value[i] = StreamToDynamoDBAttributeValue(v)
		}
		return &dynamodbtypes.AttributeValueMemberL{Value: value}
	case events.DataTypeMap:
		value := make(map[string]dynamodbtypes.AttributeValue)
		for k, v := range av.Map() {
			value[k] = StreamToDynamoDBAttributeValue(v)
		}
		return &dynamodbtypes.AttributeValueMemberM{Value: value}
	case events.DataTypeNumber:
		return &dynamodbtypes.AttributeValueMemberN{Value: av.Number()}
	case events.DataTypeNumberSet:
		return &dynamodbtypes.AttributeValueMemberNS{Value: av.NumberSet()}
	case events.DataTypeNull:
		return &dynamodbtypes.AttributeValueMemberNULL{Value: av.IsNull()}
	case events.DataTypeString:
		return &dynamodbtypes.AttributeValueMemberS{Value: av.String()}
	case events.DataTypeStringSet:
		return &dynamodbtypes.AttributeValueMemberSS{Value: av.StringSet()}
	default:
		// should panic?
		return nil
	}
}

// StreamToDynamoDBItem uses StreamToDynamoDBAttributeValue to convert an item from a DynamoDB Stream event to an item
// in DynamoDB.
//
// Useful if you're implementing a DynamoDB Stream event handler, and you need to convert the old and/or new image to
// the tagged struct by way of https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue:
//
//	item := &MyStruct{}
//	err := attributevalue.UnmarshalMap(StreamToDynamoDBItem(record.Change.NewImage), item)
func StreamToDynamoDBItem(in map[string]events.DynamoDBAttributeValue) map[string]dynamodbtypes.AttributeValue {
	out := make(map[string]dynamodbtypes.AttributeValue)
	for k, v := range in {
		out[k] = StreamToDynamoDBAttributeValue(v)
	}
	return out
}
