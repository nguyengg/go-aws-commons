package internal

import (
	"encoding/json"
	"fmt"
	"log/slog"

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
func LogErrorf(c *gin.Context, status int, err error, format string, a ...any) *gin.Error {
	logger, ok := ginmetrics.TryGetLogger(c)
	if !ok {
		return c.Error(err)
	}

	var level = slog.LevelInfo
	if status >= 500 && status < 600 {
		level = slog.LevelError
	}

	logger.LogAttrs(c, level, fmt.Sprintf(format, a...), slog.Any("error", slog.AnyValue(errorValue{err})))
	return c.Error(err)
}

// Logf is a helper method to log the given message using slog.
func Logf(c *gin.Context, status int, format string, a ...any) {
	logger, ok := ginmetrics.TryGetLogger(c)
	if !ok {
		return
	}

	var level = slog.LevelInfo
	if status >= 500 && status < 600 {
		level = slog.LevelError
	}

	logger.LogAttrs(c, level, fmt.Sprintf(format, a...))
}
