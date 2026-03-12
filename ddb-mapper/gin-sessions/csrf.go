package sessions

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/csrf"
	"github.com/nguyengg/go-aws-commons/opaque-token/hmac"
	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
)

// EnableCSRF enables CSRF generation and validation using the given key.
//
// EnableCSRF should only be called once; subsequent calls will replace the key provider, invalidating tokens created
// with inaccessible keys.
//
// [Manager.ValidateCSRF] will panic if EnableCSRF has not been called.
func (m *Manager[T]) EnableCSRF(key keys.Provider, optFns ...func(opts *csrf.Options)) *Manager[T] {
	if key == nil {
		panic("key is nil")
	}

	opts := csrf.Options{}
	for _, fn := range optFns {
		fn(&opts)
	}

	m.csrfSignVerifier = hmac.New(key)

	m.csrfOpts = opts
	if m.csrfOpts.CookieName == "" {
		m.csrfOpts.CookieName = csrf.DefaultCookieName
	}
	if len(m.csrfOpts.Sources) == 0 {
		m.csrfOpts.Sources = []csrf.Source{
			csrf.FromCookie(opts.CookieName),
			csrf.FromHeader(opts.CookieName),
			csrf.FromForm(opts.CookieName),
		}
	}
	if m.csrfOpts.MethodFilter == nil {
		m.csrfOpts.MethodFilter = csrf.DefaultMethodFilter
	}
	if m.csrfOpts.ForbiddenHandler == nil {
		m.csrfOpts.ForbiddenHandler = func(c *gin.Context) {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}

	return m
}

// ValidateCSRF creates a middleware to validate the CSRF tokens.
//
// Panics if [Manager.EnableCSRF] has not been called.
//
// Usage:
//
//	key := make([]byte, 32)
//	_, _ = rand.Read(key)
//
//	m, _ := sessions.New[Session]()
//	m.EnableCSRF(keys.Static(key)) // use something else in production, don't use static key
//
//	r := gin.Default()
//	r.PUT("/resource/:id",
//		// validate signed double-submit cookie.
//		m.ValidateCSRF(csrfSignVerifier.DoubleSubmit(csrfSignVerifier.FromCookie(), csrfSignVerifier.FromHeader())),
//		func(c *gin.Context) { /* if this handler is run, CSRF validation passes. */ })
func (m *Manager[T]) ValidateCSRF(optFns ...func(opts *csrf.Options)) gin.HandlerFunc {
	if m.csrfSignVerifier == nil {
		panic("EnableCSRF must be called prior to ValidateCSRF")
	}

	opts := m.csrfOpts
	for _, fn := range optFns {
		fn(&opts)
	}

	sources := slices.Clone(opts.Sources)
	if len(sources) == 0 {
		sources = []csrf.Source{
			csrf.FromCookie(opts.CookieName),
			csrf.FromHeader(opts.CookieName),
			csrf.FromForm(opts.CookieName),
		}
	}

	methodFilter := opts.MethodFilter
	if methodFilter == nil {
		methodFilter = csrf.DefaultMethodFilter
	}

	abortHandler := opts.ForbiddenHandler
	if abortHandler == nil {
		abortHandler = func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) }
	}

	return func(c *gin.Context) {
		if !methodFilter(c.Request.Method) {
			c.Next()
			return
		}

		// failure to retrieve session Id will abort with 500.
		sid, err := m.getSid(c)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("get session id error: %w", err))
			return
		}

		// go through all sources to get the CSRF token. if more than one sources is available, compare them right away.
		var token []byte
		for _, src := range sources {
			t, err := src(c)
			if err != nil {
				_ = c.Error(err)
				abortHandler(c)
				return
			}

			if len(token) != 0 {
				if subtle.ConstantTimeCompare(t, token) != 1 {
					_ = c.Error(fmt.Errorf("conflicting csrfSignVerifier tokens from sources"))
					abortHandler(c)
					return
				}
			}

			token = t
		}

		if len(token) == 0 {
			_ = c.Error(fmt.Errorf("no csrfSignVerifier token available"))
			abortHandler(c)
			return
		}

		ok, err := m.csrfSignVerifier.Verify(c, token, []byte(sid))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("validate csrfSignVerifier token error: %w", err))
			return
		}

		if !ok {
			_ = c.Error(fmt.Errorf("mismatched csrfSignVerifier token"))
			abortHandler(c)
			return
		}

		c.Next()
	}
}
