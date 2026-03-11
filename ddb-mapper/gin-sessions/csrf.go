package sessions

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/csrf"
	"github.com/nguyengg/go-aws-commons/opaque-token/hmac"
)

// WithCSRF configures [New] to use CSRF generation with the given signer and verifier.
//
// The same [hmac.Engine] will be used for CSRF validation as well. See [github.com/nguyengg/go-aws-commons/opaque-token/hmac]
// for various options on constructing the [hmac.Engine].
//
// [github.com/nguyengg/go-aws-commons/opaque-token/hmac]: https://pkg.go.dev/github.com/nguyengg/go-aws-commons/opaque-token/hmac
func WithCSRF(csrf hmac.Engine) func(cfg *Config) {
	return func(cfg *Config) {
		cfg.csrf = csrf
	}
}

// ValidateCSRF creates a middleware to validate the CSRF tokens that were created by the same [hmac.Engine] passed to
// [New] by way of [WithCSRF].
//
// Panics if you did not pass [WithCSRF], unable to return a middleware.
func (m *Manager[T]) ValidateCSRF(optFns ...func(opts *csrf.Options)) gin.HandlerFunc {
	if m.csrf == nil {
		panic("New was not passed WithCSRF to assign a non-nil hmac.Engine")
	}

	opts := &csrf.Options{
		Sources:      []csrf.Source{csrf.FromCookie(), csrf.FromForm(), csrf.FromHeader()},
		MethodFilter: defaultMethodFilter,
	}
	for _, fn := range optFns {
		fn(opts)
	}

	sources := slices.Clone(opts.Sources)

	methodFilter := opts.MethodFilter
	if methodFilter == nil {
		methodFilter = defaultMethodFilter
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

func defaultMethodFilter(method string) bool {
	switch method {
	case http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut:
		return true
	default:
		return false
	}
}
