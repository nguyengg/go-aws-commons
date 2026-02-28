// Package slog provides utilities to attach and retrieve slog.Logger instances from context.
//
// It also provides slog.Value implementations for JSON and error (with stack trace) types.
package slog

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

// WithContext attaches the given slog.Logger instance to the returned context.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, &loggerKey{}, logger)
}

// FromContext retrieves the slog.Logger instance that was attached with WithContext.
//
// If none is available, slog.Default is returned.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(&loggerKey{}).(*slog.Logger); ok {
		return logger
	}

	return slog.Default()
}

// UpdateContext retrieves the slog.Logger instance that was attached with WithContext, applies changes to that instance
// with fn, then attaches the new instance to the returned context.
//
// Useful if you need to retrieve a logger while also passing it to other callers via context.
func UpdateContext(ctx context.Context, fn func(logger *slog.Logger) *slog.Logger) (context.Context, *slog.Logger) {
	logger, ok := ctx.Value(&loggerKey{}).(*slog.Logger)
	if !ok {
		logger = slog.Default()
	}

	logger = fn(logger)
	return context.WithValue(ctx, &loggerKey{}, logger), logger
}
