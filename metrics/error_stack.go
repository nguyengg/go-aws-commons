package metrics

import (
	"encoding/json"
	"slices"

	"github.com/rotisserie/eris"
)

type wrappedError struct {
	err       error
	withTrace bool
}

func (w *wrappedError) toJSON() map[string]any {
	return eris.ToJSON(w.err, w.withTrace)
}

type errorStack []wrappedError

func (s *errorStack) push(err error, withTrace bool) {
	*s = slices.Insert(*s, 0, wrappedError{err: eris.Wrap(err, err.Error()), withTrace: withTrace})
}

func (s *errorStack) toJSON() []map[string]any {
	var data []map[string]any

	for err := range slices.Values(*s) {
		data = append(data, err.toJSON())
	}

	return data
}

func (s *errorStack) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.toJSON())
}
