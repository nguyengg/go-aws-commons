package model

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
)

type Options struct {
	// MustHave, if given, will fail creating the TableModel if the model doesn't have all the required attribute types.
	//
	// Usage:
	//	_, err := NewForType[Item](func(opts *Options) {
	//		opts.MustHave = AttributeModelTypeSortKey | AttributeModelTypeVersion | AttributeModelTypeCreatedTime | AttributeModelTypeModifiedTime
	//	})
	//	if err != nil {
	//		panic(err)
	//	}
	MustHave AttributeModelType
}

// NewForTypeOf creates a [TableModel] from the given i argument which should be a struct kind, or dereferences to a
// struct kind.
//
// Usage:
//
//	type Item struct {
//		ID string `dynamodbav:"id,hashkey" tablename:"Items"`
//	}
//
//	// either works.
//	NewForTypeOf(Item{})
//	NewForTypeOf(&Item{})
//
// See [TableModel] for details regarding how the struct tags are parsed.
func NewForTypeOf(i any, optFns ...func(opts *Options)) (*TableModel, error) {
	return New(reflect.TypeOf(i), optFns...)
}

// NewForType creates a [TableModel] from the given type T argument which should be a struct kind, or dereferences
// to a struct kind.
//
// Usage:
//
//	type Item struct {
//		ID string `dynamodbav:"id,hashkey" tablename:"Items"`
//	}
//
//	// either works.
//	NewForType[Item]
//	NewForType[*Item]
//
// See [TableModel] for details regarding how the struct tags are parsed.
func NewForType[T any](optFns ...func(opts *Options)) (*TableModel, error) {
	return New(reflect.TypeFor[T](), optFns...)
}

