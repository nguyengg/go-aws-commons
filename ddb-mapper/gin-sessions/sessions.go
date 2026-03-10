// Package sessions is a bring-your-own-DynamoDB-struct sessions management tightly integrated with [github.com/nguyengg/go-aws-commons/ddb-mapper].
//
// CSRF generation and validation integration is also supported out-of-the-box with
// [github.com/nguyengg/go-aws-commons/opaque-token/hmac].
//
// There are two ways to use this package. First is by constructing a manager with [New], then accessing the session
// struct by way of the [Manager] interface. This is the preferred and more type-safe usage pattern.
//
// TODO add usage.
//
// Second is by using the package-level methods [Get], [Regenerate], [Save], and [Destroy] counterparts to the [Manager]
// methods . The second usage pattern is useful if you're writing middlewares or handlers that do not have access to
// a [Manager]. However, to customise session settings in this mode, you should still create a [Manager] then add
// [Manager.Middleware] to the handler chain. Without this, the package-level methods will assume all
// default settings without the ability to generate CSRF tokens.
//
// [github.com/nguyengg/go-aws-commons/ddb-mapper]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb-mapper
// [github.com/nguyengg/go-aws-commons/opaque-token/hmac]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/opaque-token/hmac
package sessions

import (
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Get calls [Manager.Get] on the [Manager.Middleware] (or a default one) attached to the request.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the [Config.SessionIdCookieName]
// passed to [New].
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

// TryGet calls [Manager.TryGet] on the [Manager.Middleware] (or a default one) attached to the request.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the [Config.SessionIdCookieName]
// passed to [New]. Defaults to [DefaultSessionIdCookieName].
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

// Regenerate calls [Manager.Regenerate] on the [Manager.Middleware] (or a default one) attached to the request.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the [Config.SessionIdCookieName]
// passed to [New]. Defaults to [DefaultSessionIdCookieName].
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

// Save calls [Manager.Save] on the [Manager.Middleware] (or a default one) attached to the request.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the [Config.SessionIdCookieName]
// passed to [New]. Defaults to [DefaultSessionIdCookieName].
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

// Destroy calls [Manager.Destroy] on the [Manager.Middleware] (or a default one) attached to the request.
//
// The (first) name argument is the cookie name that stores the session Id, similar to the [Config.SessionIdCookieName]
// passed to [New]. Defaults to [DefaultSessionIdCookieName].
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
