package ddb

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Delete creates and executes a DynamoDB DeleteItem request from the given arguments.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB DeleteItem only requires
// key attributes. The argument key should also include the version attribute that will be used to create the
// `#version = :version` condition expression if optimistic locking is enabled.
//
// A common use case for Delete is to delete and return old values of an item; pass
// [DeleteOptions.WithReturnAllOldValues] to do so. You can also use
// [DeleteOptions.WitReturnAllOldValuesOnConditionCheckFailure] if optimistic locking is enabled; they don't
// have to be pointers to the same struct.
//
// If you need to pass SDK-level options (in addition to options specified by NewManager), use
// [DeleteOptions.WithOptions].
//
// This method is a wrapper around [DefaultManager.Delete].
func Delete(ctx context.Context, key interface{}, optFns ...func(*DeleteOptions)) (*dynamodb.DeleteItemOutput, error) {
	return DefaultManager.Delete(ctx, key, optFns...)
}

// Delete creates and executes a DynamoDB DeleteItem request from the given arguments.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB DeleteItem only requires
// key attributes. The argument key should also include the version attribute that will be used to create the
// `#version = :version` condition expression if optimistic locking is enabled.
//
// A common use case for Delete is to delete and return old values of an item; pass
// [DeleteOptions.WithReturnAllOldValues] to do so. You can also use
// [DeleteOptions.WitReturnAllOldValuesOnConditionCheckFailure] if optimistic locking is enabled; they don't
// have to be pointers to the same struct.
//
// If you need to pass SDK-level options (in addition to options specified by NewManager), use
// [DeleteOptions.WithOptions].
//
// If you need to pass SDK-level options (in addition to options specified by NewManager), use [DeleteOptions.WithOptions].
func (m *Manager) Delete(ctx context.Context, key interface{}, optFns ...func(*DeleteOptions)) (*dynamodb.DeleteItemOutput, error) {
	var err error
	m.init.Do(func() {
		err = m.initFn(ctx)
	})
	if err != nil {
		return nil, err
	}

	opts := &DeleteOptions{
		EnableOptimisticLocking: true,
		optFns:                  m.ClientOptions,
	}
	input, err := m.Builder.createDeleteItem(key, opts, optFns...)
	if err != nil {
		return nil, err
	}

	deleteItemOutput, err := m.Client.DeleteItem(ctx, input, opts.optFns...)
	return deleteItemOutput, m.decode(err, opts.oldValues, opts.oldValuesOnConditionCheckFailure, func() map[string]types.AttributeValue {
		return deleteItemOutput.Attributes
	})
}

// CreateDeleteItem creates the DeleteItem input parameters for the given key.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB DeleteItem only requires
// key attributes. The argument key should also include the version field that will be used to create the
// `#version = :version` condition expression if optimistic locking is enabled.
//
// This method is a wrapper around [DefaultFns.CreateDeleteItem].
func CreateDeleteItem(key interface{}, optFns ...func(*DeleteOptions)) (*dynamodb.DeleteItemInput, error) {
	return DefaultBuilder.CreateDeleteItem(key, optFns...)
}

// CreateDeleteItem creates the DeleteItem input parameters for the given key.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB DeleteItem only requires
// key attributes. The argument key should also include the version field that will be used to create the
// `#version = :version` condition expression if optimistic locking is enabled.
func (b *Builder) CreateDeleteItem(key interface{}, optFns ...func(*DeleteOptions)) (*dynamodb.DeleteItemInput, error) {
	return b.createDeleteItem(key, &DeleteOptions{EnableOptimisticLocking: true}, optFns...)
}

// createDeleteItem requires an initial DeleteOptions.
func (b *Builder) createDeleteItem(key interface{}, opts *DeleteOptions, optFns ...func(*DeleteOptions)) (*dynamodb.DeleteItemInput, error) {
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

	// DeleteItem only needs the key.
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

	iv := reflect.Indirect(reflect.ValueOf(key))

	if versionAttr := attrs.Version; opts.EnableOptimisticLocking && versionAttr != nil {
		version, err := versionAttr.Get(iv)
		if err != nil {
			return nil, fmt.Errorf("get version value error: %w", err)
		}

		switch {
		case version.IsZero():
			opts.And(expression.Name(attrs.HashKey.Name).AttributeNotExists())
		case version.CanInt():
			opts.And(expression.Name(versionAttr.Name).Equal(expression.Value(&types.AttributeValueMemberN{Value: strconv.FormatInt(version.Int(), 10)})))
		case version.CanUint():
			opts.And(expression.Name(versionAttr.Name).Equal(expression.Value(&types.AttributeValueMemberN{Value: strconv.FormatUint(version.Uint(), 10)})))
		case version.CanFloat():
			opts.And(expression.Name(versionAttr.Name).Equal(expression.Value(&types.AttributeValueMemberN{Value: strconv.FormatFloat(version.Float(), 'f', -1, 64)})))
		default:
			panic(fmt.Errorf("version attribute's type (%s) is unknown numeric type", version.Type()))
		}
	}

	if opts.condition.IsSet() {
		expr, err := expression.NewBuilder().WithCondition(opts.condition).Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}

		return &dynamodb.DeleteItemInput{
			Key:                                 keyAv,
			TableName:                           opts.TableName,
			ConditionExpression:                 expr.Condition(),
			ExpressionAttributeNames:            expr.Names(),
			ExpressionAttributeValues:           expr.Values(),
			ReturnConsumedCapacity:              opts.ReturnConsumedCapacity,
			ReturnItemCollectionMetrics:         opts.ReturnItemCollectionMetrics,
			ReturnValues:                        opts.ReturnValues,
			ReturnValuesOnConditionCheckFailure: opts.ReturnValuesOnConditionCheckFailure,
		}, nil
	}

	return &dynamodb.DeleteItemInput{
		Key:                                 keyAv,
		TableName:                           opts.TableName,
		ReturnConsumedCapacity:              opts.ReturnConsumedCapacity,
		ReturnItemCollectionMetrics:         opts.ReturnItemCollectionMetrics,
		ReturnValues:                        opts.ReturnValues,
		ReturnValuesOnConditionCheckFailure: opts.ReturnValuesOnConditionCheckFailure,
	}, nil
}
