package untyped

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	internaltypes "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/types"
	mappertypes "github.com/nguyengg/go-aws-commons/ddb-mapper/types"
)

// NewFromItem creates a new Mapper from the given item.
func NewFromItem(item any, optFns ...func(opts *Options)) (*Mapper, error) {
	return NewFromType(reflect.TypeOf(item), optFns...)
}

// NewFromType creates a new Mapper from the given type.
func NewFromType(structType reflect.Type, optFns ...func(opts *Options)) (m *Mapper, err error) {
	m = &Mapper{
		Options: *internal.ApplyOpts[Options](&Options{}, optFns...),
		Others:  map[string]*internaltypes.Attribute{},
	}
	if m.StructType, err = internal.IndirectTypeIsStruct(structType, false); err != nil {
		return nil, err
	}

	// for detecting duplicates.
	names := make(map[string]internaltypes.Attribute)

	// using in.Fields() requires go1.26.
	for i, n := 0, m.StructType.NumField(); i < n; i++ {
		f := m.StructType.Field(i)
		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get("dynamodbav")
		if tag == "" {
			continue
		}

		tags := strings.Split(tag, ",")
		name := tags[0]
		if name == "-" || name == "" {
			continue
		}

		if dup, ok := names[name]; ok {
			return nil, fmt.Errorf("struct %s has two fields, %q (%s) and %q (%s), that both marshal to attribute name %q", m.StructType, dup.Name, dup.Type, f.Name, f.Type, name)
		}

		attr := internaltypes.Attribute{StructField: f, StructType: m.StructType, AttrName: name}
		names[name] = attr
		isSpecial := false

		for _, tag = range tags[1:] {
			switch tag {
			case "hashkey", "pk":
				if m.HashKey != nil {
					return nil, fmt.Errorf("struct %s has two fields, %q (%s) and %q (%s), that are both hashkey", m.StructType, m.HashKey.Name, m.HashKey.Type, f.Name, f.Type)
				}
				if m.HashKey, err = internaltypes.NewKeyAttribute(f, m.StructType, name, mappertypes.AttributeTypeHashKey); err != nil {
					return nil, err
				}

				// hashKey must also have tableName tag.
				var ok bool
				if m.TableName, ok = f.Tag.Lookup("tablename"); !ok {
					if m.TableName, ok = f.Tag.Lookup("tableName"); !ok {
						return nil, fmt.Errorf("struct %s is missing tablename tag on field %s (%s)", m.StructType, f.Name, f.Type)
					}
				}

				if m.TableName == "" {
					return nil, fmt.Errorf("struct %s has empty tablename tag on field %s (%s)", m.StructType, f.Name, f.Type)
				}

				isSpecial = true

			case "sortkey", "sk", "rangekey":
				if m.SortKey != nil {
					return nil, fmt.Errorf("struct %s has two fields, %q (%s) and %q (%s), that are both sortkey", m.StructType, m.SortKey.Name, m.SortKey.Type, f.Name, f.Type)
				}
				if m.SortKey, err = internaltypes.NewKeyAttribute(f, m.StructType, name, mappertypes.AttributeTypeSortKey); err != nil {
					return nil, err
				}

				isSpecial = true

			case "version":
				if m.Version != nil {
					return nil, fmt.Errorf("struct %s has two fields, %q (%s) and %q (%s), that are both version", m.StructType, m.Version.Name, m.Version.Type, f.Name, f.Type)
				}
				if m.Version, err = internaltypes.NewVersionAttribute(f, m.StructType, name, m.VersionUpdater != nil); err != nil {
					return nil, err
				}

				isSpecial = true

			case "createdTime":
				if m.CreatedTime != nil {
					return nil, fmt.Errorf("struct %s has two fields, %q (%s) and %q (%s), that are both createdTime", m.StructType, m.CreatedTime.Name, m.CreatedTime.Type, f.Name, f.Type)
				}
				if m.CreatedTime, err = internaltypes.NewTimeAttribute(f, m.StructType, name, mappertypes.AttributeTypeCreatedTime); err != nil {
					return nil, err
				}

				isSpecial = true

			case "modifiedTime":
				if m.ModifiedTime != nil {
					return nil, fmt.Errorf("struct %s has two fields, %q (%s) and %q (%s), that are both modifiedTime", m.StructType, m.ModifiedTime.Name, m.ModifiedTime.Type, f.Name, f.Type)
				}
				if m.ModifiedTime, err = internaltypes.NewTimeAttribute(f, m.StructType, name, mappertypes.AttributeTypeModifiedTime); err != nil {
					return nil, err
				}

				isSpecial = true
			}
		}

		if !isSpecial {
			m.Others[name] = &attr
		}
	}

	if m.HashKey == nil {
		return nil, fmt.Errorf("struct %s did not model hashkey", m.StructType)
	}

	// keyStructType is a dynamic struct created from just the keys.
	structFields := []reflect.StructField{m.HashKey.StructField}
	if m.SortKey != nil {
		structFields = append(structFields, m.SortKey.StructField)
	}
	m.KeyStructType = reflect.StructOf(structFields)

	return m, nil
}
