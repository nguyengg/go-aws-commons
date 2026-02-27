package abort

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	ginmetrics "github.com/nguyengg/go-aws-commons/gin-metrics"
	"github.com/rotisserie/eris"
)

// InternalServerError is used to access helper methods to abort requests with 500 Internal Server Errors.
var InternalServerError internalServerError

// WithErrorMessagef aborts the request with http.StatusInternalServerError as status and the formatted string as
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
func (_ internalServerError) WithErrorMessagef(c *gin.Context, err error, format string, a ...any) *gin.Error {
	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": message,
		})
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
		})
	}

	err = eris.Wrapf(err, format, a...)

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, "aborted with 500 Internal Server Error", slog.Any("error", errorValue{err}))
	}

	return c.Error(err)
}

// WithMessagef aborts the request with http.StatusBadRequest as status and the formatted string as message.
//
// The message (formatted with fmt.Sprintf) will be returned to user so be mindful of its contents.
func (_ internalServerError) WithMessagef(c *gin.Context, format string, a ...any) {
	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": message,
		})

		if logger, ok := ginmetrics.TryGetLogger(c); ok {
			logger.LogAttrs(c, slog.LevelInfo, fmt.Sprintf("aborted with 500 Internal Server Error: %s", message))
		}

		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status": http.StatusInternalServerError,
	})

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, fmt.Sprintf("aborted with 500 Internal Server Error"))
	}
}

type internalServerError struct {
}

// Wrapf aborts the request with http.StatusInternalServerError as status and "Internal Server Error" as message.
//
// The message returned to user is always "Internal Server Error" so feel free to provide as much information about the
// error as possible.
//
// Use this when your handler runs into a server-fault error that should abort the request, you want to capture and log
// the error, but you do not want to report the details of that error to user. This is a variant of
// [InternalServerError.WithErrorMessagef] with a fixed message returned to user.
func Wrapf(c *gin.Context, err error, format string, a ...any) *gin.Error {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status":  http.StatusInternalServerError,
		"message": http.StatusText(http.StatusInternalServerError),
	})

	err = eris.Wrapf(err, format, a...)

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, "aborted with 500 Internal Server Error", slog.Any("error", errorValue{err}))
	}

	return c.Error(err)
}
