package session

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBGetItemClient abstracts the API needed by FromDynamoDB.
type DynamoDBGetItemClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*Options)) (*dynamodb.GetItemOutput, error)
}

// FromDynamoDBRawItem is an implementation of GetSession that retrieves from DynamoDB.
//
// The given GetItem input parameters will have its key replaced with the session Id.
func FromDynamoDBRawItem(client DynamoDBGetItemClient, partitionKeyName string, input *dynamodb.GetItemInput) GetSession[map[string]types.AttributeValue] {
	return func(ctx context.Context, sessionId string) (map[string]types.AttributeValue, error) {
		input.Key = map[string]types.AttributeValue{partitionKeyName: &types.AttributeValueMemberS{Value: sessionId}}
		getItemOutput, err := client.GetItem(ctx, input)
		if err != nil {
			return nil, err
		}

		return getItemOutput.Item, nil
	}
}
