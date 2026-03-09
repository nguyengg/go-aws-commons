// Package mapper provides [Mapper] to interact with DynamoDB tables in a type-safe way.
package mapper

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
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
// global Mapper variable in the same package that defines the struct that models the item:
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
func New[T any](optFns ...func(m *Mapper[T])) (m *Mapper[T], err error) {
	tType := reflect.TypeFor[T]()
	if tType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("New[T] requires T to be a struct kind of type, not %s", tType)
	}

	m = internal.ApplyOpts(&Mapper[T]{}, optFns...)
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
func NewMustHave[T any](flags model.AttributeModelType, optFns ...func(m *Mapper[T])) (*Mapper[T], error) {
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
func (m *Mapper[T]) CreateTable(ctx context.Context, optFns ...func(opts *ddb.CreateTableOptions)) error {
	c := internal.ApplyOpts(&ddb.CreateTableOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx)
}

// Delete uses [DynamoDB DeleteItem] to delete a single item.
//
// The item argument will not be modified while preparing for the [dynamodb.Client.DeleteItem] call.
//
// If item's version attribute is zero value, no condition will be added. Otherwise, a `#version = :eversion` condition
// will be added.
//
// [DynamoDB DeleteItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DeleteItem.html
func (m *Mapper[T]) Delete(ctx context.Context, item *T, optFns ...func(opts *ddb.DeleteOptions)) (*dynamodb.DeleteItemOutput, error) {
	c := internal.ApplyOpts(&ddb.DeleteOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx, item)
}

// Get uses [DynamoDB GetItem] to retrieve attributes for the specified item.
//
// The item argument must have its key attributes filled out to produce the [dynamodb.GetItemInput.Key] field. On
// success, all attributes from response ([dynamodb.GetItemOutput.Item]) are unmarshalled with [Options.Decoder] and set
// to the respective item's fields. Any field that doesn't exist in the response, possibly due to projection expression,
// will be set to their zero values. If the request did not succeed, or if the response is empty
// (`len(getItemOutput.Item) == 0` indicates item doesn't exist), the item will not be modified.
//
// [DynamoDB GetItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_GetItem.html
func (m *Mapper[T]) Get(ctx context.Context, item *T, optFns ...func(opts *ddb.GetOptions)) (*dynamodb.GetItemOutput, error) {
	c := internal.ApplyOpts(&ddb.GetOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx, item)
}

// Put uses [DynamoDB PutItem] to create a new item, or replace an old item with a new item.
//
// [DynamoDB PutItem] uses clobbering behaviour: the item argument is marshaled with [Options.Encoder] to produce the
// [dynamodb.PutItemInput.Item] that completely replaces or adds a new item in DynamoDB. As a result, Mapper.Put differs
// from [Mapper.Update] on how optimistic locking and timestamp generations work while preparing for the
// [dynamodb.Client.PutItem] call:
//   - The item's version will be updated to the next value using [Options.NextVersion]. If the original version is the
//     zero value, `attribute_not_exists(#hashkey)` is the [condition expression] to prevent overwriting an existing
//     item; otherwise, `#version = :version` is used. This logic can be disabled with
//     [PutOptions.DisableOptimisticLocking].
//   - The item's created time is only updated when the original value is the zero value; its modified time is always
//     updated to [time.Now].
//
// Any non-nil error will restore the item argument to its original state.
//
// [DynamoDB PutItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_PutItem.html
// [condition expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html
func (m *Mapper[T]) Put(ctx context.Context, item *T, optFns ...func(opts *ddb.PutOptions)) (*dynamodb.PutItemOutput, error) {
	c := internal.ApplyOpts(&ddb.PutOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx, item)
}

// Update uses [DynamoDB UpdateItem] to edit an existing item's attributes, or adds a new item if it does not already
// exist.
//
// Because [DynamoDB UpdateItem] only requires the key of the item ([dynamodb.UpdateItemInput.Key]) while the
// [update expression] ([dynamodb.UpdateItemOutput.UpdateExpression]) effects changes to the item's attributes, you will
// // want to use optFns to add at least one update clause. Use [UpdateOptions.Set], [UpdateOptions.Remove],
// [UpdateOptions.Add], [UpdateOptions.Delete], and/or [UpdateOptions.WithUpdateBehaviour] to do so. For the same
// reason, Mapper.Update differs from [Mapper.Put] on how optimistic locking and timestamps generation work while
// preparing for the [dynamodb.Client.UpdateItem] call:
//   - The item's version will be updated to the next value using [Config.NextVersion]. If the original version is the
//     zero value, `attribute_not_exists(#hashkey)` is the [condition expression] to prevent overwriting an existing
//     item; otherwise, `#version = :version` is used. A `SET #version :version` clause is added to the update
//     expression. This logic can be disabled with [ddb.UpdateOptions.DisableOptimisticLocking], especially if you're
//     going to manually update the version and add condition.
//   - The item's modified time attribute is always updated to [time.Now]; its created time is untouched. A
//     `SET #modifiedTime :now` clause is added to the update expression. If Mapper.Update is being used to create a new
//     item, you may want to manually set the same created time and modified time, and disable timestamp generation with
//     [UpdateOptions.DisableAutoGeneratedTimestamps].
//
// [DynamoDB UpdateItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html
// [update expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.UpdateExpressions.html
// [condition expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html
func (m *Mapper[T]) Update(ctx context.Context, item *T, optFns ...func(opts *ddb.UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	c := internal.ApplyOpts(&ddb.UpdateOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	return c.Execute(ctx, item)
}

// UpdateReturnAllNewValue is a variation of Update that uses [ALL_NEW return values] to update the given item argument.
//
// [ALL_NEW return values]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html#DDB-UpdateItem-request-ReturnValues
func (m *Mapper[T]) UpdateReturnAllNewValue(ctx context.Context, item *T, optFns ...func(opts *ddb.UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	c := internal.ApplyOpts(&ddb.UpdateOptions{Config: m.defaultConfig()}, optFns...).Resolve()
	c.TableModel = m.tableModel
	c.ReturnAllNewValues = true
	return c.Execute(ctx, item)
}

func (m *Mapper[T]) defaultConfig() config.Config {
	return config.Config{
		Client:         m.Client,
		Encoder:        m.Encoder,
		Decoder:        m.Decoder,
		VersionUpdater: m.VersionUpdater,
	}
}
