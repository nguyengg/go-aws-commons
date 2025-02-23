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
