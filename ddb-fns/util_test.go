package ddbfns

import (
	"encoding/json"
	"time"
)

// sentinel test time.
var testTime, _ = time.Parse(time.DateTime, time.DateTime)

func MustToJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(data)
}
