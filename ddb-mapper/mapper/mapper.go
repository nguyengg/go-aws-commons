// Package mapper provides [Mapper] to interact with DynamoDB tables in a type-safe way.
package mapper

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// tableModel is not exported.
type tableModel = model.TableModel

// Mapper provides a type-safe way to interact with the DynamoDB table containing items of type T.
type Mapper[T any] struct {
	config.Config
	*tableModel // model MUST NOT be mutated.
}

// New creates a new [Mapper] modeling the DynamoDB table that contains items of type T.
//
// See [model.TableModel] for details regarding how the struct tags are parsed. A common usage pattern is to create a
// global [Mapper] variable in the same package that defines the struct that models the item:
//
//	package app
//
//	import (
//		mapper "github.com/nguyengg/go-aws-commons/ddb-mapper"
//	)
//
//	type Item struct {
//		ID string `json:"id" dynamodbav:"id,hashkey" tablename:"Items"`
//	}
//
//	var Mapper *mapper.Mapper[Item]
//
//	func init() {
//		var err error
//		Mapper, err = mapper.New[Item]()
//		if err != nil {
//			panic(err)
//		}
//	}
//
// Then the [Mapper] can be used like this:
//
//	item := app.Item{ID: "id"}
//	app.Mapper.Get(context.Background(), &item)
func New[T any](optFns ...func(cfg *config.Config)) (m *Mapper[T], err error) {
	tType := reflect.TypeFor[T]()
	if tType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("New[T] requires T to be a struct kind of type, not %s", tType)
	}

	m = &Mapper[T]{Config: *internal.ApplyOpts(&config.Config{}, optFns...)}
	if m.tableModel, err = model.NewForType[T](); err != nil {
		return nil, err
	}

	return m, nil
}

