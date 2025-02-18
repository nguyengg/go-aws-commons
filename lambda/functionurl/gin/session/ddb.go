package session

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDBGetItemClient abstracts the API needed by FromDynamoDB.
type DynamoDBGetItemClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*Options)) (*dynamodb.GetItemOutput, error)
}

// FromDynamoDB provides a way to retrieve session from DynamoDB.
func FromDynamoDB[SessionType interface{}](client DynamoDBGetItemClient) func(context.Context, string) (*SessionType, error) {
	return func(ctx context.Context, sessionId string) (v *SessionType, _ error) {
		//TODO implement me
		panic("implement me")
	}
}