// New creates a [TableModel] from the given struct type argument which should be a struct kind, or dereferences
// to a struct kind.
//
// Usage:
//
//	type Item struct {
//		ID string `dynamodbav:"id,hashkey" tablename:"Items"`
//	}
//
//	// all will work.
//	New(reflect.TypeFor[Item]())
//	New(reflect.TypeOf(Item{}))
//	New(reflect.TypeFor[*Item]())
//	New(reflect.TypeOf(&Item{}))
//
// See [TableModel] for details regarding how the struct tags are parsed.
func New(in reflect.Type, optFns ...func(opts *Options)) (m *TableModel, err error) {
	opts := internal.ApplyOpts(&Options{}, optFns...)

	m = &TableModel{Others: map[string]Attribute{}}
	if m.StructType, err = internal.IndirectTypeIsStruct(in, false); err != nil {
		return nil, err
	}

	var (
		f         reflect.StructField
		tag, name string
		// dups is used to catch duplicate attribute names.
		dups = make(map[string]reflect.StructField)
		// errs is validation error that is reported at the end.
		errs = make([]error, 0)
	)

	// using in.Fields() requires go1.26.
	for i, n := 0, m.StructType.NumField(); i < n; i++ {
		if f = m.StructType.Field(i); !f.IsExported() {
			continue
		}

		if tag = f.Tag.Get("dynamodbav"); tag == "" {
			continue
		}

		tags := strings.Split(tag, ",")
		if name = tags[0]; name == "-" || name == "" {
			continue
		} else if dup, ok := dups[name]; ok {
			return nil, fmt.Errorf("struct type %s has multiple fields with same attribute name (%q): %q (%s), %q (%s)", m.StructType, name, dup.Name, dup.Type, f.Name, f.Type)
		}

		attr := Attribute{f, m.StructType, name, AttributeModelTypeOther}

		for _, tag = range tags[1:] {
			switch tag {
			case "hashkey", "pk", "primarykey":
				if m.HashKey != nil {
					return nil, fmt.Errorf("struct type %s has multiple hashkey attributes: %q (%s), %q (%s)", m.StructType, m.HashKey.Name, m.HashKey.Type, f.Name, f.Type)
				}

				m.HashKey, attr.AttrType = &attr, AttributeModelTypeHashKey

				if m.TableName = f.Tag.Get("tablename"); m.TableName == "" {
					if m.TableName = f.Tag.Get("tableName"); m.TableName == "" {
						return nil, fmt.Errorf("struct field %s{%q: %s} is missing required tablename for hashkey", m.StructType, f.Name, f.Type)
					}
				}

			case "sortkey", "sk", "rangekey":
				if m.SortKey != nil {
					return nil, fmt.Errorf("struct type %s has multiple sortkey attributes: %q (%s), %q (%s)", m.StructType, m.SortKey.Name, m.SortKey.Type, f.Name, f.Type)
				}

				m.SortKey, attr.AttrType = &attr, AttributeModelTypeSortKey

			case "version":
				if m.Version != nil {
					return nil, fmt.Errorf("struct type %s has multiple version attributes: %q (%s), %q (%s)", m.StructType, m.Version.Name, m.Version.Type, f.Name, f.Type)
				}

				m.Version, attr.AttrType = &attr, AttributeModelTypeVersion

			case "createdTime", "createdtime":
				if m.CreatedTime != nil {
					return nil, fmt.Errorf("struct type %s has multiple createdtime attributes: %q (%s), %q (%s)", m.StructType, m.CreatedTime.Name, m.CreatedTime.Type, f.Name, f.Type)
				} else if !f.Type.ConvertibleTo(timeType) {
					return nil, fmt.Errorf("struct field %s{%q: %s} is not assignable to time.Time", m.StructType, f.Name, f.Type)
				}

				m.CreatedTime, attr.AttrType = &attr, AttributeModelTypeCreatedTime

			case "modifiedTime", "modifiedtime":
				if m.ModifiedTime != nil {
					return nil, fmt.Errorf("struct type %s has multiple createdtime attributes: %q (%s), %q (%s)", m.StructType, m.ModifiedTime.Name, m.ModifiedTime.Type, f.Name, f.Type)
				} else if !f.Type.ConvertibleTo(timeType) {
					return nil, fmt.Errorf("struct field %s{%q: %s} is not assignable to time.Time", m.StructType, f.Name, f.Type)
				}

				m.ModifiedTime, attr.AttrType = &attr, AttributeModelTypeModifiedTime
			}
		}

		if attr.AttrType == AttributeModelTypeOther {
			m.Others[attr.Name] = attr
		}
	}

	// validation errors.
	if m.HashKey == nil {
		errs = append(errs, fmt.Errorf("no hashkey attribute found"))
	}
	if opts.MustHave != 0 {
		if opts.MustHave&AttributeModelTypeSortKey != 0 && m.SortKey == nil {
			errs = append(errs, errors.New("no sortkey attribute found"))
		}
		if opts.MustHave&AttributeModelTypeVersion != 0 && m.Version == nil {
			errs = append(errs, errors.New("no version attribute found"))
		}
		if opts.MustHave&AttributeModelTypeCreatedTime != 0 && m.CreatedTime == nil {
			errs = append(errs, errors.New("no createdtime attribute found"))
		}
		if opts.MustHave&AttributeModelTypeModifiedTime != 0 && m.ModifiedTime == nil {
			errs = append(errs, errors.New("no modifiedtime attribute found"))
		}
	}

	switch n := len(errs); n {
	case 0:
	case 1:
		return nil, fmt.Errorf("parse struct type %s error: %w", m.StructType, errs[0])
	default:
		return nil, fmt.Errorf("parse struct type %s error: %w", m.StructType, errors.Join(errs...))
	}

	// keyStructType is a dynamic struct created from just the keys.
	structFields := []reflect.StructField{m.HashKey.StructField}
	if m.SortKey != nil {
		structFields = append(structFields, m.SortKey.StructField)
	}
	m.KeyStructType = reflect.StructOf(structFields)

	return m, nil
}

var timeType = reflect.TypeFor[time.Time]()
