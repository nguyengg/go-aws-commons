package lambda

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// IsDebug is true if the "DEBUG" environment have value "1" or "true".
//
// The value of IsDebug is set at startup by way of init(). While many things in the lambda package use this value,
// nothing will modify it. If you want to use a different environment variable or a different way to toggle DEBUG
// behaviour, modify this value directly.
var IsDebug bool

func init() {
	switch os.Getenv("DEBUG") {
	case "1", "true":
		IsDebug = true
	}
}

const (
	// DebugLogFlags is the flag passed to log.SetFlags by SetUpLogger if IsDebug is true.
	DebugLogFlags = log.Ldate | log.Lmicroseconds | log.LUTC | log.Llongfile | log.Lmsgprefix

	// DefaultLogFlags is the flag passed to log.SetFlags by SetUpLogger if IsDebug is false.
	DefaultLogFlags = DebugLogFlags | log.Lshortfile
)

type jsonFormatter struct {
	v any
}

func (j *jsonFormatter) Format(f fmt.State, verb rune) {
	switch data, err := json.Marshal(j.v); err {
	case nil:
		_, _ = fmt.Fprintf(f, "%s", data)
	default:
		_, _ = fmt.Fprintf(f, string(verb), j.v)

	}
}

type jsonIndentFormatter struct {
	v              any
	prefix, indent string
}

func (j jsonIndentFormatter) Format(f fmt.State, verb rune) {
	switch data, err := json.MarshalIndent(j.v, j.prefix, j.indent); err {
	case nil:
		_, _ = fmt.Fprintf(f, "%s", data)
	default:
		_, _ = fmt.Fprintf(f, string(verb), j.v)
	}
}

// JSON returns a fmt.Formatter wrapper that returns the JSON representation of the given struct.
//
// If encoding the struct v fails, falls back to original formatter.
//
// Usage:
//
//	log.Printf("request=%s", JSON(v))
func JSON(v any) fmt.Formatter {
	return &jsonFormatter{v}
}

// JSONIdent is a variant of JSON that marshals with indentation.
func JSONIdent(v any, prefix, indent string) fmt.Formatter {
	return &jsonIndentFormatter{v, prefix, indent}
}
