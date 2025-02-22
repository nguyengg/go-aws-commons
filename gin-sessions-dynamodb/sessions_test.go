package sessions

import (
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

func TestSessions_New(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	sid := "my-session-id"
	client := &MockDynamoDBClient{}

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
	client := &MockDynamoDBClient{}
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
