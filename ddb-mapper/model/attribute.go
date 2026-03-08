package model

import (
	"fmt"
	"reflect"

	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
)

// Attribute models a DynamoDB attribute parsed from a `dynamodbav`-tagged struct field.
type Attribute struct {
	reflect.StructField
	// StructType is the type of the struct that defines the StructField.
	StructType reflect.Type
	// AttrName is the name of the attribute.
	AttrName string
	// AttrType is the type of the attribute.
	//
	// Note that AttributeModelType is unrelated to an attribute's DynamoDB DataType ("S", "N", "B", etc.) nor its Go
	// type (string, int etc.) It's a concept that's meaningful only to [TableModel].
	AttrType AttributeModelType
}

// Get retrieves the value of the field in the given item.
//
// Because Get is read-only, item may be either a struct or struct pointer.
func (attr Attribute) Get(item any) (any, error) {
	v, err := attr.get(item, false)
	if err == nil {
		return v.Interface(), nil
	}

	return nil, err
}

// Set updates the value of the field in the given item.
//
// Unlike Get which is read-only, Set will modify the item argument; item must be a struct pointer as a result.
func (attr Attribute) Set(item, value any) error {
	v, err := attr.get(item, false)
	if err == nil {
		v.Set(reflect.ValueOf(value))
	}

	return err
}

// get returns the [reflect.Value] so that it can be used by both Get and Set.
func (attr Attribute) get(item any, mustBePointer bool) (v reflect.Value, err error) {
	v, _, err = internal.IndirectValueIsStruct(item, mustBePointer)
	if err == nil {
		v, err = v.FieldByIndexErr(attr.Index)
	}

	return
}

func (attr Attribute) String() string {
	return fmt.Sprintf("%sAttribute(%q, %s{%q: %s})", attr.AttrType, attr.AttrName, attr.StructType, attr.Name, attr.Type)
}

// AttributeModelType is the type assigned by [TableModel] to an Attribute.
//
// AttributeModelType is unrelated to an attribute's [DynamoDB data type] ("S", "N", "B", etc.) nor its Go type (string,
// int etc.). [TableModel] specifically cares about these types of attributes:
//   - Key attributes, which are either hash/partition key and sort/range keys. DynamoDB requires that these attributes
//     marshal to "S", N", or "B" data types but [TableModel] adds no validation.
//   - Version attribute for optimistic locking. Go string, int and uint types have out-of-the box support.
//   - Created and modified time attributes for auto-generating timestamps. The Go types of those attributes must be
//     assignable to [time.Time].
//
// [DynamoDB data type]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.NamingRulesDataTypes.html#HowItWorks.DataTypes
type AttributeModelType int

const (
	// AttributeModelTypeHashKey is a hash key attribute.
	AttributeModelTypeHashKey AttributeModelType = 1 << iota
	// AttributeModelTypeSortKey is a sort key attribute.
	AttributeModelTypeSortKey
	// AttributeModelTypeVersion is the version attribute used in optimistic locking.
	AttributeModelTypeVersion
	// AttributeModelTypeCreatedTime is an attribute for created time.
	AttributeModelTypeCreatedTime
	// AttributeModelTypeModifiedTime is an attribute for modification time.
	AttributeModelTypeModifiedTime
	// AttributeModelTypeOther is an attribute that has no special semantics to Mapper.
	AttributeModelTypeOther
)

func (a AttributeModelType) String() string {
	switch a {
	case AttributeModelTypeHashKey:
		return "HashKey"
	case AttributeModelTypeSortKey:
		return "SortKey"
	case AttributeModelTypeVersion:
		return "Version"
	case AttributeModelTypeCreatedTime:
		return "CreatedTime"
	case AttributeModelTypeModifiedTime:
		return "ModifiedTime"
	case AttributeModelTypeOther:
		return ""
	default:
		return ""
	}
}
