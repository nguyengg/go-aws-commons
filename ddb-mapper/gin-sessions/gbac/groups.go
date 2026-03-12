// Package gbac provides Group-Based Access Control (GBAC) utilities.
//
// For all intents and purposes, groups and roles are interchangeable in this package.
package gbac

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/internal/groups"
)

// Groups is a string list, preferably a string set.
type Groups []string

// Test verifies that the user's groups satisfy the membership rules.
//
// Use AllOf and/or OneOf to describe how to authorise the user's groups. Each rule are implicitly AND with each other.
//
// Usage:
//
//	// user must be able to read both payments and inventory.
//	Groups([]string{...}).Test(AllOf("can_read_payment", "can_read_inventory"))
//
//	// user must be able to read both payments and inventory, but write permissions implies read as well.
//	Groups([]string{...}).Test(OneOf("can_read_payment", "can_write_payment"), OneOf("can_read_inventory", "can_write_inventory"))
func (g Groups) Test(rule Rule, more ...Rule) bool {
	return (&groups.Rules{}).Apply(rule, more...).Test(g)
}

// Rule can only be either AllOf or OneOf.
type Rule = groups.Rule

// AllOf adds a rule that the user must belong to all the groups specified here.
func AllOf(group string, more ...string) Rule {
	return groups.AllOf(group, more...)
}

// OneOf adds a rule that the user must belong to at least one of the groups specified here.
func OneOf(first, second string, more ...string) Rule {
	return groups.OneOf(first, second, more...)
}

// WithUnauthorizedHandler can be used to customise the response when the session has no user.
//
// By default, [gin.Context.AbortWithStatus] is called passing http.StatusUnauthorized.
func WithUnauthorizedHandler(f gin.HandlerFunc) Rule {
	return groups.WithUnauthorizedHandler(f)
}

// WithForbiddenHandler can be used to customise the response when the session's user's groups do not satisfy the rules.
//
// By default, [gin.Context.AbortWithStatus] is called passing http.StatusForbidden.
func WithForbiddenHandler(f gin.HandlerFunc) Rule {
	return groups.WithForbiddenHandler(f)
}
