package metrics

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

const slogLoggerKey = "nguyengg/gin-metrics/slogLoggerKey"

// Slog returns the slog.Logger instance associated with context.
//
// If none is available, return slog.Default.
//
// Equivalent to zerolog.Ctx except only receive gin.Context.
func Slog(c *gin.Context) *slog.Logger {
	if logger, ok := c.Get(slogLoggerKey); ok {
		return logger.(*slog.Logger)
	}

	return slog.Default()
}
