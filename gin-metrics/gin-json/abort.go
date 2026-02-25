package ginjson

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AbortWithStatusMessage calls [gin.Context.AbortWithStatusJSON] writing code as "status" and the formatted string as
// "message".
//
// Note that the JSON body should not contain sensitive information that may help an attacker understand your system.
func AbortWithStatusMessage(c *gin.Context, code int, format string, a ...any) {
	c.AbortWithStatusJSON(code, gin.H{
		"status":  code,
		"message": fmt.Sprintf(format, a...),
	})
}

// AbortWithStatus is a variant of AbortWithStatusMessage that supplants a default http.StatusText message.
//
// Use this if you just want to use the default text for a specific status code, such as http.StatusForbidden
// ("Forbidden") or http.StatusUnauthorized ("Unauthorized").
func AbortWithStatus(c *gin.Context, code int) {
	if m := http.StatusText(code); m != "" {
		c.AbortWithStatusJSON(code, gin.H{
			"status":  code,
			"message": m,
		})
	} else {
		c.AbortWithStatusJSON(code, gin.H{
			"status": code,
		})
	}
}

// BadRequest is a variant of AbortWithStatusMessage for http.StatusBadRequest.
//
// Note that because the message will be returned to user, it MUST NOT contain sensitive information that may be used to
// craft further attacks to your system. As a reason, be mindful what error is being wrapped here.
func BadRequest(c *gin.Context, format string, a ...any) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"status":  http.StatusBadRequest,
		"message": fmt.Sprintf(format, a...),
	})
}

// Errorf aborts the request with http.StatusInternalServerError and default text, then calls [gin.Context.Error] using
// the formatted string.
//
// Use this when your handler runs into a server-fault error that should abort the request, you want to capture and log
// the error, but you do not want to report the details of that error to user. Feel free to wrap whatever error
// encountered here.
func Errorf(c *gin.Context, format string, a ...any) *gin.Error {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status":  http.StatusInternalServerError,
		"message": http.StatusText(http.StatusInternalServerError),
	})
	return c.Error(fmt.Errorf(format, a...))
}
