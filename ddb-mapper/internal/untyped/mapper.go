package untyped

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/google/uuid"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	internaltypes "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/types"
)

// Mapper provides a type-agnostic way to interact with DynamoDB tables.
type Mapper struct {
	Options

	// StructType is the original struct type T passed to New[T].
	StructType reflect.Type
	// KeyStructType is StructType with only the key fields.
	KeyStructType reflect.Type

	TableName    string
	HashKey      *internaltypes.KeyAttribute
	SortKey      *internaltypes.KeyAttribute
	Version      *internaltypes.VersionAttribute
	CreatedTime  *internaltypes.TimeAttribute
	ModifiedTime *internaltypes.TimeAttribute
	Others       map[string]*internaltypes.Attribute
}

// updateVersion performs optimistic locking and updates the version attribute of the given item to its next value.
//
// Mapper.Version can be nil, in which case updateVersion no-ops and returns zero values.
//
// The item argument is assumed to have same type as Mapper.StructType, and fields indexed by the item are
// addressable ([reflect.Value.FieldByIndex] must return a value whose [reflect.Value.CanSet] is true). Panics
// otherwise.
//
// Ptr to item is needed if [untyped.Context.VersionUpdater] is availalble.
//
// Returns these values:
//   - undo: should be called to reset the version back to its original value on failure.
//   - cond: the condition based on the version; either "attribute_not_exists(#pk)" or "#version = :version".
func (m *Mapper) updateVersion(c *Context, item, ptr reflect.Value) (undo func(), cond expression.ConditionBuilder) {
	if m.Version == nil {
		return
	}

	v := item.FieldByIndex(m.Version.Index)
	isZero, prev := v.IsZero(), v.Interface()
	if isZero {
		cond = expression.NameNoDotSplit(m.HashKey.AttrName).AttributeNotExists()
	} else {
		cond = expression.NameNoDotSplit(m.Version.AttrName).Equal(expression.Value(prev))
	}

	if c.VersionUpdater != nil {
		c.VersionUpdater(ptr.Interface())
		undo = func() { v.Set(reflect.ValueOf(prev)) }
		return
	}

	switch {

	case v.CanInt():
		n := v.Int()
		v.SetInt(n + 1)
		undo = func() { v.SetInt(n) }

	case v.CanUint():
		n := v.Uint()
		v.SetUint(n + 1)
		undo = func() { v.SetUint(n) }

	case v.Kind() == reflect.String:
		s := v.String()
		v.SetString(uuid.NewString())
		undo = func() { v.SetString(s) }

	default:
		// can panic here because parse should have caught this case.
		panic(fmt.Errorf(`VersionUpdater must be given to support version field %s (%s) of struct type %s`, m.Version.StructField.Name, m.Version.StructField.Type, item.Type()))
	}

	return
}

// updateTimestamps performs timestamps modification to the given t value.
//
// Mapper.CreatedTime and Mapper.ModifiedTime can be nil, in which case the method no-ops and nil is returned.
// If updateCreatedTime is false, Mapper.CreatedTime will not be updated.
//
// The item argument is assumed to have same type as Mapper.StructType, and fields indexed by the item are
// addressable ([reflect.Value.FieldByIndex] must return a value whose [reflect.Value.CanSet] is true). Panics
// otherwise.
func (m *Mapper) updateTimestamps(item reflect.Value, t time.Time, updateCreatedTime bool) (undo func()) {
	var reset internal.ChainableFunc

	if updateCreatedTime && m.CreatedTime != nil {
		if v := item.FieldByIndex(m.CreatedTime.Index); v.IsZero() {
			prev := v.Interface()
			v.Set(reflect.ValueOf(t).Convert(v.Type()))
			reset = reset.And(func() { v.Set(reflect.ValueOf(prev)) })
		}
	}

	if m.ModifiedTime != nil {
		v := item.FieldByIndex(m.ModifiedTime.Index)
		prev := v.Interface()
		v.Set(reflect.ValueOf(t).Convert(v.Type()))
		reset = reset.And(func() { v.Set(reflect.ValueOf(prev)) })
	}

	return reset
}
