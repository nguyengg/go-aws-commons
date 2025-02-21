package ddb

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Update creates and executes a DynamoDB UpdateItem request from the given arguments.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB UpdateItem only requires
// key attributes. The argument key should also include the version field that will be used to create the condition
// expressions for optimistic locking.
//
// If the key's version has zero value, `attribute_not_exists(#hash_key)` is used as the condition expression to prevent
// overriding an existing item with the same key. An `ADD #version 1` update expression will be used to update the
// version.
//
// If the key's version is not zero value, `#version = :version` is used as the condition expression to add optimistic
// locking. An `ADD #version 1` update expression will be used to update the version.
//
// Modified time will always be set to [time.Now].
//
// This method is a wrapper around [DefaultManager.Update].
func Update(ctx context.Context, key interface{}, required func(*UpdateOptions), optFns ...func(*UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	return DefaultManager.Update(ctx, key, required, optFns...)
}

// Update creates and executes a DynamoDB UpdateItem request from the given arguments.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB UpdateItem only requires
// key attributes. The argument key should also include the version field that will be used to create the condition
// expressions for optimistic locking.
//
// If the key's version has zero value, `attribute_not_exists(#hash_key)` is used as the condition expression to prevent
// overriding an existing item with the same key. An `ADD #version 1` update expression will be used to update the
// version.
//
// If the key's version is not zero value, `#version = :version` is used as the condition expression to add optimistic
// locking. An `ADD #version 1` update expression will be used to update the version.
//
// Modified time will always be set to [time.Now].
func (m *Manager) Update(ctx context.Context, key interface{}, required func(*UpdateOptions), optFns ...func(*UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	var err error
	m.init.Do(func() {
		err = m.initFn(ctx)
	})
	if err != nil {
		return nil, err
	}

	opts := &UpdateOptions{
		EnableOptimisticLocking:       true,
		EnableAutoGeneratedTimestamps: true,
		optFns:                        m.ClientOptions,
	}
	required(opts)
	input, err := m.Builder.createUpdateItem(key, opts, optFns...)
	if err != nil {
		return nil, err
	}

	updateItemOutput, err := m.Client.UpdateItem(ctx, input, opts.optFns...)
	return updateItemOutput, m.decode(err, opts.values, opts.valuesOnConditionCheckFailure, func() map[string]types.AttributeValue {
		return updateItemOutput.Attributes
	})
}

// CreateUpdateItem creates the UpdateItem input parameters for the given key and at least one update expression.
//
// The argument key is a struct or pointer to struct containing the key values since DynamoDB UpdateItem only requires
// key attributes. The argument key should also include the version field that will be used to create the condition
// expressions for optimistic locking.
//
// If the key's version has zero value, `attribute_not_exists(#hash_key)` is used as the condition expression to prevent
// overriding an existing item with the same key. An `ADD #version 1` update expression will be used to update the
// version.
//
// If the key's version is not zero value, `#version = :version` is used as the condition expression to add optimistic
// locking. An `ADD #version 1` update expression will be used to update the version.
//
// Modified time will always be set to [time.Now].
func CreateUpdateItem(key interface{}, required func(*UpdateOptions), optFns ...func(opts *UpdateOptions)) (*dynamodb.UpdateItemInput, error) {
	return DefaultBuilder.CreateUpdateItem(key, required, optFns...)
}

// CreateUpdateItem creates the CreateDeleteItem input parameters for the given item.
//
// See package-level CreateUpdateItem for more information.
func (b *Builder) CreateUpdateItem(key interface{}, required func(*UpdateOptions), optFns ...func(*UpdateOptions)) (*dynamodb.UpdateItemInput, error) {
	opts := &UpdateOptions{
		EnableOptimisticLocking:       true,
		EnableAutoGeneratedTimestamps: true,
	}
	required(opts)
	return b.createUpdateItem(key, opts, optFns...)
}

// createUpdateItem requires an initial UpdateOptions.
func (b *Builder) createUpdateItem(key interface{}, opts *UpdateOptions, optFns ...func(*UpdateOptions)) (*dynamodb.UpdateItemInput, error) {
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
		opts.TableName = aws.String(attrs.TableName)
	}

	// UpdateItem only needs the key.
	var keyAv map[string]types.AttributeValue
	if av, err := b.Encoder.Encode(key); err != nil {
		return nil, err
	} else if asMap, ok := av.(*types.AttributeValueMemberM); !ok {
		return nil, fmt.Errorf("item did not encode to M type")
	} else {
		item := asMap.Value
		keyAv = map[string]types.AttributeValue{attrs.HashKey.AttributeName: item[attrs.HashKey.AttributeName]}
		if attrs.SortKey != nil {
			keyAv[attrs.SortKey.AttributeName] = item[attrs.SortKey.AttributeName]
		}
	}

	iv := reflect.Indirect(reflect.ValueOf(key))

	if versionAttr := attrs.Version; opts.EnableOptimisticLocking && versionAttr != nil {
		version, err := versionAttr.GetFieldValue(iv)
		if err != nil {
			return nil, fmt.Errorf("get version value error: %w", err)
		}

		switch {
		case version.IsZero():
			opts.And(expression.Name(attrs.HashKey.AttributeName).AttributeNotExists())
			opts.Set(versionAttr.AttributeName, &types.AttributeValueMemberN{Value: "1"})
		case version.CanInt():
			opts.And(expression.Name(versionAttr.AttributeName).Equal(expression.Value(&types.AttributeValueMemberN{Value: strconv.FormatInt(version.Int(), 10)})))
			opts.Add(versionAttr.AttributeName, 1)
		case version.CanUint():
			opts.And(expression.Name(versionAttr.AttributeName).Equal(expression.Value(&types.AttributeValueMemberN{Value: strconv.FormatUint(version.Uint(), 10)})))
			opts.Add(versionAttr.AttributeName, 1)
		case version.CanFloat():
			opts.And(expression.Name(versionAttr.AttributeName).Equal(expression.Value(&types.AttributeValueMemberN{Value: strconv.FormatFloat(version.Float(), 'f', -1, 64)})))
			opts.Add(versionAttr.AttributeName, 1)
		default:
			panic(fmt.Errorf("version attribute's type (%s) is unknown numeric type", version.Type()))
		}
	}

	now := time.Now()

	if modifiedTimeAttr := attrs.ModifiedTime; opts.EnableAutoGeneratedTimestamps && modifiedTimeAttr != nil {
		modifiedTime, err := modifiedTimeAttr.GetFieldValue(iv)
		if err != nil {
			return nil, fmt.Errorf("get modifiedTime value error: %w", err)
		}

		var av types.AttributeValue
		if modifiedTimeAttr.UnixTime {
			av, err = attributevalue.UnixTime(now).MarshalDynamoDBAttributeValue()
			if err != nil {
				return nil, fmt.Errorf("encode modifiedTime as UnixTime error: %w", err)
			}
		} else {
			updateValue := reflect.ValueOf(now).Convert(modifiedTime.Type())
			if av, err = b.Encoder.Encode(updateValue.Interface()); err != nil {
				return nil, fmt.Errorf("encode modifiedTime error: %w", err)
			}
		}

		opts.Set(modifiedTimeAttr.AttributeName, av)
	}

	var expr expression.Expression
	if opts.condition.IsSet() {
		expr, err = expression.NewBuilder().WithUpdate(opts.update).WithCondition(opts.condition).Build()
	} else {
		expr, err = expression.NewBuilder().WithUpdate(opts.update).Build()
	}
	if err != nil {
		return nil, fmt.Errorf("build expressions error: %w", err)
	}

	return &dynamodb.UpdateItemInput{
		Key:                                 keyAv,
		TableName:                           opts.TableName,
		ConditionExpression:                 expr.Condition(),
		ExpressionAttributeNames:            expr.Names(),
		ExpressionAttributeValues:           expr.Values(),
		ReturnConsumedCapacity:              opts.ReturnConsumedCapacity,
		ReturnItemCollectionMetrics:         opts.ReturnItemCollectionMetrics,
		ReturnValues:                        opts.ReturnValues,
		ReturnValuesOnConditionCheckFailure: opts.ReturnValuesOnConditionCheckFailure,
		UpdateExpression:                    expr.Update(),
	}, nil
}
