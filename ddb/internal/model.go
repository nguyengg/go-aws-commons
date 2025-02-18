package internal

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	byteSliceType = reflect.TypeOf([]byte(nil))
	timeType      = reflect.TypeOf(time.Time{})
)

// Model contain metadata about attributes that have been parsed successfully from struct tags `dynamodbav`.
type Model struct {
	// StructType is the type of the struct from which the Model instance was parsed.
	StructType   reflect.Type
	TableName    *string
	HashKey      *Attribute
	SortKey      *Attribute
	Version      *Attribute
	CreatedTime  *Attribute
	ModifiedTime *Attribute
}

// DereferencedType returns the innermost type that is not reflect.Interface or reflect.Ptr.
func DereferencedType(t reflect.Type) reflect.Type {
	for k := t.Kind(); k == reflect.Interface || k == reflect.Ptr; {
		t = t.Elem()
		k = t.Kind()
	}

	return t
}

// ParseFromStruct parses the struct tags given by an instance of the struct.
//
// Returns an error if there are validation issues.
func ParseFromStruct(v interface{}) (*Model, error) {
	return ParseFromType(reflect.TypeOf(v))
}

// ParseFromType parses the struct tags given by its type.
//
// Returns an error if there are validation issues.
func ParseFromType(t reflect.Type) (*Model, error) {
	t = DereferencedType(t)
	m := &Model{StructType: t}

	for i, n := 0, t.NumField(); i < n; i++ {
		structField := t.Field(i)
		if !structField.IsExported() {
			continue
		}

		tag := structField.Tag.Get("dynamodbav")
		if tag == "" {
			continue
		}

		tags := strings.Split(tag, ",")
		name := tags[0]
		if name == "-" || name == "" {
			continue
		}

		attr := &Attribute{Name: name, Field: structField}
		for _, tag = range tags[1:] {
			switch tag {
			case "hashkey":
				if m.HashKey != nil {
					return nil, fmt.Errorf(`found multiple hashkey fields in type "%s"`, t.Name())
				}

				if !validKeyAttribute(structField) {
					return nil, fmt.Errorf(`unsupported hashkey field type "%s"`, structField.Type)
				}

				m.HashKey = attr
				if v, ok := structField.Tag.Lookup("tableName"); !ok {
					return nil, fmt.Errorf(`missing tableName tag on hashkey field`)
				} else if v != "" {
					m.TableName = &v
				}
			case "sortkey":
				if m.SortKey != nil {
					return nil, fmt.Errorf(`found multiple sortkey fields in type "%s"`, t.Name())
				}

				if !validKeyAttribute(structField) {
					return nil, fmt.Errorf(`unsupported sortkey field type "%s"`, structField.Type)
				}

				m.SortKey = attr
			case "version":
				if m.Version != nil {
					return nil, fmt.Errorf(`found multiple version fields in type "%s"`, t.Name())
				}

				if !validVersionAttribute(structField) {
					return nil, fmt.Errorf(`unsupported version field type "%s"`, structField.Type)
				}

				m.Version = attr
			case "createdTime":
				if m.CreatedTime != nil {
					return nil, fmt.Errorf(`found multiple createdTime fields in type "%s"`, t.Name())
				}

				if !validTimeAttribute(structField) {
					return nil, fmt.Errorf(`unsupported createdTime field type "%s"`, structField.Type)
				}

				m.CreatedTime = attr
			case "modifiedTime":
				if m.ModifiedTime != nil {
					return nil, fmt.Errorf(`found multiple modifiedTime fields in type "%s"`, t.Name())
				}

				if !validTimeAttribute(structField) {
					return nil, fmt.Errorf(`unsupported modifiedTime field type "%s"`, structField.Type)
				}

				m.ModifiedTime = attr
			case "unixtime":
				attr.UnixTime = true
			}
		}
	}

	return m, nil
}

func validKeyAttribute(field reflect.StructField) bool {
	switch ft := field.Type; ft.Kind() {
	case reflect.String:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Array, reflect.Slice:
		return ft == byteSliceType || ft.Elem().Kind() == reflect.Uint8
	default:
		return false
	}
}

func validVersionAttribute(field reflect.StructField) bool {
	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func validTimeAttribute(field reflect.StructField) bool {
	return field.Type.ConvertibleTo(timeType)
}
