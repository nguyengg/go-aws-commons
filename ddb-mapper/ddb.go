// Package ddb provides a type-agnostic way to interact with items in DynamoDB tables via package-level methods such as
// [Get], [Put], [Update], etc.
//
// All package-level methods use the client returned by [DefaultClientProvider] unless explicitly overridden at the
// call level via [config.Config.Client] embedded in the options.
//
// # Struct tag modeling
//
// The package-level methods expect the item argument to be parsable with [model.NewForTypeOf], which uses the
// `dynamodbav` struct tags to mark the struct field as DynamoDB attributes.
//
//	type Item struct {
//		// "hashkey" attribute must also provide the table name.
//		ID	string	`dynamodbav:"id,hashkey|partitionkey|pk" tablename:"Items"`
//
//		// "sortkey" attribute is optional.
//		Shard	int	`dynamodbav:"shard,sortkey|sk|rangekey"`
//
//		// "version" attribute is used in optimistic locking computation.
//		Version	int	`dynamodbav:"version,version"`
//
//		// "createdtime" and "modifiedtime" time.Time fields are used in auto-generated timestamps feature.
//		Created		time.Time	`dynamodbav:"created,createdtime,unixtime"`
//		Modified	time.Time	`dynamodbav:"modified,modifiedtime,unixtime"`
//	}
//
// # Optimistic locking
//
// [Delete], [Put], and [Update] support optimistic locking. The item argument passed into those methods should have
// the expected value in its version attribute. If value is the zero-value, [Put] and [Update] will add an
// `attribute_not_exists(#pk)` to prevent unintentional overwriting of an existing item. Otherwise, [Delete], [Put], and
// [Update] will add a `#version = :version` condition. [Put] and [Update] will then modify the item's in-place version
// field to its next value depending on the Go type of the field:
//   - int and uint types simply increase the value by 1.
//   - string types are set to [uuid.NewString].
//   - All other types must explicitly set [config.Config.VersionUpdater].
//
// The encoded `map[string]AttributeValue` of the item is eventually sent to DynamoDB. On failure, the version field is
// restored to its value prior to the method call. On success, you can reasonably assume that the version in item
// matches what is in DynamoDB.
//
// # Auto-generated timestamps
//
// [Put] and [Update] support auto-updating created time and last-modified time. Similar to optimistic locking, [Put]
// and [Update] will modify the item's in-place timestamp fields before marshaling it to the `map[string]AttributeValue`
// that is sent to DynamoDB. And on failure, the fields will be restored to their original values, while on success, you
// can reasonably assume that the timestamps in item match what are in DynamoDB.
//
// Only last-modified time is always updated to [time.Now]; created time is only set if its original value is not a zero
// value.
package ddb

