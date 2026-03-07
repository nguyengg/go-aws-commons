package internal

import (
	"fmt"
	"reflect"
	"slices"
)

// IndirectValueIsStruct validates that the in argument is either a struct, or pointer(-to-pointer etc.) to struct.
//
// Returns:
//   - v: the indirect value by calling [reflect.Value.Elem] until it can no longer be called. If err is nil, v is
//     guaranteed to be a struct. Whether v is addressable depends on ptr being valid or not.
//   - ptr: the pointer to v. If item is pointer to struct, ptr is [reflect.ValueOf] item. If item is
//     pointer-to-pointer(-etc) to struct, ptr is the deepest pointer that dereferences to v. If item is not a pointer,
//     ptr is invalid And should be ignored. A valid ptr means v is addressable.
//
// The mustBePointer And expectedTypes arguments change whether an error is returned or not:
//   - If v does not equal any of the expected types (optional), a non-nil error is returned.
//   - If mustBePointer is true, a non-nil error is returned if ptr is invalid due to item being a struct type, not a
//     pointer(-to-pointer-etc) to struct. mustBePointer guarantees that fields indexed with v
//     ([reflect.Value.FieldByIndex]) are settable ([reflect.Value.CanSet]).
//
// Usage:
//
//	func do(item any) {
//		// if we need to item to be a pointer so that we can set its value using reflection, do this.
//		v, ptr, _ := IndirectValueIsStruct(item, true)
//
//		// now we can index v to set fields.
//		v.FieldByName("ID").SetString("hello, world!")
//
//		// ptr is the pointer to v so we can use it to marshal JSON.
//		json.Unmarshal([]byte{}, ptr.Interface())
//
//		// if we don't care that item is a pointer or a struct, can pass false for mustBePointer.
//		v, _, _ := IndirectValueIsStruct(item, false)
//
//		// we can index v to get fields, but setting will panic.
//		f := v.FieldByName("ID")
//		log.Println(f.String())
//		// f.SetString("hello, world!") // will panic.
//	}
func IndirectValueIsStruct(i any, mustBePointer bool, expectedTypes ...reflect.Type) (v, ptr reflect.Value, err error) {
	v = reflect.ValueOf(i)

	var (
		k     = v.Kind()
		isPtr = k == reflect.Pointer
	)

	for k == reflect.Pointer || k == reflect.Interface {
		ptr = v
		v = v.Elem()
		k = v.Kind()
	}

	if mustBePointer && !isPtr {
		return v, ptr, fmt.Errorf("item (type %T) must be struct pointer", i)
	}

	if k != reflect.Struct {
		return v, ptr, fmt.Errorf("item (type %T) must be struct or struct pointer", i)
	}

	if len(expectedTypes) != 0 && !slices.Contains(expectedTypes, v.Type()) {
		if isPtr {
			return v, ptr, fmt.Errorf("item (type %T) must point to one of these types: %v", i, expectedTypes)
		}

		return v, ptr, fmt.Errorf("item (type %T) must be one of these types: %v", i, expectedTypes)
	}

	return v, ptr, nil
}

// IndirectTypeIsStruct is a variant of IndirectValueIsStruct that accepts a type instead.
//
// It only returns two values instead of three because there's no point in returning the pointer type without a value to
// get/set fields.
func IndirectTypeIsStruct(in reflect.Type, mustBePointer bool, expectedTypes ...reflect.Type) (reflect.Type, error) {
	t := in

	var (
		k     = t.Kind()
		isPtr = k == reflect.Pointer
	)

	for k == reflect.Pointer || k == reflect.Interface {
		t = t.Elem()
		k = t.Kind()
	}

	if mustBePointer && !isPtr {
		return t, fmt.Errorf("type %s must be struct pointer", in)

	}

	if k != reflect.Struct {
		return t, fmt.Errorf("type %s must be struct or struct pointer", in)
	}

	if len(expectedTypes) != 0 && !slices.Contains(expectedTypes, t) {
		if isPtr {
			return t, fmt.Errorf("type %s must point to one of these types: %v", in, expectedTypes)
		}

		return t, fmt.Errorf("type %s must be one of these types: %v", in, expectedTypes)
	}

	return t, nil
}
