package sessions

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/opaque-token/hmac"
)

const (
	DefaultCSRFCookieName = "__Host-csrf"
	DefaultCSRFHeaderName = "X-Csrf-Token"
	DefaultCSRFFormName   = "csrf_token"
)

var (
	ErrNoCSRFCookie         = errors.New("no CSRF cookie")
	ErrNoCSRFHeader         = errors.New("no CSRF header")
	ErrNoCSRFForm           = errors.New("no CSRF form parameter")
	ErrMismatchDoubleSubmit = errors.New("no CSRF form parameter")
)

// CSRFOptions customises the CSRF middleware.
type CSRFOptions struct {
	// Sources contains the various optional ways to retrieve the CSRF token from a request.
	//
	// By default, this value is filled out with CSRFFromCookie(DefaultCSRFCookieName),
	// CSRFFromHeader(DefaultCSRFHeaderName), and CSRFFromForm(DefaultCSRFFormName), all base64.RawURLEncoding.
	Sources []CSRFSource

	// MethodFilter controls which HTTP methods receive CSRF validation.
	//
	// By default, only DELETE, PATCH, POST, and PUT are subject.
	MethodFilter func(string) bool

	// AbortHandler is invoked when the CSRF tokens are invalid.
	//
	// By default, the request chain is aborted with http.StatusForbidden.
	AbortHandler func(*gin.Context)

	hasher hmac.Hasher
}

// CSRFSource provides a way to retrieve CSRF token from request.
type CSRFSource func(*gin.Context) ([]byte, error)

// RequireCSRF creates a gin middleware for validating CSRF tokens from several potential sources.
//
// CSRF requires Sessions to have been set up to provide a valid session Id that will be used as the payload for
// verifying the CSRF token.
func RequireCSRF(hasher hmac.Hasher, optFns ...func(*CSRFOptions)) gin.HandlerFunc {
	m := &CSRFOptions{
		Sources: []CSRFSource{
			CSRFFromCookie(DefaultCSRFCookieName),
			CSRFFromHeader(DefaultCSRFHeaderName),
			CSRFFromForm(DefaultCSRFFormName),
		},
		MethodFilter: defaultMethodFilter,
		hasher:       hasher,
	}
	for _, fn := range optFns {
		fn(m)
	}

	return m.handle
}

// WithCSRF attaches to the session middleware the ability to set CSRF cookie as well when a new session is created.
//
// The cookie will use the same settings as Session.CookieOptions but with [Options.HttpOnly] set to false. The CSRF
// token will be saved to the context and can be retrieved using GetCSRF if it needs to be embedded in the response
// as hidden form input.
func WithCSRF(hasher hmac.Hasher, name string) func(*Session) {
	return func(s *Session) {
		s.hasher = hasher
		s.csrfCookieName = name
	}
}

// GetCSRF returns the CSRF token associated with the given session.
//
// The returned value is the expected CSRF token generated from the session's Id. If WithCSRF was not set up, this
// method always returns an empty string.
func GetCSRF(c *gin.Context) string {
	return Default(c).csrfValue
}

// CSRFFromCookie retrieves the CSRF base64-raw-url-encoded token from cookie with the given name.
func CSRFFromCookie(name string) CSRFSource {
	return func(c *gin.Context) ([]byte, error) {
		v, err := c.Cookie(name)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				return nil, ErrNoCSRFCookie
			}

			return nil, err
		}

		return base64.RawURLEncoding.DecodeString(v)
	}
}

// CSRFFromHeader retrieves the CSRF base64-raw-url-encoded token from request header with the given name.
func CSRFFromHeader(name string) CSRFSource {
	return func(c *gin.Context) ([]byte, error) {
		v := c.GetHeader(name)
		if v == "" {
			return nil, ErrNoCSRFHeader
		}

		return base64.RawURLEncoding.DecodeString(v)
	}
}

// CSRFFromForm retrieves the CSRF base64-raw-url-encoded token from the POST form parameter with the given name.
func CSRFFromForm(name string) CSRFSource {
	return func(c *gin.Context) ([]byte, error) {
		v, ok := c.GetPostForm(name)
		if v == "" || !ok {
			return nil, ErrNoCSRFForm
		}

		return base64.RawURLEncoding.DecodeString(v)
	}
}

// DoubleSubmit validates that all of the given CSRF sources must be available AND identical.
//
// Useful if you use double-submit cookie pattern. This method replaces the existing [CSRFOptions.Sources].
func DoubleSubmit(source CSRFSource, more ...CSRFSource) func(*CSRFOptions) {
	return func(options *CSRFOptions) {
		options.Sources = []CSRFSource{func(c *gin.Context) (token []byte, err error) {
			token, err = source(c)
			if err != nil {
				return nil, err
			}

			for _, fn := range more {
				t, err := fn(c)
				if err != nil {
					return nil, err
				}
				if subtle.ConstantTimeCompare(token, t) != 1 {
					return nil, ErrMismatchDoubleSubmit
				}
			}

			return token, nil
		}}
	}
}

func (m *CSRFOptions) handle(c *gin.Context) {
	var (
		token        []byte
		err, lastErr error
	)
	for _, s := range m.Sources {
		token, err = s(c)
		if err != nil {
			lastErr = err
			continue
		}
	}

	if lastErr != nil {
		if m.AbortHandler != nil {
			m.AbortHandler(c)
		} else {
			_ = c.AbortWithError(http.StatusForbidden, fmt.Errorf("csrf: retrieve token error: %w", lastErr))
			return
		}
	}

	ok, err := m.hasher.Verify(c, token, []byte(Default(c).ID()))
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("csrf: validate token error: %w", err))
	}

	if !ok {
		if m.AbortHandler != nil {
			m.AbortHandler(c)
		} else {
			_ = c.AbortWithError(http.StatusForbidden, fmt.Errorf("mismatched CSRF token")).SetType(gin.ErrorTypePublic)
			return
		}
	}

	c.Next()
}

func defaultMethodFilter(method string) bool {
	switch method {
	case http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut:
		return true
	default:
		return false
	}
}
