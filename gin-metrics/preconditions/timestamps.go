package preconditions

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// IfModifiedSince parses request header "If-Modified-Since" and compares it against the given time.Time.
//
// The method returns two boolean values: ignored is true according to
// https://www.rfc-editor.org/rfc/rfc9110.html#name-if-modified-since, which can be presence of "If-None-Match" (doesn't
// matter its validity), invalid "If-Modified-Since", or t argument is zero v; isModifiedSince is then true only if
// t is strictly after the "If-Modified-Since" date.
func IfModifiedSince(c *gin.Context, t time.Time) (ignored, isModifiedSince bool) {
	if t.IsZero() {
		return true, false
	}

	m, ok := c.Get(ifModifiedSinceKey)
	if !ok {
		m = parseIfModifiedSince(c.Request.Header)
		c.Set(ifModifiedSinceKey, m)
	}

	return m.(sinceMatcher)(t)
}

// IfUnmodifiedSince parses request header "If-Since" and compares it against the given time.Time.
//
// The method returns two boolean values: ignored is true according to
// https://www.rfc-editor.org/rfc/rfc9110.html#name-if-unmodified-since, which can be presence of "If-Match" (doesn't
// matter its validity), invalid "If-Unmodified-Since", or t argument is zero v; isUnmodifiedSince is then true only
// if the "If-Unmodified-Since" date is strictly after t.
func IfUnmodifiedSince(c *gin.Context, t time.Time) (ignored, isUnmodifiedSince bool) {
	if t.IsZero() {
		return true, false
	}

	m, ok := c.Get(ifUnmodifiedSinceKey)
	if !ok {
		m = parseIfUnmodifiedSince(c.Request.Header)
		c.Set(ifUnmodifiedSinceKey, m)
	}

	return m.(sinceMatcher)(t)
}

func parseIfModifiedSince(header http.Header) sinceMatcher {
	if _, ok := header["If-None-Match"]; ok {
		return ignoredSinceMatcher
	}

	values, ok := header["If-Modified-Since"]
	if !ok || len(values) != 1 {
		return ignoredSinceMatcher
	}

	since, err := time.Parse(http.TimeFormat, values[0])
	if err != nil {
		return ignoredSinceMatcher
	}

	return func(t time.Time) (bool, bool) {
		return false, t.After(since)
	}
}

func parseIfUnmodifiedSince(header http.Header) sinceMatcher {
	if _, ok := header["If-Match"]; ok {
		return ignoredSinceMatcher
	}

	values, ok := header["If-Unmodified-Since"]
	if !ok || len(values) != 1 {
		return ignoredSinceMatcher
	}

	since, err := time.Parse(http.TimeFormat, values[0])
	if err != nil {
		return ignoredSinceMatcher
	}

	return func(t time.Time) (bool, bool) {
		return false, since.After(t)
	}
}

type sinceMatcher func(t time.Time) (bool, bool)

func ignoredSinceMatcher(_ time.Time) (bool, bool) {
	return true, false
}
