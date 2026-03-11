// Package sessions is a bring-your-own-DynamoDB-struct sessions management tightly integrated with
// [github.com/nguyengg/go-aws-commons/ddb-mapper].
//
// CSRF generation and validation integration is also supported out-of-the-box with
// [github.com/nguyengg/go-aws-commons/opaque-token/hmac].
//
// There are two ways to use this package. First is by constructing a manager with [New], then accessing the session
// struct by way of [Manager]. This is the preferred and more type-safe usage pattern.
//
// TODO add usage.
//
// Second is by using the package-level methods [Get], [Regenerate], [Save], and [Destroy] counterparts to the [Manager]
// methods. This usage pattern is useful if you're writing middlewares or handlers that do not have direct access to a
// [Manager]. However, you must still have attached a [Manager.Middleware] to the handler chain; failure to do so will
// result in a panic similar to [github.com/gin-contrib/sessions] `sessions.Default`.
//
// [github.com/nguyengg/go-aws-commons/ddb-mapper]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb-mapper
// [github.com/nguyengg/go-aws-commons/opaque-token/hmac]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/opaque-token/hmac
// [github.com/gin-contrib/sessions]: https://pkg.go.dev/github.com/gin-contrib/sessions
package sessions

import (
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Get calls [Manager.Get] on the [Manager.Middleware] attached to the request.
//
// The single variadic argument is the cookie name that stores the session Id (see [Config.SessionIdCookieName]). Panics
// if no [Manager.Middleware] with the same name has been attached to the handler chain.
//
// Get returns pointer to struct type T that was used as the type argument for [New].
func Get(c *gin.Context, name ...string) (any, error) {
	m, err := getManager(c, name...)
	if err != nil {
		return nil, err
	}

	return m.get(c)
}

// TryGet calls [Manager.TryGet] on the [Manager.Middleware] attached to the request.
//
// The single variadic argument is the cookie name that stores the session Id (see [Config.SessionIdCookieName]). Panics
// if no [Manager.Middleware] with the same name has been attached to the handler chain.
//
// TryGet returns pointer to struct type T that was used as the type argument for [New].
func TryGet(c *gin.Context, name ...string) (any, error) {
	m, err := getManager(c, name...)
	if err != nil {
		return nil, err
	}

	return m.tryGet(c)
}

// Regenerate calls [Manager.Regenerate] on the [Manager.Middleware] attached to the request.
//
// The single variadic argument is the cookie name that stores the session Id (see [Config.SessionIdCookieName]). Panics
// if no [Manager.Middleware] with the same name has been attached to the handler chain.
//
// Regenerate returns pointer to struct type T that was used as the type argument for [New].
func Regenerate(c *gin.Context, name ...string) (any, error) {
	m, err := getManager(c, name...)
	if err != nil {
		return nil, err
	}

	return m.regenerate(c)
}

// Save calls [Manager.Save] on the [Manager.Middleware] (or a default one) attached to the request.
//
// The single variadic argument is the cookie name that stores the session Id (see [Config.SessionIdCookieName]). Panics
// if no [Manager.Middleware] with the same name has been attached to the handler chain.
func Save(c *gin.Context, name ...string) error {
	m, err := getManager(c, name...)
	if err != nil {
		return err
	}

	return m.Save(c)
}

// Destroy calls [Manager.Destroy] on the [Manager.Middleware] attached to the request.
//
// The single variadic argument is the cookie name that stores the session Id (see [Config.SessionIdCookieName]). Panics
// if no [Manager.Middleware] with the same name has been attached to the handler chain.
func Destroy(c *gin.Context, name ...string) error {
	m, err := getManager(c, name...)
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
