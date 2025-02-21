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

// Put creates and executes a DynamoDB PutItem request from the given arguments.
//
// The argument item is a struct or pointer to struct containing all attributes to be written to DynamoDB.
//
// If the item's version has zero value, `attribute_not_exists(#hash_key)` is used as the condition expression to
// prevent overriding an existing item with the same key. The version attribute in the returned
// [dynamodb.PutItemInput.Item] will be incremented by 1.
//
// If the item's version is not zero value, `#version = :version` is used as the condition expression to add optimistic
// locking. The version attribute in the returned [dynamodb.PutItemInput.Item] will be incremented by 1.
//
// Any zero-value created or modified timestamps will be set to [time.Now].
//
// You can use [PutOptions.WithReturnAllOldValues] if you wish to return and unmarshal the old values to a struct, or
// [PutOptions.WithReturnAllOldValuesOnConditionCheckFailure] in case optimistic locking fails; they don't have to be
// pointers to the same struct.
//
// If you need to pass SDK-level options (in addition to options specified by NewManager), use [PutOptions.WithOptions].
//
// This method is a wrapper around [DefaultManager.Put].
func Put(ctx context.Context, item interface{}, optFns ...func(opts *PutOptions)) (*dynamodb.PutItemOutput, error) {
	return DefaultManager.Put(ctx, item, optFns...)
}

// Put creates and executes a DynamoDB PutItem request from the given arguments.
//
// The argument item is a struct or pointer to struct containing all attributes to be written to DynamoDB.
//
// If the item's version has zero value, `attribute_not_exists(#hash_key)` is used as the condition expression to
// prevent overriding an existing item with the same key. The version attribute in the returned
// [dynamodb.PutItemInput.Item] will be incremented by 1.
//
// If the item's version is not zero value, `#version = :version` is used as the condition expression to add optimistic
// locking. The version attribute in the returned [dynamodb.PutItemInput.Item] will be incremented by 1.
//
// Any zero-value created or modified timestamps will be set to [time.Now].
//
// You can use [PutOptions.WithReturnAllOldValues] if you wish to return and unmarshal the old values to a struct, or
// [PutOptions.WithReturnAllOldValuesOnConditionCheckFailure] in case optimistic locking fails; they don't have to be
// pointers to the same struct.
//
// If you need to pass SDK-level options (in addition to options specified by NewManager), use [PutOptions.WithOptions].
func (m *Manager) Put(ctx context.Context, item interface{}, optFns ...func(opts *PutOptions)) (*dynamodb.PutItemOutput, error) {
	var err error
	m.init.Do(func() {
		err = m.initFn(ctx)
	})
	if err != nil {
		return nil, err
	}

	opts := &PutOptions{
		EnableOptimisticLocking:       true,
		EnableAutoGeneratedTimestamps: true,
		optFns:                        m.ClientOptions,
	}
	input, err := m.Builder.createPutItem(item, opts, optFns...)
	if err != nil {
		return nil, err
	}

	putItemOutput, err := m.Client.PutItem(ctx, input, opts.optFns...)
	return putItemOutput, m.decode(err, opts.oldValues, opts.oldValuesOnConditionCheckFailure, func() map[string]types.AttributeValue {
		return putItemOutput.Attributes
	})
}

// CreatePutItem creates the PutItem input parameters for the given item.
//
// The argument item is a struct or pointer to struct containing all attributes to be written to DynamoDB.
//
// If the item's version has zero value, `attribute_not_exists(#hash_key)` is used as the condition expression to
// prevent overriding an existing item with the same key. The version attribute in the returned
// [dynamodb.PutItemInput.Item] will be incremented by 1.
//
// If the item's version is not zero value, `#version = :version` is used as the condition expression to add optimistic
// locking. The version attribute in the returned [dynamodb.PutItemInput.Item] will be incremented by 1.
//
// Any zero-value created or modified timestamps will be set to [time.Now].
//
// This method is a wrapper around [DefaultFns.CreatePutItem].
func CreatePutItem(item interface{}, optFns ...func(*PutOptions)) (*dynamodb.PutItemInput, error) {
	return DefaultBuilder.CreatePutItem(item, optFns...)
}

// CreatePutItem creates the PutItem input parameters for the given item.
//
// The argument item is a struct or pointer to struct containing all attributes to be written to DynamoDB.
//
// If the item's version has zero value, `attribute_not_exists(#hash_key)` is used as the condition expression to
// prevent overriding an existing item with the same key. The version attribute in the returned
// [dynamodb.PutItemInput.Item] will be incremented by 1.
//
// If the item's version is not zero value, `#version = :version` is used as the condition expression to add optimistic
// locking. The version attribute in the returned [dynamodb.PutItemInput.Item] will be incremented by 1.
//
// Any zero-value created or modified timestamps will be set to [time.Now].
func (b *Builder) CreatePutItem(item interface{}, optFns ...func(*PutOptions)) (*dynamodb.PutItemInput, error) {
	opts := &PutOptions{
		EnableOptimisticLocking:       true,
		EnableAutoGeneratedTimestamps: true,
	}
	return b.createPutItem(item, opts, optFns...)
}

