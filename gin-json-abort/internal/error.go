package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	ginmetrics "github.com/nguyengg/go-aws-commons/gin-metrics"
	"github.com/rotisserie/eris"
)

type errorValue struct {
	err error
}

func (e errorValue) String() string {
	return eris.ToString(e.err, true)
}

func (e errorValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(eris.ToJSON(e.err, true))
}

// LogErrorf is a helper method to log the given error and message using slog.
func LogErrorf(c *gin.Context, status int, err error, msg string) *gin.Error {
	logger, ok := ginmetrics.TryGetLogger(c)
	if !ok {
		return c.Error(err)
	}

	var level = slog.LevelInfo
	if status >= 500 && status < 600 {
		level = slog.LevelError
	}

	if msg != "" {
		logger.LogAttrs(c, level, fmt.Sprintf("aborted with %s: %s", statusCode(status), msg), slog.Any("error", slog.AnyValue(errorValue{eris.Wrap(err, msg)})))
	} else {
		logger.LogAttrs(c, level, fmt.Sprintf("aborted with %s", statusCode(status)), slog.Any("error", slog.AnyValue(errorValue{eris.Wrap(err, msg)})))
	}

	return c.Error(err)
}

// Logf is a helper method to log the given message using slog.
func Logf(c *gin.Context, status int, msg string) {
	logger, ok := ginmetrics.TryGetLogger(c)
	if !ok {
		return
	}

	var level = slog.LevelInfo
	if status >= 500 && status < 600 {
		level = slog.LevelError
	}

	if msg != "" {
		logger.LogAttrs(c, level, fmt.Sprintf("aborted with %s: %s", statusCode(status), msg))
	} else {
		logger.LogAttrs(c, level, fmt.Sprintf("aborted with %s", statusCode(status)))
	}
}

type statusCode int

func (c statusCode) Format(f fmt.State, _ rune) {
	if m := http.StatusText(int(c)); m != "" {
		_, _ = fmt.Fprintf(f, "%d %s", int(c), m)
	} else {
		_, _ = fmt.Fprintf(f, "%d", int(c))
	}
}
