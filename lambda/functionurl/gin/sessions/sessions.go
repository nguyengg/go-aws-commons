package sessions

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/context"
	"github.com/nguyengg/go-aws-commons/ddb"
	"github.com/nguyengg/go-aws-commons/ddb/model"
)

const (
	contextKey = "github.com/go-aws-commons/lambda/functionurl/gin/sessions"
)

// Sessions returns a middleware for management sessions.
//
// Inspired by https://github.com/gin-contrib/sessions, this middleware provides type safety by requiring user to define
// the schema of the session struct (via T) using ddb.Builder struct tags. The name argument is the name of the cookie
// that contains the session Id. Default or Get can be used to retrieve a pointer to T associated with a request.
//
// Type T must define a `string` hashkey. New sessions will use NewURLSafeId as the session Id.
func Sessions[T interface{}](client ddb.ManagerAPIClient, name string) (gin.HandlerFunc, error) {
	t := reflect.TypeFor[T]()
	m, err := model.ParseFromType(t)
	if err != nil {
		return nil, err
	}
	if m.HashKey == nil {
		return nil, fmt.Errorf(`no hashkey field in type "%s"`, t.Name())
	} else if m.HashKey.DataType != model.DataTypeS {
		return nil, fmt.Errorf("hashkey type (%s) does not encode to type S", t.Name())
	}

	return func(c *gin.Context) {
		s := &store{
			name:    name,
			manager: ddb.NewManager(client),
			m:       m,
			t:       t,
		}
		c.Set(contextKey, s)
		defer context.Clear(c.Request)
		c.Next()
	}, nil
}

// MustSessions is the panicky variant of Sessions.
func MustSessions[T interface{}](client ddb.ManagerAPIClient, name string) gin.HandlerFunc {
	f, err := Sessions[T](client, name)
	if err != nil {
		panic(err)
	}

	return f
}

// Default returns pointer to the session struct associated with the given request, or create a new one if none exists.
//
// Usage:
//
//	var s *SessionType = sessions.Default(c).(*SessionType)
func Default(c *gin.Context) interface{} {
	return c.MustGet(contextKey).(*store).get(c)
}

// Get returns pointer to the session struct associated with the given request, or create a new one if none exists.
//
// Usage:
//
//	var s *SessionType = sessions.Get[SessionType](c)
func Get[T interface{}](c *gin.Context) *T {
	return c.MustGet(contextKey).(*store).get(c).(*T)
}

// New always creates a new session and return the pointer thereto.
//
// Usage:
//
//	var s *SessionType = sessions.New[SessionType](c)
func New[T interface{}](c *gin.Context) *T {
	return c.MustGet(contextKey).(*store).new().(*T)
}

// SaveOptions customises how Save writes the session cookie.
//
// Intentionally following github.com/gorilla/sessions here.
type SaveOptions struct {
	Path   string
	Domain string
	// MaxAge=0 means no Max-Age attribute specified and the cookie will be
	// deleted after the browser session ends.
	// MaxAge<0 means delete cookie immediately.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int
	Secure   bool
	HttpOnly bool
	// Defaults to http.SameSiteDefaultMode
	SameSite http.SameSite
}

// Save saves the current session to DynamoDB.
//
// Usage:
//
//	sessions.Save(c)
func Save(c *gin.Context, optFns ...func(*SaveOptions)) {
	opts := &SaveOptions{
		SameSite: http.SameSiteDefaultMode,
	}
	for _, fn := range optFns {
		fn(opts)
	}

	c.MustGet(contextKey).(*store).save(c, opts)
}

// NewURLSafeId creates a new UUID and returns its raw-URL-encoded content.
func NewURLSafeId() string {
	data := uuid.New()
	return base64.RawURLEncoding.EncodeToString(data[:])
}
