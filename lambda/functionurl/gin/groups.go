package ginadapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/lambda/functionurl/gin/rules"
)

// RequireGroupMembership returns a gin middleware that aborts the request with either http.StatusUnauthorized or
// http.StatusForbidden depending on whether there exists a user with the current session whose group membership
// satisfies the given rules.
//
// The middleware must be given a function that can retrieve the user's group from the current request. The argument fn
// returns whether the session is authenticated and the groups associated with the user. If session is not authenticated
// then the request is aborted with http.StatusUnauthorized. If the session is authenticated but the groups do not
// satisfy the rules, the request is aborted with http.StatusForbidden. Otherwise, the request goes through.
func RequireGroupMembership(fn func(*gin.Context) (authenticated bool, groups rules.Groups), rule rules.Rule, more ...rules.Rule) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, groups := fn(c)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !groups.Test(rule, more...) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
