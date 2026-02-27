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
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	ginmetrics "github.com/nguyengg/go-aws-commons/gin-metrics"
)

// WithStatusMessagef aborts the request with the given code as status and the formatted string as message.
//
// The message (formatted with fmt.Sprintf) will be returned to user so be mindful of its contents.
func WithStatusMessagef(c *gin.Context, code int, format string, a ...any) {
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
