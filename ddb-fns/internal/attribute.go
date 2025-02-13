package internal

import (
	"reflect"
)

// Attribute contains metadata about a reflect.StructField that represents a DynamoDB attribute.
type Attribute struct {
	// Field is the original struct field that was parsed.
	Field reflect.StructField
	// Name is the first tag value in the `dynamodbav` struct tag.
	Name string
	// OmitEmpty is true only if the `dynamodbav` struct tag also includes `omitempty`.
	OmitEmpty bool
	// UnixTime is true only if the `dynamodbav` struct tag also includes `unixtime`.
	UnixTime bool
}

// Get returns the reflected value from the given struct value.
func (a *Attribute) Get(value reflect.Value) (reflect.Value, error) {
	return value.FieldByIndexErr(a.Field.Index)
}
