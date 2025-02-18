package session

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

type Session struct {
	Id   string `dynamodbav:"id,hashkey" tableName:"session"`
	User *User  `dynamodbav:"user"`
}

func (s Session) GetUser() *User {
	return s.User
}

type User struct {
	Groups []string `dynamodbav:"groups,stringset" tableName:"session"`
}

func (u User) GetGroups() Groups {
	return u.Groups
}

type MockGetItemClient struct {
	mock.Mock
}

func (t *MockGetItemClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := t.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func TestMiddleware_RequireAuthorisation_StatusOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &MockGetItemClient{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "a.com", nil)
	c.Request.Header.Add("Cookie", "sid=my-session-id")

	client.
		On("GetItem", c, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			assert.Equal(t, "my-session-id", input.Key["id"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "session", *input.TableName)
			return true
		}), mock.FunctionalOptions()).
		Return(&dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "my-session-id"},
			"user": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"groups": &types.AttributeValueMemberSS{Value: []string{"a", "b"}},
			}},
		}}, nil)

	New[Session, User](client).RequireAuthorisation(OneOf("a", "b", "c"))(c)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	u, ok := c.Get("session.user")
	assert.True(t, ok)
	assert.Equal(t, &User{Groups: []string{"a", "b"}}, u)

	s, ok := c.Get("session")
	assert.True(t, ok)
	assert.Equal(t, &Session{Id: "my-session-id", User: u.(*User)}, s)

}

func TestMiddleware_RequireAuthorisation_StatusForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &MockGetItemClient{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "a.com", nil)
	c.Request.Header.Add("Cookie", "sid=my-session-id")

	client.
		On("GetItem", c, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			assert.Equal(t, "my-session-id", input.Key["id"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "session", *input.TableName)
			return true
		}), mock.FunctionalOptions()).
		Return(&dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "my-session-id"},
			"user": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"groups": &types.AttributeValueMemberSS{Value: []string{"a", "b"}},
			}},
		}}, nil)

	// differs from TestMiddleware_RequireAuthorisation_StatusOK here.
	// here, requiring user to belong to ALL a, b, and c. but because user only belongs to a and b, StatusForbidden
	// is returned.
	New[Session, User](client).RequireAuthorisation(AllOf("a", "b", "c"))(c)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)

	u, ok := c.Get("session.user")
	assert.True(t, ok)
	assert.Equal(t, &User{Groups: []string{"a", "b"}}, u)

	s, ok := c.Get("session")
	assert.True(t, ok)
	assert.Equal(t, &Session{Id: "my-session-id", User: u.(*User)}, s)
}

func TestMiddleware_RequireAuthorisation_StatusUnauthorizedNoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &MockGetItemClient{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "a.com", nil)
	c.Request.Header.Add("Cookie", "sid=my-session-id")

	client.
		On("GetItem", c, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			assert.Equal(t, "my-session-id", input.Key["id"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "session", *input.TableName)
			return true
		}), mock.FunctionalOptions()).
		Return(&dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "my-session-id"},
			// differs from TestMiddleware_RequireAuthorisation_StatusOK here. because there is no user,
			// the session is not authenticated.
		}}, nil)

	New[Session, User](client).RequireAuthorisation(AllOf("a", "b", "c"))(c)

	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	_, ok := c.Get("session.user")
	assert.False(t, ok)

	s, ok := c.Get("session")
	assert.True(t, ok)
	assert.Equal(t, &Session{Id: "my-session-id", User: nil}, s)
}

func TestMiddleware_RequireAuthorisation_StatusUnauthorizedNoSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &MockGetItemClient{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "a.com", nil)
	c.Request.Header.Add("Cookie", "sid=my-session-id")

	client.
		On("GetItem", c, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
			assert.Equal(t, "my-session-id", input.Key["id"].(*types.AttributeValueMemberS).Value)
			assert.Equal(t, "session", *input.TableName)
			return true
		}), mock.FunctionalOptions()).
		Return(&dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{}}, nil)

	New[Session, User](client).RequireAuthorisation(AllOf("a", "b", "c"))(c)

	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	_, ok := c.Get("session.user")
	assert.False(t, ok)

	_, ok = c.Get("session")
	assert.False(t, ok)
}

func TestMiddleware_RequireAuthorisation_StatusUnauthorizedNoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "a.com", nil)

	New[Session, User](nil).RequireAuthorisation(AllOf("a", "b", "c"))(c)

	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	_, ok := c.Get("session.user")
	assert.False(t, ok)

	_, ok = c.Get("session")
	assert.False(t, ok)
}
