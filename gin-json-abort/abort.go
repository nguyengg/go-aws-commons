// Package abort contains package-level methods to help abort a gin request using JSON response as well as logging.
//
// The JSON response will be in this format (see H):
//
//	{
//		"status": 400 | 500 | ...
//		"message": "details about the error"
//	}
//
// All methods will attempt to log the abort attempt as well via ginmetrics.TryGetLogger. Any method that returns
// gin.Error will have pushed the error to gin.Context via [gin.Context.Error] so that the metrics created with
// ginmetrics.Logger can log it, akin to [gin.Context.AbortWithError].
//
// Why the f suffix? https://www.jetbrains.com/help/go/2023.3/formatting-strings.html wants these methods' names to end
// with f to make use of Print-f like functionality.
package abort

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/gin-json-abort/internal"
)

// H is the struct used by all methods in this package to return JSON response.
type H struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
}

// WithStatusJSONf aborts the request with the given code as status and the formatted string as message.
//
// The message comes from fmt.Sprintf(format, a...) so be careful to not include any sensitive information.
func WithStatusJSONf(c *gin.Context, code int, format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	c.AbortWithStatusJSON(code, H{code, message})
	internal.Logf(c, code, message)
}

// WithStatus is a variant of WithStatusJSONf that supplants a default http.StatusText message.
//
// Use this if you just want to use the default text for a specific status code, such as http.StatusForbidden
// ("Forbidden") or http.StatusUnauthorized ("Unauthorized").
func WithStatus(c *gin.Context, code int) {
	message := http.StatusText(code)
	c.AbortWithStatusJSON(code, H{code, message})
	internal.Logf(c, code, message)
}

// WithErrorJSONf aborts the request with the given code as status and the formatted string as message while pushing
// the given error to context.
//
// The message comes from fmt.Sprintf(format, a...) so be careful to not include any sensitive information. The error is
// wrapped using eris.Wrapf(err, format, a...) so don't include the error in the format arguments; you can pre-wrap the
// error as well.
//
// Usage:
//
//	// this will wrap the error and log it with context including the phone number, but the error message returned
//	// to user does not include the phone number.
//	abort.WithErrorJSONf(c, 400, eris.Wrapf(err, "parse phone number %q error", phoneNumber), "invalid phone number")
//
//	// similarly, this will wrap the error and log it with context (bucket and key, and the error may contain metadata
//	// such as request Id from S3, etc.), but the error message returned to user will just read
//	// "s3 in us-west-2 is having trouble".
//	abort.WithErrorJSONf(c, 500, eris.Wrapf(err, "get s3://%s/%s error", bucket, key), "s3 in %s is having trouble", region)
func WithErrorJSONf(c *gin.Context, code int, err error, format string, a ...any) *gin.Error {
	message := fmt.Sprintf(format, a...)
	c.AbortWithStatusJSON(code, H{code, message})
	return internal.LogErrorf(c, code, err, message)
}

// WithError is a variant of WithErrorJSONf that always uses http.StatusInternalServerError for status and "Internal
// Server Error" for message.
//
// Use this when your handler runs into a server-fault error that should abort the request, you want to capture and log
// the error, but you do not want to report the details of that error to user. The message returned to user is always
// "Internal Server Error" so feel free to provide as much information about the error as possible, unlike
// WithErrorJSONf.
//
// Usage:
//
//	// this will wrap the error and log it with context (bucket and key, and the error may contain metadata such as
//	// request Id from S3, etc.), but the error message returned to user will always be "Internal Server Error".
//	abort.WithError(c, eris.Wrapf(err, "get s3://%s/%s error", bucket, key))
func WithError(c *gin.Context, err error) *gin.Error {
	c.AbortWithStatusJSON(
		http.StatusInternalServerError,
		H{http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)})
	return internal.LogErrorf(c, http.StatusInternalServerError, err, "")
}
