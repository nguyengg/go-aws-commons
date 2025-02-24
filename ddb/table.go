package ddb

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	// ErrNoAttribute is returned by Table.Get if there is no attribute with the requested name.
	ErrNoAttribute = errors.New("ddb: named attribute not present")

	byteSliceType = reflect.TypeOf([]byte(nil))
	timeType      = reflect.TypeOf(time.Time{})
)

// Table is a collection of Attribute where the hash key must always be present.
//
// `dynamodav` struct tag must be used to model these attributes:
//
//	// Hash key is required, sort key is optional. If present, their types must marshal to a valid key type
//	// (S, N, or B). The keys are required to make `attribute_not_exists` work on creating the condition expression
//	// for the PutItem request of an item that shouldn't exist in database.
//	Field string `dynamodbav:"-,hashkey" tableName:"my-table"`
//	Field string `dynamodbav:"-,sortkey"`
//
//	// Versioned attribute must have `version` show up in its `dynamodbav` tag. It must be a numeric type that
//	// marshals to type N in DynamoDB.
//	Field int64 `dynamodbav:"-,version"`
//
//	// Timestamp attributes must have `createdTime` and/or `modifiedTime` in its `dynamodbav` tag. It must be a
//	// [time.Time] value. In this example, both attributes marshal to type N in DynamoDB as epoch millisecond.
//	Field time.Time `dynamodbav:"-,createdTime,unixtime"`
//	Field time.Time `dynamodbav:"-,modifiedTime,unixtime"`
//
// All other attributes tagged with `dynamodbav` are also stored. Attributed with no names such as `dynamodbav:"-"` are
// ignored.
type Table struct {
	TableName    string
	HashKey      *Attribute
	SortKey      *Attribute
	Version      *Attribute
	CreatedTime  *Attribute
	ModifiedTime *Attribute

	// All contains all attributes parsed from `dynamodbav` struct tags, including the special ones such has HashKey
	// and SortKey.
	//
	// The map's key is the Attribute.AttributeName. Use Table.Get for a way to retrieve the value of an attribute
	// given its struct.
	All map[string]*Attribute

	// inType is the original struct type that was used to create the table.
	inType reflect.Type
}

// TableOptions customises the behaviour of the various methods to create parse struct tags for a Table.
type TableOptions struct {
	// Validator can be used to fail parsing early.
	//
	// Any non-nil error will stop the parsing process and is returned immediately to caller.
	Validator func(*Attribute) error
	// MustHaveSortKey, if true, will fail parsing if the struct does not have any field tagged as
	// `dynamodbav:",sortkey"`.
	MustHaveSortKey bool
	// MustHaveVersion, if true, will fail parsing if the struct does not have any field tagged as
	// `dynamodbav:",version"`.
	MustHaveVersion bool
	// MustHaveTimestamps, if true, will fail parsing if the struct does not have any field tagged as
	// `dynamodbav:",createdTime" or `dynamodbav:",modifiedTime".
	MustHaveTimestamps bool
}

// NewTableFromStruct parses the struct tags given by an instance of the struct.
//
// Returns an InvalidModelErr error if there are validation issues.
//
// It is recommended to add a unit test and verify your struct is modeled correctly with NewTableFromStruct.
//
// Usage:
//
//	func TestMyStruct_IsValid(t *testing.T) {
//		_, err := NewTableFromStruct(MyStruct{})
//		assert.NoError(t, err)
//	}
func NewTableFromStruct(in interface{}, optFns ...func(*TableOptions)) (*Table, error) {
	return NewTable(reflect.TypeOf(in), optFns...)
}

