package sessions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/gbac"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/internal/groups"
)

// WithGBAC enables group-based access control (GBAC) functionality when passed to [New].
//
// The extractGroups argument must be given in order to extract the groups associated with the current session's user.
// If there is no current session, or if the current session has no authenticated user, the function must return false.
//
// [Manager.Authorize] will panic if WithGBAC wasn't passed to [New].
func WithGBAC(extractGroups func(c *gin.Context) (authenticated bool, groups gbac.Groups), optFns ...func(opts *gbac.Options)) func(cfg *Config) {
	if extractGroups == nil {
		panic("extractGroups is nil")
	}

	opts := &gbac.Options{}
	for _, fn := range optFns {
		fn(opts)
	}
	return func(cfg *Config) {
		cfg.extractGroups = extractGroups
		cfg.groupsOptions = *opts
	}
}

// Authorize creates a middleware to validate that the session is authenticated, and user's groups satisfy the given
// rules.
func (m *Manager[T]) Authorize(rule gbac.Rule, more ...gbac.Rule) gin.HandlerFunc {
	if m.extractGroups == nil {
		panic("WithGBAC was not used to create sessions.Manager")
	}

	rules := (&groups.Rules{}).Apply(rule, more...)

	methodFilter := m.groupsOptions.MethodFilter
	if methodFilter == nil {
		methodFilter = gbac.DefaultMethodFilter
	}

	unauthorisedHandler := rules.UnauthorizedHandler
	if unauthorisedHandler == nil {
		if unauthorisedHandler = m.groupsOptions.UnauthorizedHandler; unauthorisedHandler == nil {
			unauthorisedHandler = func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }
		}
	}

	forbiddenHandler := rules.ForbiddenHandler
	if forbiddenHandler == nil {
		if forbiddenHandler = m.groupsOptions.ForbiddenHandler; forbiddenHandler == nil {
			forbiddenHandler = func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) }
		}
	}

	return func(c *gin.Context) {
		authenticated, gs := m.extractGroups(c)
		if !authenticated {
			unauthorisedHandler(c)
			return
		}

		if !rules.Test(gs) {
			forbiddenHandler(c)
			return
		}

		c.Next()
	}
}
