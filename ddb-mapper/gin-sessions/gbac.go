package sessions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/gbac"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/internal/groups"
)

// EnableGBAC enables group-based access control (GBAC) functionality.
//
// EnableGBAC should only be called once; subsequent calls will replace the function to extract groups.
//
// [Manager.Authorize] will panic if Authorize has not been called.
func (m *Manager[T]) EnableGBAC(getGroupsFn gbac.GetGroupsFunc, optFns ...func(opts *gbac.Options)) *Manager[T] {
	if getGroupsFn == nil {
		panic("getGroupsFn is nil")
	}
	opts := &gbac.Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	m.getGroupsFn = getGroupsFn
	m.groupsOpts = *opts
	return m
}

// Authorize creates a middleware to validate that the session is authenticated and user's groups satisfy the given
// rules.
//
// Panics if [Manager.EnableGBAC] has not been called.
//
// Usage:
//
//	type User struct {
//		Sub    string   `dynamodbav:"sub,string" tablename:"Users"`
//		Groups []string `dynamodbav:"groups,stringset"`
//	}
//
//	type Session struct {
//		ID   string `dynamodbav:"id,hashkey" tablename:"Sessions"`
//		User *User  `dynamodbav:"user"`
//	}
//
//	m, _ := sessions.New[Session]()
//
//	r := gin.Default()
//	r.PUT("/resource/:id",
//		m.EnableGBAC(func(c *gin.Context) (authenticated bool, groups []string) {
//			switch s, err := m.Get(c); {
//			case err != nil:
//				_ = c.AbortWithError(500, err)
//			case s.User == nil:
//				return false, nil
//			default:
//				return true, s.User.Groups
//			}
//		}).Authorize(gbac.AllOf("resource_admin")))
func (m *Manager[T]) Authorize(rule gbac.Rule, more ...gbac.Rule) gin.HandlerFunc {
	if m.getGroupsFn == nil {
		panic("WithGBAC was not used to create sessions.Manager")
	}

	rules := (&groups.Rules{}).Apply(rule, more...)

	methodFilter := m.groupsOpts.MethodFilter
	if methodFilter == nil {
		methodFilter = gbac.DefaultMethodFilter
	}

	unauthorisedHandler := rules.UnauthorizedHandler
	if unauthorisedHandler == nil {
		if unauthorisedHandler = m.groupsOpts.UnauthorizedHandler; unauthorisedHandler == nil {
			unauthorisedHandler = func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }
		}
	}

	forbiddenHandler := rules.ForbiddenHandler
	if forbiddenHandler == nil {
		if forbiddenHandler = m.groupsOpts.ForbiddenHandler; forbiddenHandler == nil {
			forbiddenHandler = func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) }
		}
	}

	return func(c *gin.Context) {
		switch authenticated, gs := m.getGroupsFn(c); {
		case c.IsAborted():
			// return
		case !authenticated:
			unauthorisedHandler(c)
		case !rules.Test(gs):
			forbiddenHandler(c)
		default:
			c.Next()
		}
	}
}
