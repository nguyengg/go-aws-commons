package client

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/google/uuid"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// updateVersion performs optimistic locking and updates the version attribute of the given item to its next value.
//
// The ptr argument is pointer to the item; needed if the versionUpdater argument is non-nil.
//
// Returns these values:
//   - undo: should be called to reset the version back to its original value on failure.
//   - cond: the condition based on the version; either "attribute_not_exists(#pk)" or "#version = :version".
func updateVersion(m *model.TableModel, item, ptr reflect.Value, versionUpdater func(any)) (undo func(), cond expression.ConditionBuilder, err error) {
	v := item.FieldByIndex(m.Version.Index)
	isZero, prev := v.IsZero(), v.Interface()
	if isZero {
		cond = expression.NameNoDotSplit(m.HashKey.AttrName).AttributeNotExists()
	} else {
		cond = expression.NameNoDotSplit(m.Version.AttrName).Equal(expression.Value(prev))
	}

	if versionUpdater != nil {
		versionUpdater(ptr.Interface())
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
		return nil, cond, fmt.Errorf(`a custom VersionUpdater is required for field %s{%q: %s}`, m.StructType, m.Version.Name, m.Version.Type)
	}

	return
}

// updateTimestamps performs timestamps modification to the given t value.
//
// The given [model.TableModel.CreatedTime] and [model.TableModel.ModifiedTime] can be nil, in which case the method
// no-ops and nil is returned. If updateCreatedTime is false, [model.TableModel.CreatedTime] will not be updated.
func updateTimestamps(m *model.TableModel, item reflect.Value, t time.Time, updateCreatedTime bool) (undo func()) {
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