// createPutItem requires an initial PutOptions.
func (b *Builder) createPutItem(item interface{}, opts *PutOptions, optFns ...func(*PutOptions)) (*dynamodb.PutItemInput, error) {
	b.init.Do(b.initFn)

	// apply optFns first in case we support modifying the parser.
	for _, fn := range optFns {
		fn(opts)
	}

	attrs, err := b.loadOrParse(reflect.TypeOf(item))
	if err != nil {
		return nil, err
	}

	if opts.TableName == nil {
		opts.TableName = aws.String(attrs.TableName)
	}

	// PutItem requires the entire map[string]AttributeValue itemAv.
	var itemAv map[string]types.AttributeValue
	if av, err := b.Encoder.Encode(item); err != nil {
		return nil, err
	} else if asMap, ok := av.(*types.AttributeValueMemberM); !ok {
		return nil, fmt.Errorf("itemAv did not encode to M type")
	} else {
		itemAv = asMap.Value
	}

	iv := reflect.Indirect(reflect.ValueOf(item))

	if versionAttr := attrs.Version; opts.EnableOptimisticLocking && versionAttr != nil {
		version, err := versionAttr.Get(iv)
		if err != nil {
			return nil, fmt.Errorf("get version value error: %w", err)
		}

		switch {
		case version.IsZero():
			opts.And(expression.Name(attrs.HashKey.AttributeName).AttributeNotExists())
			itemAv[versionAttr.AttributeName] = &types.AttributeValueMemberN{Value: "1"}
		case version.CanInt():
			opts.And(expression.Name(versionAttr.AttributeName).Equal(expression.Value(itemAv[versionAttr.AttributeName])))
			itemAv[versionAttr.AttributeName] = &types.AttributeValueMemberN{Value: strconv.FormatInt(version.Int()+1, 10)}
		case version.CanUint():
			opts.And(expression.Name(versionAttr.AttributeName).Equal(expression.Value(itemAv[versionAttr.AttributeName])))
			itemAv[versionAttr.AttributeName] = &types.AttributeValueMemberN{Value: strconv.FormatUint(version.Uint()+1, 10)}
		case version.CanFloat():
			opts.And(expression.Name(versionAttr.AttributeName).Equal(expression.Value(itemAv[versionAttr.AttributeName])))
			itemAv[versionAttr.AttributeName] = &types.AttributeValueMemberN{Value: strconv.FormatFloat(version.Float(), 'f', -1, 64)}
		default:
			panic(fmt.Errorf("version attribute's type (%s) is unknown numeric type", version.Type()))
		}
	}

	now := time.Now()

	if createdTimeAttr := attrs.CreatedTime; opts.EnableAutoGeneratedTimestamps && createdTimeAttr != nil {
		createdTime, err := createdTimeAttr.Get(iv)
		if err != nil {
			return nil, fmt.Errorf("get createdTime value error: %w", err)
		}

		var av types.AttributeValue
		if createdTime.IsZero() {
			if createdTimeAttr.UnixTime {
				if av, err = attributevalue.UnixTime(now).MarshalDynamoDBAttributeValue(); err != nil {
					return nil, fmt.Errorf("encode createdTime as UnixTime error: %w", err)
				}
			} else {
				updateValue := reflect.ValueOf(now).Convert(createdTime.Type())
				if av, err = b.Encoder.Encode(updateValue.Interface()); err != nil {
					return nil, fmt.Errorf("encode createdTime error: %w", err)
				}
			}

			itemAv[createdTimeAttr.AttributeName] = av
		}
	}

	if modifiedTimeAttr := attrs.ModifiedTime; opts.EnableAutoGeneratedTimestamps && modifiedTimeAttr != nil {
		modifiedTime, err := modifiedTimeAttr.Get(iv)
		if err != nil {
			return nil, fmt.Errorf("get modifiedTime value error: %w", err)
		}

		var av types.AttributeValue
		if modifiedTime.IsZero() {
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

			itemAv[modifiedTimeAttr.AttributeName] = av
		}
	}

	if opts.condition.IsSet() {
		expr, err := expression.NewBuilder().WithCondition(opts.condition).Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}

		return &dynamodb.PutItemInput{
			Item:                                itemAv,
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

	return &dynamodb.PutItemInput{
		Item:                                itemAv,
		TableName:                           opts.TableName,
		ReturnConsumedCapacity:              opts.ReturnConsumedCapacity,
		ReturnItemCollectionMetrics:         opts.ReturnItemCollectionMetrics,
		ReturnValues:                        opts.ReturnValues,
		ReturnValuesOnConditionCheckFailure: opts.ReturnValuesOnConditionCheckFailure,
	}, nil
}
