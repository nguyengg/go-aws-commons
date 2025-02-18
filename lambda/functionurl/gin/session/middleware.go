package session

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Options customises the session middleware.
//
// To configure the middleware, an instance of Options must be set to
type Options struct {
	// CookieName is the name of the cookie to retrieve the session Id.
	//
	// The default value is "sid".
	CookieName string

	// SessionKey is the key to attach a valid session with a given [gin.Context] via [gin.Context.Set].
	//
	// The default value is "session". Pass empty string to disable this feature.
	SessionKey string

	// UserKey is the key to attach a valid user with a given [gin.Context] via [gin.Context.Set].
	//
	// The default value is "user". Pass empty string to disable this feature.
	UserKey string

	rules []Rule
}

// GetSession is a function that can retrieve a session from a session Id.
type GetSession[SessionType any] func(ctx context.Context, sessionId string) (*SessionType, error)

// HasUser is the interface for the GetUser method which can be used to return information about a user attached with a
// session.
type HasUser[UserType any] interface {
	GetUser() *UserType
}

// RequireSession adds a middleware that rejects all requests with http.StatusUnauthorized that don't have a valid
// session.
//
// The argument f is responsible for retrieving the session data from a session Id which is the value of a cookie named
// by [Options.CookieName]. The return value of f, if non-nil, will be attached to the gin.Context for further use by
// key [Options.SessionKey].
func RequireSession[SessionType any](f GetSession[SessionType], optFns ...func(*Options)) gin.HandlerFunc {
	opts := &Options{
		CookieName: "sid",
		SessionKey: "session",
	}
	for _, fn := range optFns {
		fn(opts)
	}

	return func(c *gin.Context) {
		sessionId, err := c.Cookie(opts.CookieName)
		if errors.Is(err, http.ErrNoCookie) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		s, err := f(c, sessionId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if s == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if opts.SessionKey != "" {
			c.Set(opts.SessionKey, s)
		}

		c.Next()
	}
}

// RequireAuthentication is a stronger variant of RequireSession that requires the session to have an authenticated user
// by HasUser.GetUser returning a non-nil value.
//
// The argument f is responsible for retrieving the session data from a session Id which is the value of a cookie named
// by [Options.CookieName]. Both the return value of f and of HasUser.GetUser, if non-nil, will be attached to the
// gin.Context for further use by key [Options.SessionKey] and [Options.UserKey].
func RequireAuthentication[UserType interface{}](f GetSession[HasUser[UserType]], optFns ...func(*Options)) gin.HandlerFunc {
	opts := &Options{
		CookieName: "sid",
		SessionKey: "session",
		UserKey:    "user",
	}
	for _, fn := range optFns {
		fn(opts)
	}

	return func(c *gin.Context) {
		sessionId, err := c.Cookie(opts.CookieName)
		if errors.Is(err, http.ErrNoCookie) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		s, err := f(c, sessionId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if s == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if opts.SessionKey != "" {
			c.Set(opts.SessionKey, s)
		}

		user := (*s).GetUser()
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if opts.UserKey != "" {
			c.Set(opts.UserKey, user)
		}

		c.Next()
	}
}

// HasGroups is the interface for the GetGroups method which returns all the groups that the user belongs to.
//
// The return value is intended to be used with Groups.Test to test for membership.
type HasGroups interface {
	GetGroups() []string
}

// RequireAuthorisation implies RequireAuthentication (which implies RequireSession) while also test that the
// authenticated user is authorised to perform some action using group membership model.
func RequireAuthorisation[UserType HasGroups](f GetSession[HasUser[UserType]], rule Rule, more ...Rule) gin.HandlerFunc {
	if len(more) == 0 {
		return RequireAuthorisationWithOptions(f, rule)
	}

	return RequireAuthorisationWithOptions(f, rule, WithRule(more[0], more[1:]...))
}

// RequireAuthorisationWithOptions is a variant of RequireAuthorisation that accepts custom options.
//
// If you need to specify more than one rules, pass WithRule.
func RequireAuthorisationWithOptions[UserType HasGroups](f GetSession[HasUser[UserType]], rule Rule, optFns ...func(options *Options)) gin.HandlerFunc {
	opts := &Options{
		CookieName: "sid",
		SessionKey: "session",
		UserKey:    "user",
	}
	for _, fn := range optFns {
		fn(opts)
	}

	return func(c *gin.Context) {
		sessionId, err := c.Cookie(opts.CookieName)
		if errors.Is(err, http.ErrNoCookie) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		s, err := f(c, sessionId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if s == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if opts.SessionKey != "" {
			c.Set(opts.SessionKey, s)
		}

		user := (*s).GetUser()
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if opts.UserKey != "" {
			c.Set(opts.UserKey, user)
		}

		if !Groups((*user).GetGroups()).Test(rule, opts.rules...) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

// WithRule can be used with RequireAuthorisationWithOptions
func WithRule(rule Rule, more ...Rule) func(*Options) {
	return func(opts *Options) {
		opts.rules = append(opts.rules, rule)
		opts.rules = append(opts.rules, more...)
	}
}
