package gbac

import (
	"github.com/gin-gonic/gin"
)

// Options customises the Authorize middleware.
type Options struct {
	// MethodFilter controls which HTTP methods are subject to this middleware.
	//
	// By default, all methods are.
	MethodFilter func(string) bool

	// UnauthorizedHandler is invoked when the session is unauthenticated.
	//
	// By default, the request chain is aborted with http.StatusUnauthorized.
	UnauthorizedHandler func(c *gin.Context)

	// ForbiddenHandler is invoked when the user's groups do not pass membership rules.
	//
	// By default, the request chain is aborted with http.StatusForbidden.
	ForbiddenHandler func(c *gin.Context)
}

// DefaultMethodFilter is the default value for [Options.MethodFilter] that allows all methods.
func DefaultMethodFilter(_ string) bool {
	return true
}

// GetGroupsFunc determines whether the request is authenticated with valid groups for testing membership.
//
// If there is no session associated with the request, GetGroupsFunc will not be called and the request is treated as
// unauthenticated. GetGroupsFunc must return (false, nil) if any of these conditions are met:
//  1. The current session has no authenticated user.
//  2. There was an error retrieving the user's groups, in which case GetGroupsFunc should also abort the request.
type GetGroupsFunc func(c *gin.Context) (authenticated bool, groups []string)