// NewMustHave is a variation of New that receives a flag to fail model validation of the model doesn't have the
// required attributes.
//
// Usage:
//
//	m, err := NewMustHave[T](AttributeModelTypeVersion | AttributeModelTypeCreatedTime | AttributeModelTypeModifiedTime)
//	if err != nil {
//		panic(err)
//	}
func NewMustHave[T any](flags model.AttributeModelType, optFns ...func(cfg *config.Config)) (*Mapper[T], error) {
	m, err := New[T](optFns...)
	if err == nil {
		err = model.MustHave(m.tableModel, flags)
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// CreateTable uses [DynamoDB CreateTable] to add a new table and wait for the table to become active.
//
// By default, [dynamodb.CreateTableInput.BillingMode] is set to [types.BillingModePayPerRequest] so that optFns is
// optional; otherwise, you'll have to explicitly provide [dynamodb.CreateTableInput.ProvisionedThroughput] for the
// input parameters to be valid.
//
// [DynamoDB CreateTable]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateTable.html
func (m *Mapper[T]) CreateTable(ctx context.Context, optFns ...func(opts *config.CreateTableOptions)) error {
	c := internal.ApplyOpts(&config.CreateTableOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx)
}

// Delete uses [DynamoDB DeleteItem] to delete a single item.
//
// The item argument will not be modified while preparing for the [dynamodb.Client.DeleteItem] call. If optimistic
// locking is enabled (default), and only if the item's version is not the zero value, a `#version = :version` condition
// is added.
//
// [DynamoDB DeleteItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DeleteItem.html
func (m *Mapper[T]) Delete(ctx context.Context, item *T, optFns ...func(opts *config.DeleteOptions)) (*dynamodb.DeleteItemOutput, error) {
	c := internal.ApplyOpts(&config.DeleteOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx, item)
}

// Get uses [DynamoDB GetItem] to retrieve attributes for the specified item.
//
// The item argument must have its key attributes filled out to produce the [dynamodb.GetItemInput.Key] field. On
// success, all attributes from response ([dynamodb.GetItemOutput.Item]) are unmarshalled and set to the respective
// item's fields. Any field that doesn't exist in the response, possibly due to projection expression, will be set to
// their zero values. If the request did not succeed, or if the response is empty (`len(getItemOutput.Item) == 0`
// indicates item doesn't exist), the item will not be modified.
//
// [DynamoDB GetItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_GetItem.html
func (m *Mapper[T]) Get(ctx context.Context, item *T, optFns ...func(opts *config.GetOptions)) (*dynamodb.GetItemOutput, error) {
	c := internal.ApplyOpts(&config.GetOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx, item)
}

// Put uses [DynamoDB PutItem] to create a new item, or replace an old item with a new item.
//
// [DynamoDB PutItem] uses clobbering behaviour: the item argument is marshaled to produce the
// [dynamodb.PutItemInput.Item] that completely replaces or adds a new item in DynamoDB. As a result, [Mapper.Put]
// differs from [Mapper.Update] on how optimistic locking and timestamp generations work while calculating the
// [dynamodb.Client.PutItem] parameters:
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
func (m *Mapper[T]) Put(ctx context.Context, item *T, optFns ...func(opts *config.PutOptions)) (*dynamodb.PutItemOutput, error) {
	c := internal.ApplyOpts(&config.PutOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx, item)
}

// Update uses [DynamoDB UpdateItem] to edit an existing item's attributes, or adds a new item if it does not already
// exist.
//
// Because [DynamoDB UpdateItem] only requires the key of the item ([dynamodb.UpdateItemInput.Key]) while the
// [update expression] ([dynamodb.UpdateItemOutput.UpdateExpression]) effects changes to the item's attributes, you will
// want to use optFns to add at least one update clause. Use [config.UpdateOptions.Set], [config.UpdateOptions.Remove],
// [config.UpdateOptions.Add], [config.UpdateOptions.Delete], and/or [config.UpdateOptions.WithUpdateBehaviour] to do
// so. For the same reason, [Mapper.Update] differs from [Mapper.Put] on how optimistic locking and timestamps
// generation work:
//   - The item's version will be updated in-place to the next value. If the original version is the zero value,
//     `attribute_not_exists(#hashkey)` is the [condition expression] to prevent overwriting an existing item;
//     otherwise, `#version = :version` is used. Unlike [Mapper.Put] which marshals the entire item as
//     `map[string]types.AttributeValue` to update the version attribute in DynamoDB, [Mapper.Update] adds a
//     `SET #version :version` clause to the update expression instead.
//   - The item's modified time is always updated to [time.Now]; its created time is untouched. Unlike [Mapper.Put]
//     which marshals the entire item as `map[string]types.AttributeValue` to update the modified time attribute in
//     DynamoDB, [Mapper.Update] adds a `SET #modifiedTime :now` clause to the update expression instead.
//
// Any non-nil error will restore the item argument to its original state.
//
// [DynamoDB UpdateItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html
// [update expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.UpdateExpressions.html
// [condition expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html
func (m *Mapper[T]) Update(ctx context.Context, item *T, optFns ...func(opts *config.UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	c := internal.ApplyOpts(&config.UpdateOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx, item)
}

// UpdateReturnAllNewValues is a variation of [Mapper.Update] that uses [ALL_NEW return values] to update the given item
// argument on success.
//
// [ALL_NEW return values]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html#DDB-UpdateItem-request-ReturnValues
func (m *Mapper[T]) UpdateReturnAllNewValues(ctx context.Context, item *T, optFns ...func(opts *config.UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	c := internal.ApplyOpts(&config.UpdateOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	c.ReturnAllNewValues = true
	return c.Execute(ctx, item)
}

// defaultConfig creates a [config.Config] that is copied from the same settings in Mapper.
func (m *Mapper[T]) defaultConfig() config.Config {
	return config.Config{
		Client:         m.Client,
		Encoder:        m.Encoder,
		Decoder:        m.Decoder,
		VersionUpdater: m.VersionUpdater,
	}
}
