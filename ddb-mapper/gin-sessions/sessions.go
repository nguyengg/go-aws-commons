package sessions

import (
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Get is Manager.Get without access to a specific Manager instance.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the argument passed to New.
// DefaultSessionIdCookieName is used if none is given.
//
// If Manager.Middleware was not added as a gin middleware prior to this invocation, Get uses default settings.
func Get[T any](c *gin.Context, name ...string) (*T, error) {
	var n string
	if len(name) != 0 {
		n = name[0]
	} else {
		n = DefaultSessionIdCookieName
	}

	m, err := getManager[T](c, n)
	if err != nil {
		return nil, err
	}

	return m.Get(c)
}

// TryGet is Manager.TryGet without access to a specific Manager instance.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the argument passed to New.
// DefaultSessionIdCookieName is used if none is given.
//
// If Manager.Middleware was not added as a gin middleware prior to this invocation, TryGet uses default settings.
func TryGet[T any](c *gin.Context, name ...string) (*T, error) {
	var n string
	if len(name) != 0 {
		n = name[0]
	} else {
		n = DefaultSessionIdCookieName
	}

	m, err := getManager[T](c, n)
	if err != nil {
		return nil, err
	}

	return m.TryGet(c)
}

// Regenerate is Manager.Regenerate without access to a specific Manager instance.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the argument passed to New.
// DefaultSessionIdCookieName is used if none is given.
//
// If Manager.Middleware was not added as a gin middleware prior to this invocation, Regenerate uses default settings.
func Regenerate[T any](c *gin.Context, name ...string) (*T, error) {
	var n string
	if len(name) != 0 {
		n = name[0]
	} else {
		n = DefaultSessionIdCookieName
	}

	m, err := getManager[T](c, n)
	if err != nil {
		return nil, err
	}

	return m.Regenerate(c)
}

// Save is Manager.Save without access to a specific Manager instance.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the argument passed to New.
// DefaultSessionIdCookieName is used if none is given.
//
// If Manager.Middleware was not added as a gin middleware prior to this invocation, Save uses default settings.
func Save[T any](c *gin.Context, name ...string) error {
	var n string
	if len(name) != 0 {
		n = name[0]
	} else {
		n = DefaultSessionIdCookieName
	}

	m, err := getManager[T](c, n)
	if err != nil {
		return err
	}

	return m.Save(c)
}

// Destroy is Manager.Destroy without access to a specific Manager instance.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the argument passed to New.
// DefaultSessionIdCookieName is used if none is given.
//
// If Manager.Middleware was not added as a gin middleware prior to this invocation, Destroy uses default settings.
func Destroy[T any](c *gin.Context, name ...string) error {
	var n string
	if len(name) != 0 {
		n = name[0]
	} else {
		n = DefaultSessionIdCookieName
	}

	m, err := getManager[T](c, n)
	if err != nil {
		return err
	}

	return m.Destroy(c)
}

// DefaultNewSessionId creates a new UUID and returns its raw-URL-encoded content.
func DefaultNewSessionId() string {
	data := uuid.New()
	return base64.RawURLEncoding.EncodeToString(data[:])
}
