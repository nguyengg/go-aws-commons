package sessions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestSession struct {
	SessionId string `dynamodbav:"sessionId,hashkey" tableName:"session"`
}

type MockManagerAPIClient struct {
	mock.Mock
	dynamodb.Client
}

func (m *MockManagerAPIClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func (m *MockManagerAPIClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockManagerAPIClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockManagerAPIClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

func TestSessions_New(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	sid := "my-session-id"
	client := &MockManagerAPIClient{}

	r.GET("/", Sessions[TestSession]("my-session", func(s *Session) {
		s.Client = client
		s.NewSessionId = func() string {
			return sid
		}
	}), func(c *gin.Context) {
		v := New[TestSession](c)
		assert.Equal(t, v, Get[TestSession](c))
		assert.Equal(t, v, &TestSession{SessionId: sid})
	})
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, c.Request)
}

func TestSessions_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	sid := "my-session-id"
	client := &MockManagerAPIClient{}
	client.
		On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			assert.Equal(t, sid, input.Key["sessionId"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, *input.TableName, "session")
			return true
		}), mock.FunctionalOptions()).
		Return(&dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{"sessionId": &types.AttributeValueMemberS{Value: sid}},
		}, nil)

	r.GET("/", Sessions[TestSession]("my-session", func(s *Session) {
		s.Client = client
	}), func(c *gin.Context) {
		v := Get[TestSession](c)
		assert.Equal(t, v, &TestSession{SessionId: sid})
	})
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Add("Cookie", "my-session=my-session-id")
	r.ServeHTTP(w, c.Request)
}
