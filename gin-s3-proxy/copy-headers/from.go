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
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// FromGetObjectOutput parses response headers from the given GetObject output and sets them as response headers.
func FromGetObjectOutput(c *gin.Context, output *s3.GetObjectOutput) {
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

// FromHeadObjectOutput parses response headers from the given HeadObject output and sets them as response headers.
func FromHeadObjectOutput(c *gin.Context, output *s3.HeadObjectOutput) {
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

// FromPutObjectOutput parses response headers from the given PutObject output and sets them as response headers.
func FromPutObjectOutput(c *gin.Context, output *s3.PutObjectOutput) {
	if output.ETag != nil {
		c.Header("ETag", *output.ETag)
	}
}

// FromCompleteMultipartUploadOutput parses response headers from the given CompleteMultipartUpload output and sets them as response headers.
func FromCompleteMultipartUploadOutput(c *gin.Context, output *s3.CompleteMultipartUploadOutput) {
	if output.ETag != nil {
		c.Header("ETag", *output.ETag)
	}
}
