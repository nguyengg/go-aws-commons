package groups

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MustHave returns a gin middleware that aborts the request with either http.StatusUnauthorized or
// http.StatusForbidden depending on whether there exists a user with the current session whose group membership
// satisfies the given rules.
//
// The middleware must be given function fn that returns whether the session is authenticated (i.e. session exists with
// a valid user) and also retrieves the user's group from the current request. If session is not authenticated then the
// request is aborted with http.StatusUnauthorized (can be customised with WithUnauthorizedHandler). If the session is
// authenticated but the groups do not satisfy the rules, the request is aborted with http.StatusForbidden (can be
// customised with WithForbiddenHandler to provide a meaningful error message such as "user must belong to ABC group").
// Otherwise, the request goes through unimpeded.
//
// Usage:
//
//	type MySession struct {
//		SessionId string `dynamodbav:"sessionId,hashkey" tableName:"sessions"`
//		User      *User `dynamodbav:"user,omitempty"`
//	}
//
//	type User struct {
//		Sub    string `dynamodbav:"user"`
//		Groups []string `dynamodbav:"groups,stringset"
//	}
//
//	r := gin.Default()
//	r.Use(sessions.Session[MySession]("sid"))
//	r.GET(
//		"/protected/resource",
//		groups.MustHave(func (c *gin.Context) (bool, groups.Groups) {
//			var s *Session = sessions.Get[MySession](c)
//			if s.User == nil {
//				return false, nil
//			}
//
//			return true, s.User.Groups
//		}, groups.OneOf("readResource", "writeResource"))
//
// Note that if you don't pass any OneOf or AllOf rule, so long as the session has a valid user (i.e. fn argument
// returns true as its first return value), the request will not be rejected.
func MustHave(fn func(*gin.Context) (authenticated bool, groups Groups), rule Rule, more ...Rule) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := (&rules{
			unauthorizedHandler: defaultUnauthorizedHandler,
			forbiddenHandler:    defaultForbiddenHandler,
		}).apply(rule, more...)

		ok, groups := fn(c)
		if !ok {
			r.unauthorizedHandler(c)
			return
		}

		if !r.test(groups) {
			r.forbiddenHandler(c)
			return
		}

		c.Next()
	}
}

// Groups is a string list, preferably a string set.
type Groups []string

// Test verifies that the user's groups satisfy the membership rules.
//
// Use AllOf and/or OneOf to describe how to authorise the user's groups.
//
// Usage:
//
//	// user must be able to read both payments and inventory.
//	Groups([]string{...}).Test(AllOf("can_read_payment", "can_read_inventory"))
//
//	// user must be able to read both payments and inventory, but write permissions implies read as well.
//	Groups([]string{...}).Test(OneOf("can_read_payment", "can_write_payment"), OneOf("can_read_inventory", "can_write_inventory"))
//
// This function ignores WithUnauthorizedHandler and WithForbiddenHandler settings since it is intended to be used
// outside of a gin request. If you don't pass any OneOf or AllOf rule, the function always returns true.
func (groups Groups) Test(rule Rule, more ...Rule) bool {
	return (&rules{}).apply(rule, more...).test(groups)
}

// Rule can only be either AllOf or OneOf.
type Rule func(*rules)

// AllOf adds a rule that the user must belong to all the groups specified here.
func AllOf(group string, more ...string) Rule {
	return func(r *rules) {
		if r.allOf == nil {
			r.allOf = map[string]bool{group: true}
		} else {
			r.allOf[group] = true
		}

		for _, g := range more {
			r.allOf[g] = true
		}
	}
}

// OneOf adds a rule that the user must belong to at least one of the groups specified here.
func OneOf(first, second string, more ...string) Rule {
	return func(r *rules) {
		groups := map[string]bool{first: true, second: true}
		for _, group := range more {
			groups[group] = true
		}

		r.oneOf = &node{
			groups: groups,
			next:   r.oneOf,
		}
	}
}

// WithUnauthorizedHandler can be used to customise the response when the session has no user.
//
// By default, [gin.Context.AbortWithStatus] is called passing http.StatusUnauthorized.
func WithUnauthorizedHandler(f gin.HandlerFunc) Rule {
	return func(opts *rules) {
		opts.unauthorizedHandler = f
	}
}

// WithForbiddenHandler can be used to customise the response when the session's user's groups do not satisfy the rules.
//
// By default, [gin.Context.AbortWithStatus] is called passing http.StatusForbidden.
func WithForbiddenHandler(f gin.HandlerFunc) Rule {
	return func(opts *rules) {
		opts.forbiddenHandler = f
	}
}

func defaultUnauthorizedHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}

func defaultForbiddenHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusForbidden)
}

// rules contains all the rules including response handlers.
type rules struct {
	allOf               map[string]bool
	oneOf               *node
	unauthorizedHandler func(*gin.Context)
	forbiddenHandler    func(*gin.Context)
}

// node is a mini linked list implementation for OneOf rules.
type node struct {
	groups map[string]bool
	next   *node
}

func (r *rules) apply(rule Rule, more ...Rule) *rules {
	rule(r)
	for _, fn := range more {
		fn(r)
	}

	return r
}

func (r *rules) test(groups []string) bool {
	// in the case of empty rules at the start, return true.
	if len(r.allOf) == 0 && r.oneOf == nil {
		return true
	}

	// the one-pass algorithm goes through the groups that user belongs to then checks off the list.
	// if any rules remain then user is not authorised.
	for _, group := range groups {
		delete(r.allOf, group)

		var p, n *node
		for n = r.oneOf; n != nil; p, n = n, n.next {
			if n.groups[group] {
				if p == nil {
					r.oneOf = n.next
					continue
				}

				p.next = n.next
			}
		}
	}

	return len(r.allOf) == 0 && r.oneOf == nil
}
