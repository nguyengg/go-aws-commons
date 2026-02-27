// Package abort contains package-level methods to help abort a gin request using JSON response as well as logging.
//
// The JSON response will be in this format:
//
//	{
//		"status": 400 | 500 | ...
//		"message": "details about the error"
//	}
//
// Any method that returns gin.Error will have pushed the error to gin.Context via [gin.Context.Error] so that the
// metrics created with ginmetrics.Logger can log it.
//
// Any method whose name end with Wrapf will use eris.Wrapf so that you don't have to pre-wrap the error.
//
// All methods will attempt to log the abort attempt as well via ginmetrics.TryGetLogger.
package abort

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/gin-metrics/internal"
	"github.com/rotisserie/eris"
)

// WithStatusMessagef aborts the request with the given code as status and the formatted string as message.
//
// The message (formatted with fmt.Sprintf) will be returned to user so be mindful of its contents.
func WithStatusMessagef(c *gin.Context, code int, format string, a ...any) {
	// https://www.jetbrains.com/help/go/2023.3/formatting-strings.html wants these methods' names to end with f.

	if message := fmt.Sprintf(format, a...); message != "" {
		c.AbortWithStatusJSON(code, gin.H{
			"status":  code,
			"message": message,
		})

		internal.Logf(c, code, "aborted with %d %s: %s", code, http.StatusText(code), message)
		return
	}

	WithStatus(c, code)
}

// WithStatus is a variant of WithStatusMessagef that supplants a default http.StatusText message.
//
// Use this if you just want to use the default text for a specific status code, such as http.StatusForbidden
// ("Forbidden") or http.StatusUnauthorized ("Unauthorized").
func WithStatus(c *gin.Context, code int) {
	if message := http.StatusText(code); message != "" {
		c.AbortWithStatusJSON(code, gin.H{
			"status":  code,
			"message": message,
		})

		internal.Logf(c, code, "aborted with %d %s", code, message)
		return
	}

	c.AbortWithStatusJSON(code, gin.H{
		"status": code,
	})

	internal.Logf(c, code, "aborted with %d", code)
}

// Wrapf aborts the request with http.StatusInternalServerError as status and "Internal Server Error" as message.
//
// The message returned to user is always "Internal Server Error" so feel free to provide as much information about the
// error as possible.
//
// Use this when your handler runs into a server-fault error that should abort the request, you want to capture and log
// the error, but you do not want to report the details of that error to user. This is a variant of
// internalservererror.AbortWithErrorMessagef with a fixed message returned to user.
func Wrapf(c *gin.Context, err error, format string, a ...any) *gin.Error {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status":  http.StatusInternalServerError,
		"message": http.StatusText(http.StatusInternalServerError),
	})

	return internal.LogErrorf(c, http.StatusInternalServerError, eris.Wrapf(err, format, a...), "aborted with 500 Internal Server Error")
}
