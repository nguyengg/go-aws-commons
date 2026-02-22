package slog

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
)

// JSONValue returns a slog.Value that will cause slog.JSONHandler to print the data as-is if it is valid JSON.
//
// If the logger isn't using slog.JSONHandler, or if data is not valid JSON, then the default logic for printing bytes
// is used, which should be printing its base64 standard encoding.
//
// Useful if you want to log a response from a GET call for example that you expect to be JSON, but you don't want to
// parse it into valid JSON first.
func JSONValue(data []byte) slog.Value {
	if json.Valid(data) {
		return slog.AnyValue(jsonValue{data})
	}

	return slog.AnyValue(data)
}

// JSONStringValue is a variant of JSONValue that receives a string instead.
//
// If the raw bytes of the string argument is not valid JSON, or if the logger isn't using slog.JSONHandler, the default
// logic for printing string will be used.
func JSONStringValue(s string) slog.Value {
	data := []byte(s)
	if json.Valid(data) {
		return slog.AnyValue(jsonStringValue{data})
	}

	return slog.StringValue(s)
}

// JSON is a slog.Any wrapper around JSONValue.
func JSON(key string, data []byte) slog.Attr {
	return slog.Any(key, JSONValue(data))
}

// JSONString is a slog.Any wrapper around JSONStringValue.
func JSONString(key, value string) slog.Attr {
	return slog.Any(key, JSONStringValue(value))
}

type jsonValue struct {
	data []byte
}

func (j jsonValue) String() string {
	return base64.StdEncoding.EncodeToString(j.data)
}

func (j jsonValue) MarshalJSON() ([]byte, error) {
	return j.data, nil
}

type jsonStringValue struct {
	data []byte
}

func (j jsonStringValue) String() string {
	return string(j.data)
}

func (j jsonStringValue) MarshalJSON() ([]byte, error) {
	return j.data, nil
}
