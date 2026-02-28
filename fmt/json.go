// Package fmt provides [fmt.Formatter] implementations for printing/logging any data as JSON.
package fmt

import (
	"encoding/json"
	"fmt"
)

type jsonFormatter struct {
	v any
}

var _ fmt.Formatter = &jsonFormatter{}
var _ fmt.Formatter = (*jsonFormatter)(nil)

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
//	import . "github.com/nguyengg/go-aws-commons/fmt"
//
//	log.Printf("request=%s", JSON(v))
func JSON(v any) fmt.Formatter {
	return &jsonFormatter{v}
}

// JSONIdent is a variant of JSON that marshals with indentation.
//
//	import . "github.com/nguyengg/go-aws-commons/fmt"
//
//	log.Printf("request=%s", JSONIndent(v, "", "  "))
func JSONIdent(v any, prefix, indent string) fmt.Formatter {
	return &jsonIndentFormatter{v, prefix, indent}
}
