// Package mapper provides [Mapper] to interact with DynamoDB tables in a type-safe way.
//
// See [New] for information on how to use `dynamodbav` struct tags on T to model the table.
package mapper

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/ddb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/untyped"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/types"
)

// Mapper provides a type-safe way to interact with the DynamoDB table containing items of type T.
//
// See [New] for information on how to use `dynamodbav` struct tags on T to model the table.
type Mapper[T any] struct {
	*untyped.Mapper

	ddb.Config
}

// CreateTable uses [DynamoDB CreateTable] to add a new table and wait for the table to become active.
//
// By default, [dynamodb.CreateTableInput.BillingMode] is set to [types.BillingModePayPerRequest] so that optFns is
// optional; otherwise, you'll have to explicitly provide [dynamodb.CreateTableInput.ProvisionedThroughput] for the
// input parameters to be valid.
//
// [DynamoDB CreateTable]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateTable.html
func (m *Mapper[T]) CreateTable(ctx context.Context, optFns ...func(opts *ddb.CreateTableOptions)) error {
	return m.Mapper.CreateTable(ctx, internal.ApplyOpts(&ddb.CreateTableOptions{}, optFns...).CopyTo)
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
	return m.Mapper.Delete(ctx, item, internal.ApplyOpts(&ddb.DeleteOptions{}, optFns...).CopyTo)
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
	return m.Mapper.Get(ctx, item, internal.ApplyOpts(&ddb.GetOptions{}, optFns...).CopyTo)
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
	return m.Mapper.Put(ctx, item, internal.ApplyOpts(&ddb.PutOptions{}, optFns...).CopyTo)
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
//     expression. This logic can be disabled with [PutOptions.DisableOptimisticLocking], especially if you're going to
//     manually update the version.
//   - The item's modified time attribute is always updated to [time.Now]; its created time is untouched. A
//     `SET #modifiedTime :now` clause is added to the update expression. If Mapper.Update is being used to create a new
//     item, you may want to manually set the same created time and modified time, and disable timestamp generation with
//     [UpdateOptions.DisableAutoGeneratedTimestamps].
//
// [DynamoDB UpdateItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html
// [update expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.UpdateExpressions.html
// [condition expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html
func (m *Mapper[T]) Update(ctx context.Context, item *T, optFns ...func(opts *ddb.UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	return m.Mapper.Update(ctx, item, internal.ApplyOpts(&ddb.UpdateOptions{}, optFns...).CopyTo)
}

// TableName returns the name of the table modeled by this [Mapper].
func (m *Mapper[T]) TableName() string {
	return m.Mapper.TableName
}

// HashKey returns the required hash key attribute.
func (m *Mapper[T]) HashKey() types.Attribute {
	return m.Mapper.HashKey
}

// SortKey returns the optional sort key attribute.
func (m *Mapper[T]) SortKey() types.Attribute {
	return m.Mapper.SortKey
}

// All returns an iterator over all attributes modeled by this [Mapper].
func (m *Mapper[T]) All() iter.Seq[types.Attribute] {
	return func(yield func(types.Attribute) bool) {
		if !yield(m.Mapper.HashKey) {
			return
		}
		if m.Mapper.SortKey != nil && !yield(m.Mapper.SortKey) {
			return
		}
		if m.Mapper.Version != nil && !yield(m.Mapper.Version) {
			return
		}
		if m.CreatedTime != nil && !yield(m.CreatedTime) {
			return
		}
		if m.ModifiedTime != nil && !yield(m.ModifiedTime) {
			return
		}

		for _, attr := range m.Others {
			if !yield(attr) {
				return
			}
		}
	}
}

// wrapVersionUpdater returns the version updater function that [untyped.Mapper] uses.
func (m *Mapper[T]) wrapVersionUpdater() func(item any) {
	if fn := m.VersionUpdater; fn != nil {
		return func(item any) {
			m.VersionUpdater(item.(*T))
		}
	}

	return nil
}
