package ddb

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
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
// Other attributes tagged with `dynamodbav` can also be modeled by Table by passing a filter function to
// NewTableFromStructWithFilter or NewTableWithFilter. Note that the filter only affects attributes that can be returned
// with [Table.Get]. The five attributes described above are always stored by the Table.
type Table struct {
	TableName    string
	HashKey      *Attribute
	SortKey      *Attribute
	Version      *Attribute
	CreatedTime  *Attribute
	ModifiedTime *Attribute
	others       map[string]*Attribute
}

// NewTableFromStruct parses the struct tags given by an instance of the struct.
//
// Returns an error if there are validation issues. By default, all attributes are parsed and stored in the returned
// Table. To have control and validation over the attributes, use NewTableFromStructWithFilter instead.
//
// It is recommended to add a unit test and verify your struct is modeled correctly with NewTableFromStruct.
//
// Usage:
//
//	func TestMyStruct_IsValid(t *testing.T) {
//		_, err := NewTableFromStruct(MyStruct{})
//		assert.NoError(t, err)
//	}
func NewTableFromStruct(in interface{}) (*Table, error) {
	return NewTableWithFilter(reflect.TypeOf(in), func(attribute *Attribute) (bool, error) {
		return true, nil
	})
}

// NewTableFromStructWithFilter is a variant of NewTableFromStruct that allows caller to filter and validate additional
// fields.
//
// The filter function must return true in order for the field to be modeled in Table. If there is a validation error
// with the attribute (its type is not expected, etc.), the function may return an error which will stop the parsing
// process and return the error to caller. Hash and sort keys will always be stored.
func NewTableFromStructWithFilter(in interface{}, filter func(*Attribute) (bool, error)) (*Table, error) {
	return NewTableWithFilter(reflect.TypeOf(in), filter)
}

// NewTable parses the struct tags given by its type.
//
// Returns an error if there are validation issues. By default, all attributes are parsed and stored in the returned
// Table.
//
// It is recommended to add a unit test and verify your struct is modeled correctly with NewTableFromStruct.
//
// Usage:
//
//	func TestMyStruct_IsValid(t *testing.T) {
//		_, err := NewTable[MyStruct]()
//		assert.NoError(t, err)
//	}
func NewTable(t reflect.Type) (*Table, error) {
	return NewTableWithFilter(t, func(attribute *Attribute) (bool, error) {
		return true, nil
	})
}

// NewTableWithFilter is a variant of NewTable that allows caller to filter and validate additional
// fields.
//
// The filter function must return true in order for the field to be modeled in Table. If there is a validation error
// with the attribute (its type is not expected, etc.), the function may return an error which will stop the parsing
// process and return the error to caller. Hash and sort keys will always be stored.
func NewTableWithFilter(in reflect.Type, filter func(*Attribute) (bool, error)) (*Table, error) {
	return newTable(in, filter, attributevalue.NewEncoder())
}

// newTable requires an encoder to be passed.
func newTable(in reflect.Type, filter func(*Attribute) (bool, error), encoder *attributevalue.Encoder) (table *Table, err error) {
	in = DereferencedType(in)
	table = &Table{others: make(map[string]*Attribute)}

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

		if _, ok = table.others[name]; ok {
			return nil, fmt.Errorf(`found multiple attributes using name "%s" in type "%s"`, name, in.Name())
		}

		attr := &Attribute{StructField: structField, AttributeName: name, encoder: encoder}
		for _, tag = range tags[1:] {
			switch tag {
			case "hashkey":
				if table.HashKey != nil {
					return nil, fmt.Errorf(`found multiple hashkey fields in type "%s"`, in.Name())
				}

				if ok = validKeyAttribute(structField); !ok {
					return nil, fmt.Errorf(`unsupported hashkey field type "%s"`, structField.Type)
				}

				table.HashKey = attr
				if v, ok := structField.Tag.Lookup("tableName"); !ok {
					return nil, fmt.Errorf(`missing tableName tag on hashkey field`)
				} else if v != "" {
					table.TableName = v
				}
			case "sortkey":
				if table.SortKey != nil {
					return nil, fmt.Errorf(`found multiple sortkey fields in type "%s"`, in.Name())
				}

				if ok = validKeyAttribute(structField); !ok {
					return nil, fmt.Errorf(`unsupported sortkey field type "%s"`, structField.Type)
				}

				table.SortKey = attr
			case "version":
				if table.Version != nil {
					return nil, fmt.Errorf(`found multiple version fields in type "%s"`, in.Name())
				}

				if !validVersionAttribute(structField) {
					return nil, fmt.Errorf(`unsupported version field type "%s"`, structField.Type)
				}

				table.Version = attr
			case "createdTime":
				if table.CreatedTime != nil {
					return nil, fmt.Errorf(`found multiple createdTime fields in type "%s"`, in.Name())
				}

				if !validTimeAttribute(structField) {
					return nil, fmt.Errorf(`unsupported createdTime field type "%s"`, structField.Type)
				}

				table.CreatedTime = attr
			case "modifiedTime":
				if table.ModifiedTime != nil {
					return nil, fmt.Errorf(`found multiple modifiedTime fields in type "%s"`, in.Name())
				}

				if !validTimeAttribute(structField) {
					return nil, fmt.Errorf(`unsupported modifiedTime field type "%s"`, structField.Type)
				}

				table.ModifiedTime = attr
			case "omitempty":
				attr.OmitEmpty = true
			case "unixtime":
				attr.UnixTime = true
			}
		}

		if ok, err = filter(attr); err != nil {
			return nil, err
		} else if ok {
			table.others[attr.AttributeName] = attr
		}
	}

	return table, nil
}

// Get returns the attribute with the given `dynamodbav` struct tag name.
//
// This is not the name of the struct field (which is usually capital case to be exported).
func (t *Table) Get(name string) *Attribute {
	return t.others[name]
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