import (
	"context"
	"reflect"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// CreateTable uses [DynamoDB CreateTable] to add a new table and wait for the table to become active.
//
// The item argument must be a struct or struct pointer that is parseable by [model.NewForTypeOf]; it will not be
// modified by CreateTable.
//
// By default, [dynamodb.CreateTableInput.BillingMode] is set to [types.BillingModePayPerRequest] so that optFns is
// optional; otherwise, you'll have to explicitly provide [dynamodb.CreateTableInput.ProvisionedThroughput] for the
// input parameters to be valid.
//
// [DynamoDB CreateTable]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateTable.html
func CreateTable(ctx context.Context, item any, optFns ...func(opts *config.CreateTableOptions)) (err error) {
	c := internal.ApplyOpts(&config.CreateTableOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = newModel(item); err != nil {
		return err
	}

	return c.Execute(ctx)
}

// Delete uses [DynamoDB DeleteItem] to delete a single item.
//
// The item argument must be a struct or struct pointer that is parseable by [model.NewForTypeOf]; it will not be
// modified while preparing for the [dynamodb.Client.DeleteItem] call. If optimistic locking is enabled (default), and
// only if the item's version is not the zero value, a `#version = :version` condition is added.
//
// [DynamoDB DeleteItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DeleteItem.html
func Delete(ctx context.Context, item any, optFns ...func(opts *config.DeleteOptions)) (_ *dynamodb.DeleteItemOutput, err error) {
	c := internal.ApplyOpts(&config.DeleteOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = newModel(item); err != nil {
		return nil, err
	}

	return c.Execute(ctx, item)
}

// Get uses [DynamoDB GetItem] to retrieve attributes for the specified item.
//
// The item argument must be a struct or struct pointer that is parseable by [model.NewForTypeOf]; its must have its key
// attributes filled out to produce the [dynamodb.GetItemInput.Key] field. On success, all attributes from response
// ([dynamodb.GetItemOutput.Item]) are unmarshalled and set to the respective item's fields. Any field that doesn't
// exist in the response, possibly due to projection expression, will be set to their zero values. If the request did
// (`len(getItemOutput.Item) == 0` indicates item doesn't exist), the item will not be modified.
//
// [DynamoDB GetItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_GetItem.html
func Get(ctx context.Context, item any, optFns ...func(opts *config.GetOptions)) (_ *dynamodb.GetItemOutput, err error) {
	c := internal.ApplyOpts(&config.GetOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = newModel(item); err != nil {
		return nil, err
	}

	return c.Execute(ctx, item)
}

// Put uses [DynamoDB PutItem] to create a new item, or replace an old item with a new item.
//
// The item argument must be parseable by [model.NewForTypeOf], and must be a struct pointer since the struct's fields
// may be modified on success.
//
// [DynamoDB PutItem] uses clobbering behaviour: the item argument is marshaled with to produce the
// [dynamodb.PutItemInput.Item] that completely replaces or adds a new item in DynamoDB. As a result, Put differs from
// [Update] on how optimistic locking and timestamp generations work while calculating the [dynamodb.Client.PutItem]
// parameters:
//   - The item's version will be updated in-place to the next value. If the original version is the zero value,
//     `attribute_not_exists(#hashkey)` is the [condition expression] to prevent overwriting an existing item;
//     otherwise, `#version = :version` is used.
//   - The item's created time is only updated in-place when the original value is the zero value; its modified time is
//     always set to [time.Now].
//
// Any non-nil error will restore the item argument to its original state.
//
// [DynamoDB PutItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_PutItem.html
// [condition expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html
func Put(ctx context.Context, item any, optFns ...func(opts *config.PutOptions)) (_ *dynamodb.PutItemOutput, err error) {
	c := internal.ApplyOpts(&config.PutOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = newModel(item); err != nil {
		return nil, err
	}

	return c.Execute(ctx, item)
}

// Update uses [DynamoDB UpdateItem] to edit an existing item's attributes, or adds a new item if it does not already
// exist.
//
// The item argument must be parseable by [model.NewForTypeOf], and must be a struct pointer since the struct's fields
// may be modified on success.
//
// Because [DynamoDB UpdateItem] only requires the key of the item ([dynamodb.UpdateItemInput.Key]) while the
// [update expression] ([dynamodb.UpdateItemOutput.UpdateExpression]) effects changes to the item's attributes, you will
// want to use optFns to add at least one update clause. Use [config.UpdateOptions.Set], [config.UpdateOptions.Remove],
// [config.UpdateOptions.Add], [config.UpdateOptions.Delete], and/or [config.UpdateOptions.WithUpdateBehaviour] to do
// so. For the same reason, [Update] differs from [Put] on how optimistic locking and timestamps generation work:
//   - The item's version will be updated in-place to the next value. If the original version is the zero value,
//     `attribute_not_exists(#hashkey)` is the [condition expression] to prevent overwriting an existing item;
//     otherwise, `#version = :version` is used. Unlike [Put] which marshals the entire item as
//     `map[string]types.AttributeValue` to update the version attribute in DynamoDB, [Update] adds a
//     `SET #version :version` clause to the update expression instead.
//   - The item's modified time is always updated to [time.Now]; its created time is untouched. Unlike [Put]
//     which marshals the entire item as `map[string]types.AttributeValue` to update the modified time attribute in
//     DynamoDB, [Update] adds a `SET #modifiedTime :now` clause to the update expression instead.
//
// Any non-nil error will restore the item argument to its original state.
//
// [DynamoDB UpdateItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html
// [update expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.UpdateExpressions.html
// [condition expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html
func Update(ctx context.Context, item any, optFns ...func(opts *config.UpdateOptions)) (_ *dynamodb.UpdateItemOutput, err error) {
	c := internal.ApplyOpts(&config.UpdateOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = newModel(item); err != nil {
		return nil, err
	}

	return c.Execute(ctx, item)
}

// UpdateReturnAllNewValues is a variation of [Update] that uses [ALL_NEW return values] to update the given item
// argument on success.
//
// The item argument must be parseable by [model.NewForTypeOf], and must be a struct pointer since the struct's fields
// may be modified on success.
//
// [ALL_NEW return values]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html#DDB-UpdateItem-request-ReturnValues
func UpdateReturnAllNewValues(ctx context.Context, item any, optFns ...func(opts *config.UpdateOptions)) (_ *dynamodb.UpdateItemOutput, err error) {
	c := internal.ApplyOpts(&config.UpdateOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = newModel(item); err != nil {
		return nil, err
	}
	c.ReturnAllNewValues = true

	return c.Execute(ctx, item)
}

var (
	_     uuid.UUID
	_     types.BillingMode
	cache sync.Map
)

// newModel caches the model for subsequent usages.
func newModel(i any) (*model.TableModel, error) {
	t, err := internal.IndirectTypeIsStruct(reflect.TypeOf(i), false)
	if err == nil {
		if m, ok := cache.Load(t); ok {
			return m.(*model.TableModel), nil
		}
	}

	m, err := model.New(t)
	if err == nil {
		cache.Store(m.StructType, m)
	}

	return m, err
}

// defaultConfig creates a [config.Config] with its [config.Config.Client] set to DefaultClientProvider's return value.
func defaultConfig(ctx context.Context) (cfg config.Config) {
	if c, err := config.DefaultClientProvider.Provide(ctx); err == nil {
		cfg.Client = c
	}

	return
}
