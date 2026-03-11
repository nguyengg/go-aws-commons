package groups

import "github.com/gin-gonic/gin"

type Rules struct {
	AllOf               map[string]bool
	OneOf               *Node
	UnauthorizedHandler func(*gin.Context)
	ForbiddenHandler    func(*gin.Context)
}

type Rule func(*Rules)

type Node struct {
	Groups map[string]bool
	Next   *Node
}

func AllOf(group string, more ...string) Rule {
	return func(r *Rules) {
		if r.AllOf == nil {
			r.AllOf = map[string]bool{group: true}
		} else {
			r.AllOf[group] = true
		}

		for _, g := range more {
			r.AllOf[g] = true
		}
	}
}

// WithUnauthorizedHandler can be used to customise the response when the session has no user.
//
// By default, [gin.Context.AbortWithStatus] is called passing http.StatusUnauthorized.
func WithUnauthorizedHandler(f gin.HandlerFunc) Rule {
	return func(opts *Rules) {
		opts.UnauthorizedHandler = f
	}
}

// WithForbiddenHandler can be used to customise the response when the session's user's groups do not satisfy the rules.
//
// By default, [gin.Context.AbortWithStatus] is called passing http.StatusForbidden.
func WithForbiddenHandler(f gin.HandlerFunc) Rule {
	return func(opts *Rules) {
		opts.ForbiddenHandler = f
	}
}

func OneOf(first, second string, more ...string) Rule {
	return func(r *Rules) {
		groups := map[string]bool{first: true, second: true}
		for _, group := range more {
			groups[group] = true
		}

		r.OneOf = &Node{
			Groups: groups,
			Next:   r.OneOf,
		}
	}
}

func (r *Rules) Apply(rule Rule, more ...Rule) *Rules {
	rule(r)
	for _, fn := range more {
		fn(r)
	}

	return r
}

func (r *Rules) Test(groups []string) bool {
	// in the case of empty rules at the start, return true.
	if len(r.AllOf) == 0 && r.OneOf == nil {
		return true
	}

	// the one-pass algorithm goes through the groups that user belongs to then checks off the list.
	// if any rules remain then user is not authorised.
	for _, group := range groups {
		delete(r.AllOf, group)

		var p, n *Node
		for n = r.OneOf; n != nil; p, n = n, n.Next {
			if n.Groups[group] {
				if p == nil {
					r.OneOf = n.Next
					continue
				}

				p.Next = n.Next
			}
		}
	}

	return len(r.AllOf) == 0 && r.OneOf == nil
}
