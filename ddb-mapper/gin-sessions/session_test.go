package sessions

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestSession struct {
	SessionId string `dynamodbav:"sessionId,hashkey" tableName:"session"`
	User      string `dynamodbav:"user"`
	Version   int64  `dynamodbav:"version,version"` // making use of ddb's optimistic locking.
}

func (s *TestSession) marshal(t *testing.T) map[string]dynamodbtypes.AttributeValue {
	item, err := attributevalue.MarshalMap(s)
	assert.NoErrorf(t, err, "TestSession.marshal(%#v) error: %v", s, err)
	return item
}

func (s *TestSession) unmarshal(t *testing.T, item map[string]dynamodbtypes.AttributeValue) {
	newSession := TestSession{}
	err := attributevalue.UnmarshalMap(item, &newSession)
	assert.NoErrorf(t, err, "TestSession.unmarshal(%#v) error: %v", item, err)
	*s = newSession
}

type MockManagerAPIClient struct {
	mock.Mock
	dynamodb.Client
}

func (m *MockManagerAPIClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockManagerAPIClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockManagerAPIClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

// mockGetItem mocks GetItem to receive the given sid as primary key, and return given TestSession marshaled.
//
// If result is nil, the response's Item attribute is effectively empty.
func (m *MockManagerAPIClient) mockGetItem(t *testing.T, sid string, out *TestSession) {
	var item = make(map[string]dynamodbtypes.AttributeValue)
	if out != nil {
		item = out.marshal(t)
	}

	m.
		On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			assert.Equalf(t, *input.TableName, "session", "expected tableName=%s, got %s", "session", *input.TableName)

			// the key should contain only the pk.
			expectedKey := &TestSession{SessionId: sid}
			actualKey := &TestSession{}
			actualKey.unmarshal(t, input.Key)
			assert.Equalf(t, expectedKey, actualKey, "expected=%#v, got=%#v", expectedKey, actualKey)
			return true
		}), mock.FunctionalOptions()).
		Return(&dynamodb.GetItemOutput{Item: item}, nil).
		Once()
}

// mockPutItem mocks PutItem to expect PutItemInput to contain s, and return out if given.
//
// If out is not given, the response will be in with its version increased by 1.
func (m *MockManagerAPIClient) mockPutItem(t *testing.T, in *TestSession, out *TestSession) {
	if out == nil {
		out = &(*in)
	}

	m.
		On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
			assert.Equalf(t, *input.TableName, "session", "expected tableName=%s, got %s", "session", *input.TableName)

			item := &TestSession{}
			item.unmarshal(t, input.Item)
			assert.Equalf(t, in, item, "expected=%#v, got=%#v", in, item)

			// this actually tests the "version" logic.
			assert.Equal(t, map[string]string{"#0": "version"}, input.ExpressionAttributeNames)
			assert.Equal(t, map[string]dynamodbtypes.AttributeValue{":0": &dynamodbtypes.AttributeValueMemberN{Value: "1"}}, input.ExpressionAttributeValues)
			assert.Equal(t, "#0 = :0", *input.ConditionExpression)

			return true
		}), mock.FunctionalOptions()).
		Return(&dynamodb.PutItemOutput{Attributes: out.marshal(t)}, nil).
		Once()
}

// mockDestroyItem mocks DeleteItem.
func (m *MockManagerAPIClient) mockDestroyItem(t *testing.T, sid string) {
	m.
		On("DeleteItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.DeleteItemInput) bool {
			assert.Equalf(t, *input.TableName, "session", "expected tableName=%s, got %s", "session", *input.TableName)

			// the key should contain only the pk.
			expectedKey := &TestSession{SessionId: sid}
			actualKey := &TestSession{}
			actualKey.unmarshal(t, input.Key)
			assert.Equalf(t, expectedKey, actualKey, "expected=%#v, got=%#v", expectedKey, actualKey)
			return true
		}), mock.FunctionalOptions()).
		Return(&dynamodb.DeleteItemOutput{}, nil).
		Once()
}
