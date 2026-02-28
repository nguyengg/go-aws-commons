// Package slogging provides utilities to attach and retrieve slog.Logger instances from context.
//
// It also provides slog.Value implementations for JSON and error (with stack trace) types.
package slogging

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

// WithContext attaches the given slog.Logger instance to the returned context.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, &loggerKey{}, logger)
}

// Get retrieves the slog.Logger instance that was attached with WithContext.
//
// If none is available, slog.Default is returned.
func Get(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(&loggerKey{}).(*slog.Logger); ok {
		return logger
	}

	return slog.Default()
}

// TryGet is a variant of Get that returns (nil, false) if no instance was attached.
func TryGet(ctx context.Context) (*slog.Logger, bool) {
	logger, ok := ctx.Value(&loggerKey{}).(*slog.Logger)
	return logger, ok
}

// GetWith is a variant of Get that allows updating of the logger with attributes.
//
// [slog.Logger.With] is used to update the logger. You should pass any number slog.Attr as the args.
//
// The modified logger is returned along with the updated context.
func GetWith(ctx context.Context, args ...any) (context.Context, *slog.Logger) {
	logger, ok := ctx.Value(&loggerKey{}).(*slog.Logger)
	if !ok {
		logger = slog.Default()
	}

	if len(args) != 0 {
		logger = logger.With(args...)
		ctx = context.WithValue(ctx, &loggerKey{}, logger)
	}

	return ctx, logger
}

// UpdateContext is a variant of GetWith that receives a function to update the logger instead of attributes.
//
// The logger returned by fn is attached to context and returned.
//
// Useful if you need to update the logger with [slog.Logger.WithGroup], for example, instead of attributes.
func UpdateContext(ctx context.Context, fn func(logger *slog.Logger) *slog.Logger) (context.Context, *slog.Logger) {
	logger, ok := ctx.Value(&loggerKey{}).(*slog.Logger)
	if !ok {
		logger = slog.Default()
	}

	logger = fn(logger)
	return context.WithValue(ctx, &loggerKey{}, logger), logger
}
