package metrics

import (
	"log/slog"
)

// ReplaceAttr creates a [slog.HandlerOptions.ReplaceAttr] that works best for me.
func ReplaceAttr() func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		switch a.Value.Kind() {

		case slog.KindString:
			// drop empty strings.
			if a.Value.String() == "" {
				return slog.Attr{}
			}

		case slog.KindAny:
			// drop nil pointers/interface values.
			if a.Value.Any() == nil {
				return slog.Attr{}
			}

		default:
			return a
		}

		return a
	}
}
