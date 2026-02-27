package badrequest

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/gin-metrics/internal"
	"github.com/rotisserie/eris"
)

// AbortWithErrorMessagef aborts the request with http.StatusBadRequest as status and the formatted string as
// message while pushing the wrapped error to context.
//
// The given error is wrapped using eris.Wrapf for logging so don't include the wrapped error inside the arguments.
// Similarly, the message that is returned to user is formatted using fmt.Sprintf so be mindful of what goes into the
// message.
//
// Usage:
//
//	// this will wrap the error and log it with context including the phone number, but the error message returned
//	// to user does not include the phone number.
//	badrequest.AbortWithErrorMessagef(c, eris.Wrapf(err, "parse phone number %q error", phoneNumber), "invalid phone number")
//
// If you don't have an error to log, use WithMessagef.
func AbortWithErrorMessagef(c *gin.Context, err error, format string, a ...any) *gin.Error {
	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": message,
		})

		return internal.LogErrorf(c, http.StatusBadRequest, eris.Wrapf(err, format, a...), "aborted with 400 Bad Request: %s", message)
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"status": http.StatusBadRequest,
	})

	return internal.LogErrorf(c, http.StatusBadRequest, eris.Wrapf(err, format, a...), "aborted with 400 Bad Request")
}

// AbortWithMessagef aborts the request with http.StatusBadRequest as status and the formatted string as message.
//
// The message (formatted with fmt.Sprintf) will be returned to user so be mindful of its contents.
func AbortWithMessagef(c *gin.Context, format string, a ...any) {
	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": message,
		})

		internal.Logf(c, http.StatusBadRequest, "aborted with 400 Bad Request: %s", message)
		return
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"status": http.StatusBadRequest,
	})

	internal.Logf(c, http.StatusBadRequest, "aborted with 400 Bad Request")
}
