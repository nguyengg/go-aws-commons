package abort

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	ginmetrics "github.com/nguyengg/go-aws-commons/gin-metrics"
	"github.com/rotisserie/eris"
)

// BadRequest is used to access helper methods to abort requests with 400 Bad Request.
var BadRequest badRequest

// WithErrorMessagef aborts the request with http.StatusBadRequest as status and the formatted string as message while
// pushing the wrapped error to context.
//
// The given error is wrapped using eris.Wrapf for logging so don't include the wrapped error inside the arguments.
// Similarly, the message that is returned to user is formatted using fmt.Sprintf so be mindful of what goes into the
// message.
//
// Usage:
//
//	// this will wrap the error and log it with context including the phone number, but the error message returned
//	// to user does not include the phone number.
//	abort.BadRequest.WithErrorMessagef(c, eris.Wrapf(err, "parse phone number %q error", phoneNumber), "invalid phone number")
//
// If you don't have an error to log, use WithMessagef.
func (_ badRequest) WithErrorMessagef(c *gin.Context, err error, format string, a ...any) *gin.Error {
	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": message,
		})
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
		})
	}

	err = eris.Wrapf(err, format, a...)

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, "aborted with 400 Bad Request", slog.Any("error", errorValue{err}))
	}

	return c.Error(err)
}

// WithMessagef aborts the request with http.StatusBadRequest as status and the formatted string as message.
//
// The message (formatted with fmt.Sprintf) will be returned to user so be mindful of its contents.
func (_ badRequest) WithMessagef(c *gin.Context, format string, a ...any) {
	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": message,
		})

		if logger, ok := ginmetrics.TryGetLogger(c); ok {
			logger.LogAttrs(c, slog.LevelInfo, fmt.Sprintf("aborted with 400 Bad Request: %s", message))
		}

		return
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"status": http.StatusBadRequest,
	})

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, fmt.Sprintf("aborted with 400 Bad Request"))
	}
}

type badRequest struct {
}
