package ginjson

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	ginmetrics "github.com/nguyengg/go-aws-commons/gin-metrics"
	"github.com/rotisserie/eris"
)

// AbortWithError aborts the request with JSON response containing http.StatusInternalServerError as status and
// default http.StatusText as message, then passes [gin.Context.Error] the given error and returning its result.
//
// Use this when your handler runs into a server-fault error that should abort the request, you want to capture and log
// the error, but you do not want to report the details of that error to user. Feel free to use fmt.Errorf to wrap
// whatever additional information is needed here.
func AbortWithError(c *gin.Context, err error) *gin.Error {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status":  http.StatusInternalServerError,
		"message": http.StatusText(http.StatusInternalServerError),
	})

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, "aborted with 500 Internal Server Error", slog.Any("error", errorValue{eris.Wrap(err, err.Error())}))
	}

	return c.Error(err)
}

// AbortWithStatusf aborts the request with JSON response containing given code as status and the formatted string as
// message.
//
// Note that the JSON body should not contain sensitive information that may help an attacker understand your system.
// fmt.Sprintf is used to format the "message" so don't use %w verb.
func AbortWithStatusf(c *gin.Context, code int, format string, a ...any) {
	// https://www.jetbrains.com/help/go/2023.3/formatting-strings.html wants these methods' names to end with f.

	message := fmt.Sprintf(format, a...)
	c.AbortWithStatusJSON(code, gin.H{
		"status":  code,
		"message": message,
	})

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, fmt.Sprintf("aborted with %d %s: %s", code, http.StatusText(code), message))
	}
}

// AbortWithStatus is a variant of AbortWithStatusf that supplants a default http.StatusText message.
//
// Use this if you just want to use the default text for a specific status code, such as http.StatusForbidden
// ("Forbidden") or http.StatusUnauthorized ("Unauthorized").
func AbortWithStatus(c *gin.Context, code int) {
	if message := http.StatusText(code); message != "" {
		c.AbortWithStatusJSON(code, gin.H{
			"status":  code,
			"message": message,
		})

		if logger, ok := ginmetrics.TryGetLogger(c); ok {
			logger.LogAttrs(c, slog.LevelInfo, fmt.Sprintf("aborted with %d %s", code, message))
		}

		return
	}

	c.AbortWithStatusJSON(code, gin.H{
		"status": code,
	})

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, fmt.Sprintf("aborted with %d", code))
	}
}

// BadRequestf is a variant of AbortWithStatusf for http.StatusBadRequest specifically.
//
// The formatted string (using fmt.Sprintf) will be returned to user as error message. As a result, it MUST NOT contain
// sensitive information.
//
// Use BadRequestErrorfc if you have an error that you want to log with stack trace. Use this method if you don't have
// an external error that needs logging, or if you don't need a stack trace for it.
func BadRequestf(c *gin.Context, format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"status":  http.StatusBadRequest,
		"message": message,
	})

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, fmt.Sprintf("aborted with 400 Bad Request: %s", message))
	}
}

// BadRequestWrapf is a combination of AbortWithError and BadRequestf that will call [gin.Context.Error] on the
// wrapped error.
//
// The formatted string (using fmt.Sprintf) will be returned to user as error message. As a result, it MUST NOT contain
// sensitive information.
//
// Use this method if you have an error that you want to log with stack trace. Use BadRequestf if you don't have an
// external error that needs logging, or if you don't need a stack trace for it.
func BadRequestWrapf(c *gin.Context, err error, format string, a ...any) *gin.Error {
	message := fmt.Sprintf(format, a...)
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"status":  http.StatusBadRequest,
		"message": message,
	})

	err = eris.Wrapf(err, format, a...)

	if logger, ok := ginmetrics.TryGetLogger(c); ok {
		logger.LogAttrs(c, slog.LevelInfo, "aborted with 400 Bad Request", slog.Any("error", errorValue{err}))
	}

	return c.Error(err)
}
