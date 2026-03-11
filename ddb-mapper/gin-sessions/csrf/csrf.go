// Package csrf provides configuration and sources for validating CSRF tokens.
package csrf

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultCookieName is the default cookie name storing CSRF token.
	//
	// From the request, it is optionally used as a CSRF source for validation. It is also set on the response as
	// part of CSRF generation workflow so that user can attach the CSRF token with subsequent requests.
	DefaultCookieName = "__Host-csrf"

	// DefaultHeaderName is the default request header value storing CSRF token.
	//
	// Optionally used as a CSRF source for validation.
	DefaultHeaderName = "X-Csrf-Token"

	// DefaultFormName is the default request form parameter storing CSRF token.
	//
	// Optionally used as a CSRF source for validation.
	DefaultFormName = "csrf_token"
)

// Options customises the CSRF middleware.
type Options struct {
	// Sources contains the various optional ways to retrieve the CSRF token from a request.
	//
	// By default, this value is filled out with FromCookie, FromHeader, and FromForm. Only one of the sources needs to
	// produce a valid and matching CSRF token. If multiple tokens are available from multiple sources, they must all
	// be identical.
	Sources []Source

	// MethodFilter controls which HTTP methods receive CSRF validation.
	//
	// By default, only DELETE, PATCH, POST, and PUT are subject.
	MethodFilter func(string) bool

	// AbortHandler is invoked when the CSRF tokens are invalid.
	//
	// By default, the request chain is aborted with http.StatusForbidden.
	AbortHandler func(*gin.Context)
}

func defaultMethodFilter(method string) bool {
	switch method {
	case http.MethodDelete, http.MethodPatch, http.MethodPost, http.MethodPut:
		return true
	default:
		return false
	}
}
