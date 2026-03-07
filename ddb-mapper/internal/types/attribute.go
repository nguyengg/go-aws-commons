// Package types defines various [mappertypes.Attribute] implementations.
package types

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	mappertypes "github.com/nguyengg/go-aws-commons/ddb-mapper/types"
)

// Attribute implements Attribute.
type Attribute struct {
	reflect.StructField
	// StructType is the type of the struct that defines the StructField.
	StructType reflect.Type
	// AttrName is the name of the attribute; implements [mappertypes.Attribute.AttributeName].
	AttrName string
	// AttrType is the type of the attribute.
	AttrType mappertypes.AttributeType
}

func (attr Attribute) String() string {
	return fmt.Sprintf("%sAttribute(%q, %s{%q: %s})", attr.AttrType, attr.AttrName, attr.StructType, attr.Name, attr.Type)
}

func (attr Attribute) AttributeName() string {
	return attr.AttrName
}

func (attr Attribute) AttributeType() mappertypes.AttributeType {
	return attr.AttrType
}

func (attr Attribute) Get(item any) (any, error) {
	v, err := attr.get(item, false)
	if err == nil {
		return v.Interface(), nil
	}

	return nil, err
}

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

// KeyAttribute extends Attribute with key metadata.
type KeyAttribute struct {
	Attribute
	types.ScalarAttributeType
	types.KeyType
}

// NewKeyAttribute creates a new key attribute with validation.
func NewKeyAttribute(f reflect.StructField, structType reflect.Type, attrName string, attrType mappertypes.AttributeType) (*KeyAttribute, error) {
	k := &KeyAttribute{
		Attribute: Attribute{
			StructField: f,
			StructType:  structType,
			AttrName:    attrName,
			AttrType:    attrType,
		},
	}

	switch attrType {
	case mappertypes.AttributeTypeHashKey:
		k.KeyType = types.KeyTypeHash
	case mappertypes.AttributeTypeSortKey:
		k.KeyType = types.KeyTypeRange
	default:
		panic(fmt.Errorf("invalid attribute type %q", attrType))
	}

	switch t := f.Type; t.Kind() {
	case reflect.String:
		k.ScalarAttributeType = types.ScalarAttributeTypeS
		return k, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		k.ScalarAttributeType = types.ScalarAttributeTypeN
		return k, nil

	case reflect.Array, reflect.Slice:
		if t == byteSliceType || t.Elem().Kind() == reflect.Uint8 {
			k.ScalarAttributeType = types.ScalarAttributeTypeB
			return k, nil
		}
		fallthrough

	default:
		return k, fmt.Errorf("struct %s field %q (%s) cannot be used as key Attribute", structType, f.Name, f.Type)
	}
}

func (k *KeyAttribute) String() string {
	return fmt.Sprintf("%sAttribute(%q, %s, %s{%q: %s})", k.AttrType, k.AttrName, k.ScalarAttributeType, k.StructType, k.Name, k.Type)
}

// VersionAttribute extends Attribute for optimistic locking functionality.
type VersionAttribute struct {
	Attribute
}

// NewVersionAttribute creates a new version attribute with validation.
func NewVersionAttribute(f reflect.StructField, structType reflect.Type, attrName string, hasVersionUpdater bool) (*VersionAttribute, error) {
	v := &VersionAttribute{
		Attribute: Attribute{
			StructField: f,
			StructType:  structType,
			AttrName:    attrName,
			AttrType:    mappertypes.AttributeTypeVersion,
		},
	}

	switch t := f.Type; t.Kind() {
	case reflect.String:
		return v, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v, nil

	default:
		if hasVersionUpdater {
			return v, nil
		}

		return nil, fmt.Errorf("struct %s field %q (%s) requires custom VersionUpdater", structType, f.Name, f.Type)
	}
}

// TimeAttribute extends Attribute for auto-generated timestamps functionality.
type TimeAttribute struct {
	Attribute
}

// NewTimeAttribute create a new timestamp attribute with validation.
func NewTimeAttribute(f reflect.StructField, structType reflect.Type, attrName string, attrType mappertypes.AttributeType) (*TimeAttribute, error) {
	if f.Type.ConvertibleTo(timeType) {
		return &TimeAttribute{Attribute{
			StructField: f,
			StructType:  structType,
			AttrName:    attrName,
			AttrType:    attrType,
		}}, nil
	}

	return nil, fmt.Errorf("struct %s field %q (%s) cannot be used as timestamp Attribute", structType, f.Name, f.Type)
}

var (
	byteSliceType = reflect.TypeFor[[]byte]()
	timeType      = reflect.TypeFor[time.Time]()
)
