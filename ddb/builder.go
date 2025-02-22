package ddb

import (
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

func (b *Builder) loadOrParse(in reflect.Type) (*Table, error) {
	b.init.Do(b.initFn)

	in = DereferencedType(in)
	v, ok := b.cache.Load(in)
	if ok {
		return v.(*Table), nil
	}

	m, err := NewTable(in)
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
