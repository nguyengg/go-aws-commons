package sessions

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/context"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
	"github.com/nguyengg/go-aws-commons/ddb"
)

const (
	// DefaultKey is the gin context key for Session instance.
	DefaultKey = "github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb"
)

// Sessions creates a gin middleware for managing sessions of struct type T.
//
// The name argument is the name of the cookie that stores the session Id. Type T must have these struct tags:
//
//	// Hash key is required, and its type must be a string since only string session Ids are supported.
//	Field string `dynamodbav:"sessionId,hashkey" tableName:"my-table"`
//
// See ddb.Table for more information on how the struct tags are parsed. If type T does not implement the required tags
// or the tags fail validation, the function will panic.
//
// Use WithCSRF if you want Save to also create a new CSRF token if the session is new.
func Sessions[T interface{}](name string, optFns ...func(*Session)) gin.HandlerFunc {
	table, err := ddb.NewTable(reflect.TypeFor[T](), func(options *ddb.TableOptions) {
		options.Validator = validator
	})
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		s := &Session{
			NewSessionId: DefaultNewSessionId,
			CookieOptions: CookieOptions{
				MaxAge:   0,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteDefaultMode,
			},
			c:     c,
			name:  name,
			table: table,
			t:     reflect.TypeFor[T](),
		}
		for _, fn := range optFns {
			fn(s)
		}
		if s.Client == nil {
			cfg, err := configcache.Get(c)
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			s.Client = dynamodb.NewFromConfig(cfg)
		}

		s.manager = ddb.NewManager(s.Client)

		c.Set(DefaultKey, s)
		defer context.Clear(c.Request)
		c.Next()
	}
}

// Default returns the Session instance attached to the request.
//
// There are two ways to interact with the session middleware; this is one of them by letting you interact with the
// Session wrapper.
func Default(c *gin.Context) *Session {
	s := c.MustGet(DefaultKey).(*Session)
	if _, err := s.get(); err != nil {
		panic(err)
	}

	return s
}

// Get returns the pointer to the session struct attached to the request.
//
// There are two ways to interact with the session middleware; this is the more type-safe way.
//
// Usage:
//
//	type MySession struct {
//		Id string `dynamodbav:"sessionId,hashkey" tableName:"session"`
//	}
//
//	r := gin.Default()
//	r.Use(Sessions[MySession]("sid"))
//	r.GET("/", func (c *gin.Context) {
//		var s *MySession = Get[MySession](c)
//	})
func Get[T interface{}](c *gin.Context) *T {
	v, err := c.MustGet(DefaultKey).(*Session).get()
	if err != nil {
		panic(err)
	}

	return v.(*T)
}

// New always create a new session and return the pointer thereto.
//
// Usage:
//
//	type MySession struct {
//		Id string `dynamodbav:"sessionId,hashkey" tableName:"session"`
//	}
//
//	r := gin.Default()
//	r.Use(Sessions[MySession]("sid"))
//	r.GET("/", func (c *gin.Context) {
//		var s *MySession = New[MySession](c)
//	})
func New[T interface{}](c *gin.Context) *T {
	v, err := c.MustGet(DefaultKey).(*Session).new()
	if err != nil {
		panic(err)
	}

	return v.(*T)
}

// Save can be used to save the current session to DynamoDB.
//
// If you are not using Default and only use the type-safe Get and New, Save can be used instead of Session.Save.
func Save(c *gin.Context) error {
	return c.MustGet(DefaultKey).(*Session).Save()
}

// SetCookieOptions can be used to modify the cookie options for the current session.
//
// If you are not using Default and only use the type-safe Get and New, SetCookieOptions can be used instead of
// Session.Options.
func SetCookieOptions(c *gin.Context, options CookieOptions) {
	c.MustGet(DefaultKey).(*Session).Options(options)
}

// DefaultNewSessionId creates a new UUID and returns its raw-URL-encoded content.
func DefaultNewSessionId() string {
	data := uuid.New()
	return base64.RawURLEncoding.EncodeToString(data[:])
}

func validator(a *ddb.Attribute) error {
	// hashkey must be string.
	tags := strings.Split(a.Tag.Get("dynamodbav"), ",")
	if len(tags) > 1 {
		switch tags[1] {
		case "hashkey":
			if a.Type.Kind() != reflect.String {
				return fmt.Errorf("hashkey of type %s is not supported; must be string", a.Type)
			}
		}
	}

	return nil
}

// CookieOptions customises the session and/or CSRF cookie.
//
// Fields are a subset of http.Cookie fields.
//
// This is a clone from "github.com/gin-contrib/sessions" and "github.com/gorilla/sessions" which are both named
// "sessions" to help avoid import naming conflicts. Additionally, Expires is supported.
type CookieOptions struct {
	Path    string
	Domain  string
	Expires time.Time
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}
