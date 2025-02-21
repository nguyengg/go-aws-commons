package ddb

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

// Builder parses and creates Table instances which then are used to create DynamoDB request input parameters.
//
// See Table for the `dynamodbav` tags that must be used. Builder is an abstraction on top of Table to create input
// parameters for DynamoDB DeleteItem (CreateDeleteItem), GetItem (CreateGetItem), PutItem (CreatePutItem), and
// UpdateItem (CreateUpdateItem).
//
// The zero-value Builder instance is ready for use.
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

// ParseFromStruct parses and caches the struct tags given by an instance of the struct.
//
// Returns an error if there are validation issues.
func (b *Builder) ParseFromStruct(in interface{}, optFns ...func(*BuildOptions)) (*Table, error) {
	return b.ParseFromType(reflect.TypeOf(in), optFns...)
}

// ParseFromType parses and caches the struct tags given by its type.
//
// Returns an error if there are validation issues.
func (b *Builder) ParseFromType(in reflect.Type, optFns ...func(*BuildOptions)) (table *Table, err error) {
	b.init.Do(b.initFn)

	opts := BuildOptions{}
	for _, fn := range optFns {
		fn(&opts)
	}

	table, err = newTable(in, func(_ *Attribute) (bool, error) {
		return true, nil
	}, b.Encoder)
	if err != nil {
		return nil, err
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

	b.cache.Store(in, table)
	return table, nil
}

func (b *Builder) loadOrParse(in reflect.Type) (*Table, error) {
	b.init.Do(b.initFn)

	in = DereferencedType(in)
	v, ok := b.cache.Load(in)
	if ok {
		return v.(*Table), nil
	}

	m, err := newTable(in, func(_ *Attribute) (bool, error) {
		return true, nil
	}, b.Encoder)
	if err != nil {
		return nil, err
	}

	b.cache.Store(in, m)
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
