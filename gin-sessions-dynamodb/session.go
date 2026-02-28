package sessions

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb"
	"github.com/nguyengg/go-aws-commons/opaque-token/hmac"
)

// Session implements gin Session in a type-safe way.
type Session struct {
	// Client is the DynamoDB client for saving session data.
	//
	// By default, `configcache.Get` will be used to provide an instance.
	Client ddb.ManagerAPIClient

	// ClientOptions is passed to every DynamoDB call.
	ClientOptions []func(*dynamodb.Options)

	// NewSessionId is used to create the Id for a new session.
	//
	// By default, DefaultNewSessionId is used.
	NewSessionId func() string

	// CookieOptions modifies the session cookie settings.
	CookieOptions CookieOptions

	c       *gin.Context
	name    string
	table   *ddb.Table
	manager *ddb.Manager
	t       reflect.Type
	v       any

	// csrf stuff.
	hasher         hmac.Hasher
	csrfCookieName string
	csrfValue      string
}

func (s *Session) get() (any, error) {
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

	if s.hasher != nil {
		token, err := s.hasher.Sign(s.c, []byte(sid), 16)
		if err != nil {
			return nil, fmt.Errorf("create CSRF token error: %w", err)
		}

		s.csrfValue = base64.RawURLEncoding.EncodeToString(token)
	}

	return s.v, nil
}

func (s *Session) new() (any, error) {
	v := reflect.New(s.t)
	s.v = v.Interface()

	var sid string
	if v, err := s.table.HashKey.GetFieldValue(v.Elem()); err != nil {
		return nil, fmt.Errorf("get field value error: %w", err)
	} else {
		sid = s.NewSessionId()
		v.SetString(sid)
	}

	if s.hasher != nil {
		token, err := s.hasher.Sign(s.c, []byte(sid), 16)
		if err != nil {
			return nil, fmt.Errorf("create CSRF token error: %w", err)
		}

		s.csrfValue = base64.RawURLEncoding.EncodeToString(token)
	}

	return s.v, nil
}

// ID returns the session Id string.
func (s *Session) ID() string {
	v, err := s.table.HashKey.GetFieldValue(reflect.Indirect(reflect.ValueOf(s.v)))
	if err != nil {
		panic(fmt.Errorf("get field value error: %w", err))
	}

	return v.String()
}

// Get returns the attribute given the string key.
//
// Panics if key is not type string or key does is not tagged as a `dynamodbav` struct tag like:
//
//	type MySession struct {
//		Id    string `dynamodbav:"sessionId,hashkey" tableName:"session"`
//		Field string `dynamodbav:"key"`
//	}
func (s *Session) Get(key any) any {
	switch key.(type) {
	case string:
		return s.table.MustGet(s.v, key.(string))
	default:
		panic(fmt.Errorf("unsupported key type %T; only strings are supported", key))
	}
}

// Set modifies the attribute with the given string key.
//
// Panics if key is not type string, or key does is not tagged as a `dynamodbav` struct tag (like below), or the type of
// val argument is not assignable to the field in struct:
//
//	type MySession struct {
//		Id    string `dynamodbav:"sessionId,hashkey" tableName:"session"`
//		Field string `dynamodbav:"key"`
//	}
func (s *Session) Set(key any, val any) {
	switch key.(type) {
	case string:
		a, ok := s.table.All[key.(string)]
		if !ok {
			panic(fmt.Errorf(`session type %s has no attribute with name "%s"`, s.t, key))
		}

		v, err := a.GetFieldValue(reflect.Indirect(reflect.ValueOf(s.v)))
		if err != nil {
			panic(fmt.Errorf("get field value error: %w", err))
		}

		v.Set(reflect.ValueOf(val))
	default:
		panic(fmt.Errorf("unsupported key type %T; only strings are supported", key))
	}
}

// Delete deletes (sets to zero value) the attribute with the given string key.
//
// Panics if key is not type string, or key does is not tagged as a `dynamodbav` struct tag (like below), or the type of
// val argument does not have a zero value (is this even possible?):
//
//	type MySession struct {
//		Id    string `dynamodbav:"sessionId,hashkey" tableName:"session"`
//		Field string `dynamodbav:"key"`
//	}
func (s *Session) Delete(key any) {
	switch key.(type) {
	case string:
		a, ok := s.table.All[key.(string)]
		if !ok {
			panic(fmt.Errorf(`session type %s has no attribute with name "%s"`, s.t, key))
		}

		v, err := a.GetFieldValue(reflect.Indirect(reflect.ValueOf(s.v)))
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

		v, err := a.GetFieldValue(reflect.Indirect(reflect.ValueOf(s.v)))
		if err != nil {
			panic(fmt.Errorf("get field value error: %w", err))
		}

		v.Set(reflect.Zero(v.Type()))
	}
}

// AddFlash is not supported at the moment.
func (s *Session) AddFlash(value any, vars ...string) {
	//TODO implement me
	panic("implement me")
}

// Flashes is not supported at the moment.
func (s *Session) Flashes(vars ...string) []any {
	//TODO implement me
	panic("implement me")
}

// Options changes the cookie options.
func (s *Session) Options(options CookieOptions) {
	s.CookieOptions = options
}

// Save writes the session data to DynamoDB as well as updating the session (and CSRF if enabled) cookies.
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

	// use our own setCookie which also uses Expires.
	c := &http.Cookie{
		Name:     s.name,
		Value:    v.String(),
		Expires:  s.CookieOptions.Expires,
		MaxAge:   s.CookieOptions.MaxAge,
		Path:     s.CookieOptions.Path,
		Domain:   s.CookieOptions.Domain,
		SameSite: s.CookieOptions.SameSite,
		Secure:   s.CookieOptions.Secure,
		HttpOnly: s.CookieOptions.HttpOnly,
	}
	if c.Path == "" {
		c.Path = "/"
	}

	http.SetCookie(s.c.Writer, c)

	if s.csrfValue != "" {
		c.Name = s.csrfCookieName
		c.Value = s.csrfValue
		c.HttpOnly = false
		http.SetCookie(s.c.Writer, c)
	}

	return nil
}
