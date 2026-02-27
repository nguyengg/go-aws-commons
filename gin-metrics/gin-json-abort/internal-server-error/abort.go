package internalservererror

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/gin-metrics/internal"
	"github.com/rotisserie/eris"
)

// AbortWithErrorMessagef aborts the request with http.StatusInternalServerError as status and the formatted string as
// message while pushing the wrapped error to context.
//
// The given error is wrapped using eris.Wrapf for logging so don't include the wrapped error inside the arguments.
// Similarly, the message that is returned to user is formatted using fmt.Sprintf so be mindful of what goes into the
// message.
//
// Usage:
//
//	// this will wrap the error and log it with context (such as request Id from S3, etc.), but the error message
//	// returned to user will just read "s3 in us-west-2 is having trouble".
//	abort.InternalServerError.WithErrorMessagef(c, eris.Wrapf(err, "getObject s3://%s/%s error", bucket, key), "s3 in %s having trouble", region)
//
// If you don't have an error to log, use WithMessagef.
func AbortWithErrorMessagef(c *gin.Context, err error, format string, a ...any) *gin.Error {
	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": message,
		})

		return internal.LogErrorf(c, http.StatusInternalServerError, eris.Wrapf(err, format, a...), "aborted with 500 Internal Server Error: %s", message)
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status": http.StatusInternalServerError,
	})

	return internal.LogErrorf(c, http.StatusInternalServerError, eris.Wrapf(err, format, a...), "aborted with 500 Internal Server Error")
}

// AbortWithMessagef aborts the request with http.StatusInternalServerError as status and the formatted string as
// message.
//
// The message (formatted with fmt.Sprintf) will be returned to user so be mindful of its contents.
func AbortWithMessagef(c *gin.Context, format string, a ...any) {
	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": message,
		})

		internal.Logf(c, http.StatusInternalServerError, "aborted with 500 Internal Server Error: %s", message)
		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status": http.StatusInternalServerError,
	})

	internal.Logf(c, http.StatusInternalServerError, "aborted with 500 Internal Server Error")
}
