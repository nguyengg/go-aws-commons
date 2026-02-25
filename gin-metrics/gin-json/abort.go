package ginjson

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
	return c.Error(err)
}

// AbortWithStatusf aborts the request with JSON response containing given code as status and the formatted string as
// message.
//
// Note that the JSON body should not contain sensitive information that may help an attacker understand your system.
// fmt.Sprintf is used to format the "message" so don't use %w verb.
func AbortWithStatusf(c *gin.Context, code int, format string, a ...any) {
	// https://www.jetbrains.com/help/go/2023.3/formatting-strings.html wants these methods' names to end with f.
	c.AbortWithStatusJSON(code, gin.H{
		"status":  code,
		"message": fmt.Sprintf(format, a...),
	})
}

// AbortWithStatus is a variant of AbortWithStatusf that supplants a default http.StatusText message.
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

// BadRequestf is a variant of AbortWithStatusf for http.StatusBadRequest specifically.
//
// Note that because the message will be returned to user, it MUST NOT contain sensitive information that may be used to
// craft further attacks to your system. As a reason, be mindful what error is being printed here.
//
// fmt.Sprintf is used to format the "message" so don't use %w verb.
func BadRequestf(c *gin.Context, format string, a ...any) {
	// https://www.jetbrains.com/help/go/2023.3/formatting-strings.html wants these methods' names to end with f.
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"status":  http.StatusBadRequest,
		"message": fmt.Sprintf(format, a...),
	})
}
