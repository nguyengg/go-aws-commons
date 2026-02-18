package metrics

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/rs/zerolog"
)

type metricsLogger interface {
	Log(ctx context.Context, m *Metrics) error
}

type jsonLogger struct {
	io.Writer
}

func (l jsonLogger) Log(ctx context.Context, m *Metrics) error {
	err := json.NewEncoder(l.Writer).Encode(m)

	if err == nil {
		_, err = l.Write([]byte("\n"))
	}

	return err
}

// SlogOptions customises how LogWithSlog works.
type SlogOptions struct {
	// Logger is the slog.Logger instance to use.
	//
	// Default to slog.Default.
	Logger *slog.Logger

	// Level is the log level to use.
	//
	// By default, the log level is dynamic. If the Metrics instance indicates an error state with non-zero fault
	// and/or panicked counter, slog.LevelError is used. Otherwise, slog.LevelInfo is used.
	Level *slog.Level

	// Msg is the message to be used with the log.
	Msg string

	// Group is the name of the slog.Group to place the Metrics content.
	//
	// By default, the Metrics instance is logged as top-level fields (no group).
	//
	// Equivalent to ZerologOptions.Dict.
	Group string

	// NoCustomFormatter disables specialised formatting for timestamps and durations.
	//
	// By default, to match LogJSON, the start time is logged as epoch millisecond, the end time using RFC1123, and
	// all time.Duration using FormatDuration. To disable this behaviour and rely on slog's own formatters, set
	// NoCustomFormatter to true.
	NoCustomFormatter bool
}

func (opts *SlogOptions) Log(ctx context.Context, m *Metrics) error {
	attrs := m.attrs(opts.NoCustomFormatter)

	var logLevel = slog.LevelInfo
	if opts.Level != nil {
		logLevel = *opts.Level
	} else if c := m.counters[CounterKeyFault]; c.isPositive() {
		logLevel = slog.LevelError
	} else if c := m.counters[CounterKeyPanicked]; c.isPositive() {
		logLevel = slog.LevelError
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	if opts.Group != "" {
		logger.LogAttrs(ctx, logLevel, opts.Msg, slog.GroupAttrs(opts.Group, attrs...))
	} else {
		logger.LogAttrs(ctx, logLevel, opts.Msg, attrs...)
	}

	return nil
}

// ZerologOptions customises how LogWithZerolog works.
type ZerologOptions struct {
	// Logger is the zerolog.Logger instance to use.
	//
	// Default to zerolog.Ctx.
	Logger *zerolog.Logger

	// Level is the log level to use.
	//
	// By default, the log level is dynamic. If the Metrics instance indicates an error state with non-zero fault
	// and/or panicked counter, zerolog.ErrorLevel is used. Otherwise, zerolog.InfoLevel is used.
	Level *zerolog.Level

	// Msg is the message to be used with the log.
	Msg string

	// Dict is the name of the [zerolog.Event.Dict] to place the Metrics content.
	//
	// Equivalent to SlogLogger.Group.
	Dict string

	// NoCustomFormatter disables specialised formatting for timestamps and durations.
	//
	// By default, to match LogJSON, the start time is logged as epoch millisecond, the end time using RFC1123, and
	// all time.Duration using FormatDuration. To disable this behaviour and rely on slog's own formatters, set
	// NoCustomFormatter to true.
	NoCustomFormatter bool
}

func (opts *ZerologOptions) Log(ctx context.Context, m *Metrics) error {
	logger := zerolog.Ctx(ctx)

	var e *zerolog.Event
	if c := m.counters[CounterKeyFault]; c.isPositive() {
		e = logger.Error()
	} else if c := m.counters[CounterKeyPanicked]; c.isPositive() {
		e = logger.Error()
	} else {
		e = logger.Info()
	}

	if opts.Dict != "" {
		e = e.Dict(opts.Dict, m.e(zerolog.Dict(), opts.NoCustomFormatter))
	} else {
		e = m.e(e, opts.NoCustomFormatter)
	}

	e.Msg(opts.Msg)
	return nil
}
