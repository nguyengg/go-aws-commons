// Package copyheaders provide convenient methods to copy request headers from gin.Context to an S3 input (e.g.
// GetObjectInput) and vice versa — from GetObjectOutput to gin response.
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
package copyheaders

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// ToGetObjectInput copies the conditional request headers from gin.Context into the given GetObject input.
func ToGetObjectInput(c *gin.Context, input *s3.GetObjectInput) *s3.GetObjectInput {
	header := c.Request.Header
	input.IfMatch = getIfMatch(header)
	input.IfModifiedSince = getIfModifiedSince(header)
	input.IfNoneMatch = getIfNoneMatch(header)
	input.IfUnmodifiedSince = getIfUnmodifiedSince(header)
	input.Range = getRange(header)
	return input
}

// ToHeadObjectInput copies the conditional request headers from gin.Context into the given HeadObject input.
func ToHeadObjectInput(c *gin.Context, input *s3.HeadObjectInput) *s3.HeadObjectInput {
	header := c.Request.Header
	input.IfMatch = getIfMatch(header)
	input.IfModifiedSince = getIfModifiedSince(header)
	input.IfNoneMatch = getIfNoneMatch(header)
	input.IfUnmodifiedSince = getIfUnmodifiedSince(header)
	input.Range = getRange(header)
	return input
}

// ToDeleteObjectInput copies the conditional request headers from gin.Context into the given DeleteObject input.
func ToDeleteObjectInput(c *gin.Context, input *s3.DeleteObjectInput) *s3.DeleteObjectInput {
	header := c.Request.Header
	input.IfMatch = getIfMatch(header)
	return input
}

// ToPutObjectInput copies the conditional request headers from gin.Context into the given PutObject input.
func ToPutObjectInput(c *gin.Context, input *s3.PutObjectInput) *s3.PutObjectInput {
	header := c.Request.Header
	input.IfMatch = getIfMatch(header)
	input.IfNoneMatch = getIfNoneMatch(header)
	return input
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
