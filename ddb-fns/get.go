package ddbfns

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Get creates and executes a DynamoDB GetRequest request from the given arguments.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB GetItem only requires
// key attributes.
//
// The argument out is an optional pointer to struct to decode the returned attributes with
// [attributevalue.Decoder.Decode]. A common use case for Get is to make a GetItem request and unmarshall the DynamoDB
// response to a struct; pass a pointer to struct as the out argument to do so.
//
// If you need to pass SDK-level options (in addition to options specified by NewManager), use [GetOptions.WithOptions].
//
// This method is a wrapper around [DefaultManager.Get].
func Get(ctx context.Context, key interface{}, out interface{}, optFns ...func(opts *GetOptions)) (*dynamodb.GetItemOutput, error) {
	return DefaultManager.Get(ctx, key, out, optFns...)
}

// Get creates and executes a DynamoDB GetRequest request from the given arguments.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB GetItem only requires
// key attributes.
//
// The argument out is an optional pointer to struct to decode the returned attributes with
// [attributevalue.Decoder.Decode]. A common use case for Get is to make a GetItem request and unmarshall the DynamoDB
// response to a struct; pass a pointer to struct as the out argument to do so.
//
// If you need to pass SDK-level options (in addition to options specified by NewManager), use [GetOptions.WithOptions].
func (m *Manager) Get(ctx context.Context, key interface{}, out interface{}, optFns ...func(opts *GetOptions)) (*dynamodb.GetItemOutput, error) {
	var err error
	m.init.Do(func() {
		err = m.initFn(ctx)
	})
	if err != nil {
		return nil, err
	}

	opts := &GetOptions{optFns: m.ClientOptions}
	input, err := m.Builder.createGetItem(key, opts, optFns...)
	if err != nil {
		return nil, err
	}

	getItemOutput, err := m.Client.GetItem(ctx, input, opts.optFns...)
	return getItemOutput, m.decode(err, out, nil, func() map[string]types.AttributeValue {
		return getItemOutput.Item
	})
}

// CreateGetItem creates the GetItem input parameters for the given key.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB GetItem only requires
// key attributes.
//
// This method is a wrapper around [DefaultFns.CreateGetItem].
func CreateGetItem(key interface{}, optFns ...func(*GetOptions)) (*dynamodb.GetItemInput, error) {
	return DefaultBuilder.CreateGetItem(key, optFns...)
}

// CreateGetItem creates the GetItem input parameters for the given key.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB GetItem only requires
// key attributes.
func (b *Builder) CreateGetItem(key interface{}, optFns ...func(*GetOptions)) (*dynamodb.GetItemInput, error) {
	return b.createGetItem(key, &GetOptions{}, optFns...)
}

// createGetItem requires an initial GetOptions.
func (b *Builder) createGetItem(key interface{}, opts *GetOptions, optFns ...func(*GetOptions)) (*dynamodb.GetItemInput, error) {
	b.init.Do(b.initFn)

	// apply optFns first in case we support modifying the parser.
	for _, fn := range optFns {
		fn(opts)
	}

	attrs, err := b.loadOrParse(reflect.TypeOf(key))
	if err != nil {
		return nil, err
	}

	if opts.TableName == nil {
		opts.TableName = attrs.TableName
	}

	// GetItem only needs the key.
	var keyAv map[string]types.AttributeValue
	if av, err := b.Encoder.Encode(key); err != nil {
		return nil, err
	} else if asMap, ok := av.(*types.AttributeValueMemberM); !ok {
		return nil, fmt.Errorf("item did not encode to M type")
	} else {
		item := asMap.Value
		keyAv = map[string]types.AttributeValue{attrs.HashKey.Name: item[attrs.HashKey.Name]}
		if attrs.SortKey != nil {
			keyAv[attrs.SortKey.Name] = item[attrs.SortKey.Name]
		}
	}

	getItemInput := &dynamodb.GetItemInput{
		Key:                    keyAv,
		TableName:              opts.TableName,
		ConsistentRead:         opts.ConsistentRead,
		ReturnConsumedCapacity: opts.ReturnConsumedCapacity,
	}

	if names := opts.names; len(names) != 0 {
		projection := expression.NamesList(expression.Name(names[0]))
		for _, name := range names[1:] {
			projection = projection.AddNames(expression.Name(name))
		}

		expr, err := expression.NewBuilder().WithProjection(projection).Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}

		getItemInput.ExpressionAttributeNames = expr.Names()
		getItemInput.ProjectionExpression = expr.Projection()
	}

	return getItemInput, nil
}
