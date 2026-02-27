package abort

import (
	"encoding/json"

	"github.com/rotisserie/eris"
)

type errorValue struct {
	err error
}

func (e errorValue) String() string {
	return eris.ToString(e.err, true)
}

func (e errorValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(eris.ToJSON(e.err, true))
}
