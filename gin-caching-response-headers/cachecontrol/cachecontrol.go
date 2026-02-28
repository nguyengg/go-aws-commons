// Package cachecontrol provides a declarative way to create the Cache-Control directives.
package cachecontrol

import (
	"strconv"
	"strings"
	"time"
)

// Directive describes a Cache-Control response directive.
type Directive string

// MaxAge creates a "max-age" Cache-Control response directive.
func MaxAge(d time.Duration) Directive {
	return Directive(duration("max-age", d))
}

// SMaxAge creates an "s-maxage" Cache-Control response directive.
func SMaxAge(d time.Duration) Directive {
	return Directive(duration("s-maxage", d))
}

// StaleWhileRevalidate creates a "stale-while-revalidate" Cache-Control response directive.
func StaleWhileRevalidate(d time.Duration) Directive {
	return Directive(duration("stale-while-revalidate", d))
}

// StaleIfError creates a "stale-if-error" Cache-Control response directive.
func StaleIfError(d time.Duration) Directive {
	return Directive(duration("stale-if-error", d))
}

// Additional plain-text Cache-Control response directives.
const (
	NoCache         Directive = "no-cache"
	NoStore                   = "no-store"
	NoTransform               = "no-transform"
	MustRevalidate            = "must-revalidate"
	ProxyRevalidate           = "proxy-revalidate"
	MustUnderstand            = "must-understand"
	Private                   = "private"
	Public                    = "public"
	Immutable                 = "immutable"
)

// Join creates a new Cache-Control value from the given list of response directives.
// The list will be joined by separator ", ".
func Join(directives ...Directive) string {
	switch len(directives) {
	case 0:
		return ""
	case 1:
		return string(directives[0])
	case 2:
		return string(directives[0]) + ", " + string(directives[1])
	case 3:
		return string(directives[0]) + ", " + string(directives[1]) + ", " + string(directives[2])
	default:
		elems := make([]string, len(directives))
		for i, d := range directives {
			elems[i] = string(d)
		}
		return strings.Join(elems, ", ")
	}
}

// Creates a "key=value" pair where value is the second part of the given duration.
func duration(key string, duration time.Duration) string {
	return key + "=" + strconv.FormatInt(int64(duration/time.Second), 10)
}
