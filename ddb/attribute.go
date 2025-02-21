package ddb

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	byteSliceType = reflect.TypeOf([]byte(nil))
	timeType      = reflect.TypeOf(time.Time{})
)

// Attribute is a reflect.StructField that is tagged with `dynamodbav` (or whatever TagKey passed to Parse).
type Attribute struct {
	reflect.StructField

	// AttributeName is the first tag value in the `dynamodbav` struct tag which is also the name of the attribute
	// in DynamoDB.
	AttributeName string
	// OmitEmpty is true only if the `dynamodbav` struct tag also includes `omitempty`.
	OmitEmpty bool
	// UnixTime is true only if the `dynamodbav` struct tag also includes `unixtime`.
	UnixTime bool

	encoder *attributevalue.Encoder
}

// Get is given a struct v (wrapped as a [reflect.Value]) that contains the Attribute.StructField in order to return the
// value (also wrapped as a [reflect.Value]) of that StructField in the struct v.
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
//	nameAv.Get(reflect.ValueOf(v)).String() // would return "John"
//	nameAv.Get(reflect.ValueOf(v)).SetString("Jane") // would change v.Name to "Jane"
func (s *Attribute) Get(v reflect.Value) (reflect.Value, error) {
	return v.FieldByIndexErr(s.Index)
}

// Encode encodes the given argument "in".
//
// TypeMismatchError is returned if value's type does not match the Attribute's type.
//
// Encode will honour OmitEmpty and UnixTime as much as it can.
func (s *Attribute) Encode(in interface{}) (types.AttributeValue, error) {
	switch v := reflect.ValueOf(in); {
	// for numeric types, the types are often interchangeable.
	case v.CanInt():
		switch s.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if s.OmitEmpty && v.IsZero() {
				return nil, nil
			}

			return &types.AttributeValueMemberN{Value: strconv.FormatInt(v.Int(), 10)}, nil
		default:
			return nil, &TypeMismatchError{
				Expected: s.Type,
				Actual:   v.Type(),
			}
		}

	case v.CanUint():
		switch s.Type.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if s.OmitEmpty && v.IsZero() {
				return nil, nil
			}

			return &types.AttributeValueMemberN{Value: strconv.FormatUint(v.Uint(), 10)}, nil
		default:
			return nil, &TypeMismatchError{
				Expected: s.Type,
				Actual:   v.Type(),
			}
		}

	case v.CanFloat():
		switch s.Type.Kind() {
		case reflect.Float32, reflect.Float64:
			if s.OmitEmpty && v.IsZero() {
				return nil, nil
			}

			return &types.AttributeValueMemberN{Value: strconv.FormatFloat(v.Float(), 'f', -1, 64)}, nil
		default:
			return nil, &TypeMismatchError{
				Expected: s.Type,
				Actual:   v.Type(),
			}
		}

	case s.Type != v.Type():
		return nil, &TypeMismatchError{
			Expected: s.Type,
			Actual:   v.Type(),
		}

	case s.OmitEmpty && v.IsZero():
		return nil, nil

	case v.Type().ConvertibleTo(timeType):
		var t time.Time
		t = v.Convert(timeType).Interface().(time.Time)
		if t.IsZero() && s.OmitEmpty {
			return nil, nil
		}
		fallthrough

	default:

		// at this point, there's no tag we can apply anymore so just fallback to the encoder.
		return s.encoder.Encode(in)
	}
}

// TypeMismatchError is returned by Attribute.Get if the type of the argument "in" does not match the Attribute's type.
type TypeMismatchError struct {
	Expected, Actual reflect.Type
}

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("mismatched type: expected %s, got %s", e.Expected, e.Actual)
}

// DereferencedType returns the innermost type that is not reflect.Interface or reflect.Ptr.
func DereferencedType(t reflect.Type) reflect.Type {
	for k := t.Kind(); k == reflect.Interface || k == reflect.Ptr; {
		t = t.Elem()
		k = t.Kind()
	}

	return t
}
