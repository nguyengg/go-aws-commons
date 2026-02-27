package slog

import (
	"encoding/json"
	"log/slog"

	"github.com/rotisserie/eris"
)

// ErrorValue returns a slog.Value for the given error.
//
// eris.Wrap will be used to add the stack trace to the error. Use Wrapf if you'd like to customise the wrap message.
func ErrorValue(err error) slog.Value {
	return slog.AnyValue(errorValue{err: eris.Wrap(err, err.Error())})
}

// AnError is a slog.Any wrapper around ErrorValue.
//
// eris.Wrap will be used to add the stack trace to the error. Use Wrapf if you'd like to customise the wrap message.
func AnError(key string, err error) slog.Attr {
	return slog.Any(key, ErrorValue(err))
}

// Wrapf returns a slog.Any wrapper for the given key and error that will be wrapped using eris.Wrapf.
func Wrapf(key string, err error, format string, a ...any) slog.Attr {
	return slog.Any(key, slog.AnyValue(errorValue{err: eris.Wrapf(err, format, a...)}))
}

type errorValue struct {
	err error
}

func (e errorValue) String() string {
	return eris.ToString(e.err, true)
}

func (e errorValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(eris.ToJSON(e.err, true))
}
