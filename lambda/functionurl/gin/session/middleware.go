package session

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb"
)

// HasUser is the interface for the GetUser method which can be used to return information about a user attached with a
// session.
type HasUser[UserType interface{}] interface {
	GetUser() *UserType
}

// HasGroups is the interface for the GetGroups method which returns all the groups that the user belongs to.
//
// The return value is intended to be used with Groups.Test to test for membership.
type HasGroups interface {
	GetGroups() Groups
}

// MiddlewareAPIClient abstracts the API needed by Middleware.
type MiddlewareAPIClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

// Middleware is a gin middleware that can be used to retrieve session data from DynamoDB.
//
// The zero-value is ready for use.
type Middleware[SessionType HasUser[UserType], UserType HasGroups] struct {
	// Client is the DynamoDB client for making GetItem calls.
	//
	// If not given, one will be created from `config.LoadDefaultConfig(...)`.
	Client MiddlewareAPIClient

	// CookieName is the name of the cookie to retrieve the session Id.
	//
	// The default value is "sid".
	CookieName string

	// SessionContextKey is the key to attach a valid session with a given [gin.Context] via [gin.Context.Set].
	//
	// The default value is "session". Pass empty string to disable this feature.
	SessionContextKey string

	// UserContextKey is the key to attach a valid user with a given [gin.Context] via [gin.Context.Set].
	//
	// The default value is "session.user". Pass empty string to disable this feature.
	UserContextKey string

	// ClientOptions is passed to each GetItem call.
	ClientOptions []func(*dynamodb.Options)
}

// New is a convenient function to create a Middleware instance and modifies it.
func New[SessionType HasUser[UserType], UserType HasGroups](client MiddlewareAPIClient, optFns ...func(*Middleware[SessionType, UserType])) *Middleware[SessionType, UserType] {
	m := &Middleware[SessionType, UserType]{
		Client:            client,
		CookieName:        "sid",
		SessionContextKey: "session",
		UserContextKey:    "session.user",
	}
	for _, fn := range optFns {
		fn(m)
	}

	return m
}

// getSession makes the GetItem to retrieve the session.
func (m *Middleware[SessionType, UserType]) getSession(ctx context.Context, sessionId string) (*SessionType, error) {
	key, tableName, err := ddb.CreateKey[SessionType](sessionId)
	if err != nil {
		return nil, err
	}

	getItemOutput, err := m.Client.GetItem(ctx, &dynamodb.GetItemInput{
		Key:       key,
		TableName: &tableName,
	}, m.ClientOptions...)
	if err != nil || len(getItemOutput.Item) == 0 {
		return nil, err
	}

	v := new(SessionType)
	if err = attributevalue.UnmarshalMap(getItemOutput.Item, &v); err != nil {
		return nil, err
	}

	return v, nil
}

// RequireSession adds a middleware that rejects all requests with http.StatusUnauthorized that don't have a valid
// session.
func (m *Middleware[SessionType, UserType]) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId, err := c.Cookie(m.CookieName)
		if errors.Is(err, http.ErrNoCookie) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		s, err := m.getSession(c, sessionId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if s == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if m.SessionContextKey != "" {
			c.Set(m.SessionContextKey, s)
		}

		c.Next()
	}
}

// RequireAuthentication adds a middleware that rejects all requests with http.StatusUnauthorized that don't have a
// valid session, or if the session is not authenticated (HasUser.GetUser returning a nil value).
func (m *Middleware[SessionType, UserType]) RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId, err := c.Cookie(m.CookieName)
		if errors.Is(err, http.ErrNoCookie) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		s, err := m.getSession(c, sessionId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if s == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if m.SessionContextKey != "" {
			c.Set(m.SessionContextKey, s)
		}

		user := (*s).GetUser()
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if m.UserContextKey != "" {
			c.Set(m.UserContextKey, user)
		}

		c.Next()
	}
}

// RequireAuthorisation implies RequireAuthentication while also rejects requests with http.StatusForbidden if the user
// returned by [HasUser.GetUser] does not have permission according to some set of rules.
func (m *Middleware[SessionType, UserType]) RequireAuthorisation(rule Rule, more ...Rule) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId, err := c.Cookie(m.CookieName)
		if errors.Is(err, http.ErrNoCookie) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		s, err := m.getSession(c, sessionId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if s == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if m.SessionContextKey != "" {
			c.Set(m.SessionContextKey, s)
		}

		user := (*s).GetUser()
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if m.UserContextKey != "" {
			c.Set(m.UserContextKey, user)
		}

		if !(*user).GetGroups().Test(rule, more...) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
