package ddb

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

// Builder parses the attributes from DynamoDB struct tags `dynamodbav` to build DynamoDB request input parameters.
//
// Specifically, Builder parses and understands these custom tag values:
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
// Having parsed the struct tags successfully, Builder can be used to create input parameters for DynamoDB DeleteItem
// (CreateDeleteItem), GetItem (CreateGetItem), PutItem (CreatePutItem), and UpdateItem (CreateUpdateItem).
//
// The zero-value Builder instance is ready for use. Prefer NewBuilder which can perform validation on the struct type.
type Builder struct {
	// Encoder is the attributevalue.Encoder to marshal structs into DynamoDB items.
	//
	// If nil, a default one will be created.
	Encoder *attributevalue.Encoder
	// Decoder is the attributevalue.Decoder to unmarshal results from DynamoDB.
	//
	// If nil, a default one will be created.
	Decoder *attributevalue.Decoder

	init  sync.Once
	cache sync.Map
}

// BuildOptions customises how Builder parses struct tags.
type BuildOptions struct {
	// MustHaveVersion, if true, will fail parsing if the struct does not have any field tagged as
	// `dynamodbav:",version"`.
	MustHaveVersion bool
	// MustHaveTimestamps, if true, will fail parsing if the struct does not have any field tagged as
	// `dynamodbav:",createdTime" or `dynamodbav:",modifiedTime".
	MustHaveTimestamps bool
}

// NewBuilder can be used to parse and validate the struct tags.
//
// This method should be called at least once (can be in the unit test) for every struct that will be used with Builder.
func NewBuilder[T any](optFns ...func(*BuildOptions)) (*Builder, error) {
	opts := &BuildOptions{}
	for _, fn := range optFns {
		fn(opts)
	}

	f := &Builder{}
	f.init.Do(f.initFn)

	if err := f.ParseFromType(reflect.TypeFor[T]()); err != nil {
		return nil, err
	}

	return f, nil
}

// ParseFromStruct parses and caches the struct tags given by an instance of the struct.
//
// Returns an error if there are validation issues.
func (b *Builder) ParseFromStruct(v interface{}, optFns ...func(*BuildOptions)) error {
	return b.ParseFromType(reflect.TypeOf(v), optFns...)
}

// ParseFromType parses and caches the struct tags given by its type.
//
// Returns an error if there are validation issues.
func (b *Builder) ParseFromType(t reflect.Type, optFns ...func(*BuildOptions)) error {
	opts := BuildOptions{}
	for _, fn := range optFns {
		fn(&opts)
	}

	m, err := ParseFromType(t)
	if err != nil {
		return err
	}

	if m.HashKey == nil {
		return fmt.Errorf(`no hashkey field in type "%s"`, t.Name())
	}
	if opts.MustHaveVersion && m.Version == nil {
		return fmt.Errorf(`no version field in type "%s"`, t.Name())
	}
	if opts.MustHaveTimestamps && m.CreatedTime == nil && m.ModifiedTime == nil {
		return fmt.Errorf(`no timestamp fields in type "%s"`, t.Name())
	}

	b.cache.Store(t, m)
	return nil
}

func (b *Builder) loadOrParse(t reflect.Type) (*Model, error) {
	t = DereferencedType(t)
	v, ok := b.cache.Load(t)
	if ok {
		return v.(*Model), nil
	}

	m, err := ParseFromType(t)
	if err != nil {
		return nil, err
	}

	b.cache.Store(t, m)
	return m, nil
}

func (b *Builder) initFn() {
	if b.Encoder == nil {
		b.Encoder = attributevalue.NewEncoder()
	}
	if b.Decoder == nil {
		b.Decoder = attributevalue.NewDecoder()
	}
}

// DefaultBuilder is the zero-value Builder instance used by CreateDeleteItem, CreateGetItem, CreatePutItem, and
// CreateUpdateItem.
var DefaultBuilder = &Builder{}
