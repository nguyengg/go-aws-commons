package sessions

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb"
)

// Session implements gin sessions.Session in a type-safe way.
type Session struct {
	// Client is the DynamoDB client for saving session data.
	//
	// By default, `config.LoadDefaultConfig` will be used to provide an instance.
	Client ddb.ManagerAPIClient

	// ClientOptions is passed to every DynamoDB call.
	ClientOptions []func(*dynamodb.Options)

	// NewSessionId is used to create the Id for a new session.
	//
	// By default, DefaultNewSessionId is used.
	NewSessionId func() string

	// CookieOptions modify the cookie settings.
	CookieOptions sessions.Options

	c       *gin.Context
	name    string
	table   *ddb.Table
	manager *ddb.Manager
	t       reflect.Type
	v       interface{}
}

func (s *Session) get() (interface{}, error) {
	if s.v != nil {
		return s.v, nil
	}

	sid, err := s.c.Cookie(s.name)
	if errors.Is(err, http.ErrNoCookie) {
		return s.new()
	}

	v := reflect.New(s.t)
	s.v = v.Interface()

	if v, err = s.table.HashKey.GetFieldValue(v.Elem()); err != nil {
		return nil, fmt.Errorf("get field value error: %w", err)
	} else {
		v.SetString(sid)
	}

	if _, err = s.manager.Get(s.c, s.v, s.v); err != nil {
		return nil, fmt.Errorf("get session from dynamodb error: %w", err)
	}

	return s.v, nil
}

func (s *Session) new() (interface{}, error) {
	v := reflect.New(s.t)
	s.v = v.Interface()

	if v, err := s.table.HashKey.GetFieldValue(v.Elem()); err != nil {
		return nil, fmt.Errorf("get field value error: %w", err)
	} else {
		v.SetString(s.NewSessionId())
	}

	return s.v, nil
}

var _ sessions.Session = &Session{}
var _ sessions.Session = (*Session)(nil)

func (s *Session) ID() string {
	v, err := s.table.HashKey.GetFieldValue(reflect.ValueOf(s.v))
	if err != nil {
		panic(fmt.Errorf("get field value error: %w", err))
	}

	return v.String()
}

func (s *Session) Get(key interface{}) interface{} {
	switch key.(type) {
	case string:
		return s.table.MustGet(s.v, key.(string))
	default:
		panic(fmt.Errorf("unsupported key type %T; only strings are supported", key))
	}
}

func (s *Session) Set(key interface{}, val interface{}) {
	switch key.(type) {
	case string:
		a, ok := s.table.All[key.(string)]
		if !ok {
			panic(fmt.Errorf(`session type %s has no attribute with name "%s"`, s.t, key))
		}

		v, err := a.GetFieldValue(reflect.ValueOf(s.v))
		if err != nil {
			panic(fmt.Errorf("get field value error: %w", err))
		}

		v.Set(reflect.ValueOf(val))
	default:
		panic(fmt.Errorf("unsupported key type %T; only strings are supported", key))
	}
}

func (s *Session) Delete(key interface{}) {
	switch key.(type) {
	case string:
		a, ok := s.table.All[key.(string)]
		if !ok {
			panic(fmt.Errorf(`session type %s has no attribute with name "%s"`, s.t, key))
		}

		v, err := a.GetFieldValue(reflect.ValueOf(s.v))
		if err != nil {
			panic(fmt.Errorf("get field value error: %w", err))
		}

		v.Set(reflect.Zero(v.Type()))
	default:
		panic(fmt.Errorf("unsupported key type %T; only strings are supported", key))
	}
}

// Clear deletes all values in the session.
//
// The hashkey (session Id) will not be deleted, and any fields not tagged with `dynamodbav` will also be ignored.
func (s *Session) Clear() {
	for _, a := range s.table.All {
		if a == s.table.HashKey {
			continue
		}

		v, err := a.GetFieldValue(reflect.ValueOf(s.v))
		if err != nil {
			panic(fmt.Errorf("get field value error: %w", err))
		}

		v.Set(reflect.Zero(v.Type()))
	}
}

func (s *Session) AddFlash(value interface{}, vars ...string) {
	//TODO implement me
	panic("implement me")
}

func (s *Session) Flashes(vars ...string) []interface{} {
	//TODO implement me
	panic("implement me")
}

func (s *Session) Options(options sessions.Options) {
	s.CookieOptions = options
}

func (s *Session) Save() error {
	// if the session value is nil then that means there has not been any changes to the session, so skip the save.
	if s.v == nil {
		return nil
	}

	if _, err := s.manager.Put(s.c, s.v); err != nil {
		return err
	}

	s.v = reflect.New(s.t)
	v, err := s.table.HashKey.GetFieldValue(reflect.ValueOf(s.v))
	if err != nil {
		return err
	}

	s.c.SetSameSite(s.CookieOptions.SameSite)
	s.c.SetCookie(s.name, v.String(), s.CookieOptions.MaxAge, s.CookieOptions.Path, s.CookieOptions.Domain, s.CookieOptions.Secure, s.CookieOptions.HttpOnly)
	return nil
}
