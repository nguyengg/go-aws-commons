// Package s3headers provide convenient methods to copy request headers from gin.Context to an S3 GetObject or
// HeadObject, and vice versa — from GetObject/HeadObject output to gin response.
//
// These gin request headers are copied into S3 input parameters:
//   - If-Match and If-None-Match
//   - If-Modified-Since and If-Unmodified-Since
//   - Range
//
// These attributes from S3 output parameters are copied into gin response headers:
//   - Cache-Control
//   - Content-Disposition
//   - Content-Encoding
//   - Content-Language
//   - Content-Length
//   - Content-Range
//   - ETag
//   - ExpiresString (preferred over deprecated Expires)
//   - LastModified
package s3headers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// CopyToGetObjectInput copies the conditional request headers from gin.Context into the given GetObject input.
func CopyToGetObjectInput(c *gin.Context, input *s3.GetObjectInput) *s3.GetObjectInput {
	header := c.Request.Header
	input.IfMatch = getIfMatch(header)
	input.IfModifiedSince = getIfModifiedSince(header)
	input.IfNoneMatch = getIfNoneMatch(header)
	input.IfUnmodifiedSince = getIfUnmodifiedSince(header)
	input.Range = getRange(header)
	return input
}

// CopyToHeadObjectInput copies the conditional request headers from gin.Context into the given HeadObject input.
func CopyToHeadObjectInput(c *gin.Context, input *s3.HeadObjectInput) *s3.HeadObjectInput {
	header := c.Request.Header
	input.IfMatch = getIfMatch(header)
	input.IfModifiedSince = getIfModifiedSince(header)
	input.IfNoneMatch = getIfNoneMatch(header)
	input.IfUnmodifiedSince = getIfUnmodifiedSince(header)
	input.Range = getRange(header)
	return input
}

// CopyFromGetObjectOutput parses response headers from the given GetObject output and sets them as response headers.
func CopyFromGetObjectOutput(c *gin.Context, output *s3.GetObjectOutput) {
	if output.CacheControl != nil {
		c.Header("Cache-Control", *output.CacheControl)
	}
	if output.ContentDisposition != nil {
		c.Header("Content-Disposition", *output.ContentDisposition)
	}
	if output.ContentEncoding != nil {
		c.Header("Content-Encoding", *output.ContentEncoding)
	}
	if output.ContentLanguage != nil {
		c.Header("Content-Language", *output.ContentLanguage)
	}
	if output.ContentLength != nil {
		c.Header("Content-Length", strconv.FormatInt(*output.ContentLength, 10))
	}
	if output.ContentRange != nil {
		c.Header("Content-Range", *output.ContentRange)
	}
	if output.ContentType != nil {
		c.Header("Content-Type", *output.ContentType)
	}
	if output.ETag != nil {
		c.Header("ETag", *output.ETag)
	}
	if output.ExpiresString != nil {
		c.Header("Expires", *output.ExpiresString)
	}
	if output.LastModified != nil {
		c.Header("Last-Modified", output.LastModified.Format(http.TimeFormat))
	}
}

// CopyFromHeadObjectOutput parses response headers from the given HeadObject output and sets them as response headers.
func CopyFromHeadObjectOutput(c *gin.Context, output *s3.HeadObjectOutput) {
	if output.CacheControl != nil {
		c.Header("Cache-Control", *output.CacheControl)
	}
	if output.ContentDisposition != nil {
		c.Header("Content-Disposition", *output.ContentDisposition)
	}
	if output.ContentEncoding != nil {
		c.Header("Content-Encoding", *output.ContentEncoding)
	}
	if output.ContentLanguage != nil {
		c.Header("Content-Language", *output.ContentLanguage)
	}
	if output.ContentLength != nil {
		c.Header("Content-Length", strconv.FormatInt(*output.ContentLength, 10))
	}
	if output.ContentRange != nil {
		c.Header("Content-Range", *output.ContentRange)
	}
	if output.ContentType != nil {
		c.Header("Content-Type", *output.ContentType)
	}
	if output.ETag != nil {
		c.Header("ETag", *output.ETag)
	}
	if output.ExpiresString != nil {
		c.Header("Expires", *output.ExpiresString)
	}
	if output.LastModified != nil {
		c.Header("Last-Modified", output.LastModified.Format(http.TimeFormat))
	}
}

func getIfMatch(header http.Header) *string {
	value := header.Get("If-Match")
	if value == "" {
		return nil
	}
	return &value
}

func getIfModifiedSince(header http.Header) *time.Time {
	value := header.Get("If-Modified-Since")
	if value == "" {
		return nil
	}

	t, err := http.ParseTime(value)
	if err != nil {
		return nil
	}

	return &t
}

func getIfNoneMatch(header http.Header) *string {
	value := header.Get("If-None-Match")
	if value == "" {
		return nil
	}
	return &value
}

func getIfUnmodifiedSince(header http.Header) *time.Time {
	value := header.Get("If-Unmodified-Since")
	if value == "" {
		return nil
	}

	t, err := http.ParseTime(value)
	if err != nil {
		return nil
	}

	return &t
}

func getRange(header http.Header) *string {
	value := header.Get("Range")
	if value == "" {
		return nil
	}
	return &value
}
