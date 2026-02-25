package caching

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Headers sets the caching headers (Cache-Control, ETag, and/or Last-Modified) retrieved from the given obj.
//
// Objects should implement HasCacheControl, HasETag, and/or HasLastModified for this method to do any work.
func Headers(c *gin.Context, obj any) {
	if v, ok := obj.(HasCacheControl); ok {
		if s := v.GetCacheControl(); s != "" {
			c.Header("Cache-Control", s)
		}
	}

	if v, ok := obj.(HasETag); ok {
		if s := v.GetETag(); s != "" {
			c.Header("ETag", s)
		}
	}

	if v, ok := obj.(HasLastModified); ok {
		if t := v.GetLastModified(); !t.IsZero() {
			c.Header("Last-Modified", t.UTC().Format(http.TimeFormat))
		}
	}
}

// HasCacheControl implements GetCacheControl for objects that should be returned with response header "Cache-Control".
type HasCacheControl interface {
	// GetCacheControl returns the "Cache-Control" response header.
	GetCacheControl() string
}

// HasETag implements GetETag for objects that should be returned with response header "ETag".
type HasETag interface {
	// GetETag should return a valid ETag, which is in format "<etag_value>" (with quotes) or W/"<etag_value>" for weak
	// ETags.
	//
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/ETag.
	//
	// Note: Headers does not care if the returned value is valid or not, just that it's not an empty string.
	GetETag() string
}

// HasLastModified implements GetLastModified for objects that should be returned with response header "Last-Modified".
type HasLastModified interface {
	GetLastModified() time.Time
}
