// Package preconditions provides helper methods to parse and compare conditional headers such as If-Match,
// If-None-Match, If-Modified-Since, and If-Unmodified-Since.
package preconditions

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// New returns the preconditions middleware that will parse the request precondition headers for use.
//
// See https://www.rfc-editor.org/rfc/rfc9110.html#name-preconditions.
//
// Specifically, only these headers are parsed:
//  1. If-Match
//  2. If-None-Match
//  3. If-Unmodified-Since
//  4. If-Modified-Since
//
// TODO add support for If-Range once I understand what it does :D.
//
// You don't have to set up the middleware; you can directly use IfMatch, IfNoneMatch, etc. to parse the request headers
// on-the-fly too. But if you want a middleware to validate those headers for you, this middleware can abort the request
// before your handler is run which can shave off some time actually computing the ETag of the resource being acted
// upon. It also caches the parsing of those headers for you.
func New(optFns ...func(cfg *Config)) gin.HandlerFunc {
	cfg := &Config{}
	for _, fn := range optFns {
		fn(cfg)
	}

	return cfg.handle
}

// Config customises the preconditions middleware.
type Config struct {
	// RequireIfMatch will force the request to have a valid If-Match request header.
	//
	// Useful for PUT or POST requests that must specify the If-Match request header to avoid "lost updates". If the
	// request does not have a valid If-Match header, 400 Bad Request is returned by the middleware right away.
	//
	// Your handler should use IfMatch to compare against the request header.
	RequireIfMatch func(c *gin.Context) bool
}

const (
	ifMatchKey           = "github.com/nguyengg/go-aws-commons/gin-metrics/if-matches"
	ifNoneMatchKey       = "github.com/nguyengg/go-aws-commons/gin-metrics/if-none-matches"
	ifModifiedSinceKey   = "github.com/nguyengg/go-aws-commons/gin-metrics/if-modified-since"
	ifUnmodifiedSinceKey = "github.com/nguyengg/go-aws-commons/gin-metrics/if-unmodifie-since"
)

func (cfg *Config) handle(c *gin.Context) {
	header := c.Request.Header

	var (
		hasIfMatch     bool
		hasIfNoneMatch bool
	)

	if _, hasIfMatch = header["If-Match"]; hasIfMatch {
		if m, exists, err := parseIfMatch(header); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
			return
		} else if !exists && cfg.RequireIfMatch != nil && cfg.RequireIfMatch(c) {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("If-Match header is required")).SetType(gin.ErrorTypePublic)
			return
		} else {
			c.Set(ifMatchKey, m)
		}
	} else {
		c.Set(ifModifiedSinceKey, parseIfModifiedSince(header))
	}

	if _, hasIfNoneMatch = header["If-None-Match"]; hasIfNoneMatch {
		if m, exists, err := parseIfNoneMatch(header); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
			return
		} else if exists {
			c.Set(ifNoneMatchKey, m)
		}
	} else {
		c.Set(ifUnmodifiedSinceKey, parseIfUnmodifiedSince(header))
	}

	c.Next()
}
