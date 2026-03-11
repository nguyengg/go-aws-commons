package gbac

import (
	"github.com/gin-gonic/gin"
)

// Options customises the RequireGroups middleware.
type Options struct {
	// MethodFilter controls which HTTP methods are subject to this middleware.
	//
	// By default, all methods are.
	MethodFilter func(string) bool

	// UnauthorizedHandler is invoked when the session is unauthenticated.
	//
	// By default, the request chain is aborted with http.StatusUnauthorized with no additional messaging.
	UnauthorizedHandler func(c *gin.Context)

	// ForbiddenHandler is invoked when the user's groups do not pass membership rules.
	//
	// By default, the request chain is aborted with http.StatusForbidden with no additional messaging.
	ForbiddenHandler func(c *gin.Context)
}

// DefaultMethodFilter is the default value for [Options.MethodFilter] that allows all methods.
func DefaultMethodFilter(_ string) bool {
	return true
}
