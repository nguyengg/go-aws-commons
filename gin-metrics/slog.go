package ginmetrics

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

const slogLoggerKey = "nguyengg/gin-metrics/slogLoggerKey"

// Deprecated: use GetLogger for a more meaningful name.
func Slog(c *gin.Context) *slog.Logger {
	return GetLogger(c)
}

// GetLogger returns the slog.Logger instance associated with context.
//
// If none is available, return slog.Default. The logger can be modified further using withFns; any changes here will be
// attached back to gin.Context, and the modified logger is also returned.
//
// Equivalent to zerolog.Ctx except only receive gin.Context.
func GetLogger(c *gin.Context, withFns ...func(*slog.Logger) *slog.Logger) *slog.Logger {
	var logger *slog.Logger
	if v, ok := c.Get(slogLoggerKey); ok {
		logger = v.(*slog.Logger)
	} else {
		logger = slog.Default()
	}

	if len(withFns) != 0 {
		for _, fn := range withFns {
			logger = fn(logger)
		}

		c.Set(slogLoggerKey, logger)
	}

	return logger
}
