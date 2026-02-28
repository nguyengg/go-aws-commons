package ginmetrics

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/slogging"
)

// GetLogger returns the slog.Logger instance associated with context.
//
// If none is available, return slog.Default. The logger can be modified further using withFns; any changes here will be
// attached back to gin.Context, and the modified logger is also returned.
//
// You should always use GetLogger and its variants when retrieving from gin.Context. If you must use slogging.Get, be
// sure to use slogging.Get(c.Request.Context) unless you have enabled [gin.Engine.ContextWithFallback].
func GetLogger(c *gin.Context, withFns ...func(*slog.Logger) *slog.Logger) *slog.Logger {
	ctx := c.Request.Context()
	logger, ok := slogging.TryGet(ctx)
	if !ok {
		logger = slog.Default()
	}

	if len(withFns) != 0 {
		for _, fn := range withFns {
			logger = fn(logger)
		}

		c.Request = c.Request.WithContext(slogging.WithContext(ctx, logger))
	}

	return logger
}

// GetLoggerWith returns the slog.Logger instance associated with context.
//
// If none is available, return slog.Default. The logger can be modified further using args which will be passed to
// slog.With; the modified logger is attached back to gin.Context and returned.
//
// You should always use GetLogger and its variants when retrieving from gin.Context. If you must use slogging.Get, be
// sure to use slogging.Get(c.Request.Context) unless you have enabled [gin.Engine.ContextWithFallback].
func GetLoggerWith(c *gin.Context, args ...any) *slog.Logger {
	ctx := c.Request.Context()
	logger, ok := slogging.TryGet(ctx)
	if !ok {
		logger = slog.Default()
	}

	if len(args) != 0 {
		logger = logger.With(args...)
		c.Request = c.Request.WithContext(slogging.WithContext(ctx, logger))
	}

	return logger
}

// TryGetLogger is a variant of GetLogger that will return (nil, false) if no slog.Logger instance is attached to
// context.
//
// You should always use GetLogger and its variants when retrieving from gin.Context. If you must use slogging.Get, be
// sure to use slogging.Get(c.Request.Context) unless you have enabled [gin.Engine.ContextWithFallback].
func TryGetLogger(c *gin.Context) (*slog.Logger, bool) {
	return slogging.TryGet(c.Request.Context())
}
