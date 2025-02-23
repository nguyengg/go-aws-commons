package ginadapter

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AbortJSON aborts the current request with http.StatusInternalServerError and returns a generic JSON error message.
//
// The error message will have this format;
//
//	{
//		"status": 500,
//		"message": "Internal Server Error"
//	}
func AbortJSON(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status":  http.StatusInternalServerError,
		"message": http.StatusText(http.StatusInternalServerError),
	})
}

// AbortWithStatusJSON aborts the current request with the given status and the message comes from the default status
// text for the message.
//
// The error message will have this format;
//
//	{
//		"status": status,
//		"message": http.StatusText(status)
//	}
func AbortWithStatusJSON(c *gin.Context, status int) {
	c.AbortWithStatusJSON(status, gin.H{
		"status":  status,
		"message": http.StatusText(status),
	})
}

// AbortWithStatusJSONf aborts the current request with the given status and formatted message.
//
// The error message will have this format;
//
//	{
//		"status": status,
//		"message": fmt.Sprintf(format, a...)
//	}
func AbortWithStatusJSONf(c *gin.Context, status int, format string, a ...any) {
	c.AbortWithStatusJSON(status, gin.H{
		"status":  status,
		"message": fmt.Sprintf(format, a...),
	})
}

// JSONErrorHandler is a middleware that makes sure errors are always returned to caller as JSON content.
//
// Useful if your handler is handling API endpoints which are returning JSON responses already. The request chain must
// have been aborted for the middleware to take action. If status code is not explicitly set, the middleware will set it
// to http.StatusInternalServerError. If response body is already written or the request has no errors (which implies
// the request has not been aborted), the middleware will not write any JSON content.
//
// Otherwise, the JSON error message looks like this:
//
//	{
//		"status": 500|400|...,
//		"message": "message describing the error"
//	}
//
// If the error type is gin.ErrorTypeBind or gin.ErrorTypePublic, its message will be returned as the "message" field.
// Any other error types will be hidden with a default message retrieved from http.StatusText.
func JSONErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if !c.IsAborted() || c.Writer.Written() {
			return
		}

		err := c.Errors.Last()
		if err == nil {
			return
		}

		status := c.Writer.Status()
		if status == 0 {
			status = http.StatusInternalServerError
		}

		switch err.Type {
		case gin.ErrorTypeBind, gin.ErrorTypePublic:
			err.MarshalJSON()
			c.JSON(status, gin.H{
				"status":  status,
				"message": err.Error(),
			})
		default:
			c.JSON(status, gin.H{
				"status":  status,
				"message": http.StatusText(status),
			})
		}
	}
}