// NewTable parses the struct tags given by its type.
//
// Returns an InvalidModelErr error if there are validation issues.
//
// It is recommended to add a unit test and verify your struct is modeled correctly with NewTableFromStruct.
//
// Usage:
//
//	func TestMyStruct_IsValid(t *testing.T) {
//		_, err := NewTable[MyStruct]()
//		assert.NoError(t, err)
//	}
func NewTable(in reflect.Type, optFns ...func(*TableOptions)) (table *Table, err error) {
	in = DereferencedType(in)

	opts := &TableOptions{}
	for _, fn := range optFns {
		fn(opts)
	}

	table = &Table{All: make(map[string]*Attribute), inType: in}

	var ok bool

	for i, n := 0, in.NumField(); i < n; i++ {
		structField := in.Field(i)
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

		if _, ok = table.All[name]; ok {
			return nil, InvalidModelErr{in, fmt.Errorf(`found multiple attributes using name "%s"`, name)}
		}

		attr := &Attribute{StructField: structField, AttributeName: name}
		table.All[name] = attr

		for _, tag = range tags[1:] {
			switch tag {
			case "hashkey":
				if table.HashKey != nil {
					return nil, InvalidModelErr{in, fmt.Errorf("found multiple hashkey fields")}
				}

				if ok = validKeyAttribute(structField); !ok {
					return nil, InvalidModelErr{in, fmt.Errorf(`unsupported hashkey field type "%s"`, structField.Type)}
				}

				table.HashKey = attr
				if v, ok := structField.Tag.Lookup("tableName"); !ok {
					return nil, InvalidModelErr{in, fmt.Errorf(`missing tableName tag on hashkey field`)}
				} else if v != "" {
					table.TableName = v
				}
			case "sortkey":
				if table.SortKey != nil {
					return nil, InvalidModelErr{in, fmt.Errorf("found multiple sortkey fields")}
				}

				if ok = validKeyAttribute(structField); !ok {
					return nil, InvalidModelErr{in, fmt.Errorf(`unsupported sortkey field type "%s"`, structField.Type)}
				}

				table.SortKey = attr
			case "version":
				if table.Version != nil {
					return nil, InvalidModelErr{in, fmt.Errorf("found multiple version fields")}
				}

				if !validVersionAttribute(structField) {
					return nil, InvalidModelErr{in, fmt.Errorf(`unsupported version field type "%s"`, structField.Type)}
				}

				table.Version = attr
			case "createdTime":
				if table.CreatedTime != nil {
					return nil, InvalidModelErr{in, fmt.Errorf("found multiple createdTime fields")}
				}

				if !validTimeAttribute(structField) {
					return nil, InvalidModelErr{in, fmt.Errorf(`unsupported createdTime field type "%s"`, structField.Type)}
				}

				table.CreatedTime = attr
			case "modifiedTime":
				if table.ModifiedTime != nil {
					return nil, InvalidModelErr{in, fmt.Errorf("found multiple modifiedTime fields")}
				}

				if !validTimeAttribute(structField) {
					return nil, InvalidModelErr{in, fmt.Errorf(`unsupported modifiedTime field type "%s"`, structField.Type)}
				}

				table.ModifiedTime = attr
			case "omitempty":
				attr.OmitEmpty = true
			case "unixtime":
				attr.UnixTime = true
			}
		}

		if opts.Validator != nil {
			if err = opts.Validator(attr); err != nil {
				return nil, InvalidModelErr{In: in, Cause: fmt.Errorf("validate error: %w", err)}
			}
		}
	}

	if table.HashKey == nil {
		return nil, fmt.Errorf(`no hashkey field in type "%s"`, in.Name())
	}
	if opts.MustHaveVersion && table.Version == nil {
		return nil, fmt.Errorf(`no version field in type "%s"`, in.Name())
	}
	if opts.MustHaveTimestamps && table.CreatedTime == nil && table.ModifiedTime == nil {
		return nil, fmt.Errorf(`no timestamp fields in type "%s"`, in.Name())
	}

	return table, nil
}

// Get returns the value of the attribute with the given name in the given struct v.
//
// Returns TypeMismatchErr if in's type is not the same as the struct type that was used to create the Table. Return
// ErrNoAttribute if Table.All contains no attribute with the given name.
//
// Usage example:
//
//	type MyStruct struct {
//		Name string `dynamodbav:"name"`
//	}
//
//	t, _ := NewTableFromStruct(&MyStruct{})
//	name, _ := t.Get(&MyStruct{Name: "hello"}, "name") // notice the lowercase "name" matching the dynamodbav tag.
//	fmt.Printf("%s", name.(string)); // # hello
func (t *Table) Get(in interface{}, name string) (_ interface{}, err error) {
	if inType := DereferencedType(reflect.TypeOf(in)); inType != t.inType {
		return nil, TypeMismatchErr{Expected: t.inType, Actual: t.inType}
	}

	a, ok := t.All[name]
	if !ok {
		return nil, ErrNoAttribute
	}

	v, err := a.GetFieldValue(reflect.Indirect(reflect.ValueOf(in)))
	if err != nil {
		return nil, err
	}

	return v.Interface(), nil
}

// MustGet is a variant of Get that panics instead of returning any error.
//
// Usage example:
//
//	type MyStruct struct {
//		Name string `dynamodbav:"-"`
//	}
//
//	t, _ := NewTableFromStruct(&MyStruct{})
//	var name string = t.MustGet(MyStruct{Name: "hello"}).(string)
func (t *Table) MustGet(in interface{}, name string) interface{} {
	v, err := t.Get(in, name)
	if err != nil {
		panic(err)
	}

	return v
}

// InvalidModelErr is returned by NewTable or NewTableFromStruct if the struct type is invalid.
type InvalidModelErr struct {
	In    reflect.Type
	Cause error
}

func (e InvalidModelErr) Error() string {
	return fmt.Sprintf("ddb: parse struct type %s error: %s", e.In, e.Cause)
}

// TypeMismatchErr is returned by Table.Get if the type of the argument "in" does not match the Table's type.
type TypeMismatchErr struct {
	Expected, Actual reflect.Type
}

func (e TypeMismatchErr) Error() string {
	return fmt.Sprintf("ddb: mismatched type: expected %s, got %s", e.Expected, e.Actual)
}

// DereferencedType returns the innermost type that is not reflect.Interface or reflect.Ptr.
func DereferencedType(t reflect.Type) reflect.Type {
	for k := t.Kind(); k == reflect.Interface || k == reflect.Ptr; {
		t = t.Elem()
		k = t.Kind()
	}

	return t
}

func validKeyAttribute(field reflect.StructField) bool {
	switch ft := field.Type; ft.Kind() {
	case reflect.String:
		return true
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
