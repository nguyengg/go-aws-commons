package csrf

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Source provides a way to retrieve CSRF token from request.
//
// If the CSRF token is missing, (nil, nil) should be returned. Invalid tokens must return a non-nil error to abort the
// request.
type Source func(*gin.Context) ([]byte, error)

// FromCookie retrieves the CSRF base64-raw-url-encoded token from cookie.
//
// The single variadic name argument provides the optional name which defaults to [DefaultCookieName].
func FromCookie(name ...string) Source {
	n := firstOrDefault(DefaultCookieName, name...)

	return func(c *gin.Context) ([]byte, error) {
		v, err := c.Cookie(n)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				return nil, nil
			}

			return nil, err
		}

		return base64.RawURLEncoding.DecodeString(v)
	}
}

// FromHeader retrieves the CSRF base64-raw-url-encoded token from request header.
//
// The single variadic name argument provides the optional header key which defaults to [DefaultHeaderName].
func FromHeader(name ...string) Source {
	n := firstOrDefault(DefaultHeaderName, name...)

	return func(c *gin.Context) ([]byte, error) {
		switch vs := c.Request.Header.Values(n); len(vs) {
		case 0:
			return nil, nil
		case 1:
			return base64.RawURLEncoding.DecodeString(vs[0])
		default:
			return nil, errors.New("multiple CSRF header values found")
		}
	}
}

// FromForm retrieves the CSRF base64-raw-url-encoded token from the POST form parameter.
//
// The single variadic name argument provides the optional header key which defaults to [DefaultFormName].
func FromForm(name ...string) Source {
	n := firstOrDefault(DefaultFormName, name...)

	return func(c *gin.Context) ([]byte, error) {
		vs, ok := c.GetPostFormArray(n)
		if !ok {
			return nil, nil
		}

		switch len(vs) {
		case 0:
			return nil, nil
		case 1:
			return base64.RawURLEncoding.DecodeString(vs[0])
		default:
			return nil, errors.New("multiple CSRF form values found")
		}
	}
}

// DoubleSubmit validates that all the given CSRF sources must be available AND match.
//
// Useful if you use double-submit cookie pattern. This method replaces the existing [Options.Sources].
//
// Usage:
//
//	r := gin.Default()
//	m, _ := sessions.New[Session]()
//	// this will require that request has identical CSRF token from both cookie and header.
//	// the token will also be validated against the session Id as well.
//	r.Use(m.ValidateCSRF(csrf.DoubleSubmit(csrf.FromCookie(), csrf.FromHeader())))
func DoubleSubmit(source Source, more ...Source) func(opts *Options) {
	return func(opts *Options) {
		opts.Sources = []Source{func(c *gin.Context) (token []byte, err error) {
			if token, err = source(c); err != nil {
				return nil, err
			}

			for _, fn := range more {
				t, err := fn(c)
				if err != nil {
					return nil, err
				}
				if subtle.ConstantTimeCompare(token, t) != 1 {
					return nil, errors.New("conflicting CSRF token from sources")
				}
			}

			return token, nil
		}}
	}
}

// firstOrDefault returns def if options is empty, or the first option otherwise.
func firstOrDefault(def string, options ...string) string {
	if len(options) == 0 {
		return def
	}

	return options[0]
}
