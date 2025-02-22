package ddb

import "reflect"

// Attribute is a reflect.StructField that is tagged with `dynamodbav` (or whatever TagKey passed to Parse).
type Attribute struct {
	reflect.StructField

	// AttributeName is the first tag value in the `dynamodbav` struct tag which is also the name of the attribute
	// in DynamoDB.
	//
	// Note: AttributeName is often different from [Attribute.StructField.Name]. For example, given this struct:
	//
	//	type MyStruct struct {
	//		Field string `dynamodbav:"field"`
	//	}
	//
	// The Attribute for Field would have AttributeName="field" while Name="Field".
	AttributeName string
	// OmitEmpty is true only if the `dynamodbav` struct tag also includes `omitempty`.
	OmitEmpty bool
	// UnixTime is true only if the `dynamodbav` struct tag also includes `unixtime`.
	UnixTime bool
}

// GetFieldValue is given a struct v (wrapped as a [reflect.Value]) that contains the Attribute.StructField in order to
// return the value (also wrapped as a [reflect.Value]) of that StructField in the struct v.
//
// Usage example:
//
//	type MyStruct struct {
//		Name string `dynamodbav:"-"`
//	}
//
//	// nameAV is a Attribute modeling the Name field above, this can be used to get and/or set the value of that
//	// field like this:
//	v := MyStruct{Name: "John"}
//	nameAv.GetFieldValue(reflect.ValueOf(v)).String() // would return "John"
//	nameAv.GetFieldValue(reflect.ValueOf(v)).SetString("Jane") // would change v.Name to "Jane"
//
// You can use the returned [reflect.Value] to further set arbitrary values in the struct, though this will panic if
// the types aren't compatible.
func (s *Attribute) GetFieldValue(v reflect.Value) (reflect.Value, error) {
	return v.FieldByIndexErr(s.Index)
}
