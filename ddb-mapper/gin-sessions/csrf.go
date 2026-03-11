package sessions

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/csrf"
)

// DisableCSRF disables CSRF generation and validation when passed to [New].
//
// [Manager.ValidateCSRF] will panic if DisableCSRF was passed to [New] to create the [Manager].
func DisableCSRF() func(cfg *Config) {
	return func(cfg *Config) {
		cfg.csrfDisabled = true
	}
}

// ValidateCSRF creates a middleware to validate the CSRF tokens.
//
// Panics if [New] was called [DisableCSRF].
//
// Usage:
//
//	r := gin.Default()
//	m, _ := sessions.New[Session]()
//	// this will require that request has identical CSRF token from both cookie and header.
//	// the token will also be validated against the session Id as well.
//	r.Use(m.ValidateCSRF(csrf.DoubleSubmit(csrf.FromCookie(), csrf.FromHeader())))
func (m *Manager[T]) ValidateCSRF(optFns ...func(opts *csrf.Options)) gin.HandlerFunc {
	if m.csrfDisabled {
		panic("DisableCSRF was used to create sessions.Manager")
	}

	opts := &csrf.Options{
		Sources:      []csrf.Source{csrf.FromCookie(), csrf.FromForm(), csrf.FromHeader()},
		MethodFilter: csrf.DefaultMethodFilter,
	}
	for _, fn := range optFns {
		fn(opts)
	}

	sources := slices.Clone(opts.Sources)

	methodFilter := opts.MethodFilter
	if methodFilter == nil {
		methodFilter = csrf.DefaultMethodFilter
	}

	abortHandler := opts.AbortHandler
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
					_ = c.Error(fmt.Errorf("conflicting csrf tokens from sources"))
					abortHandler(c)
					return
				}
			}

			token = t
		}

		if len(token) == 0 {
			_ = c.Error(fmt.Errorf("no csrf token available"))
			abortHandler(c)
			return
		}

		ok, err := m.csrf.Verify(c, token, []byte(sid))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("validate csrf token error: %w", err))
			return
		}

		if !ok {
			_ = c.Error(fmt.Errorf("mismatched csrf token"))
			abortHandler(c)
			return
		}

		c.Next()
	}
}
