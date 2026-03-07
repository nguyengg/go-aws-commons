package mapper

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/untyped"
)

// Delete uses [DynamoDB DeleteItem] to delete a single item.
//
// The item argument will not be modified while preparing for the [dynamodb.Client.DeleteItem] call.
//
// If item's version attribute is zero value, no condition will be added. Otherwise, a `#version = :eversion` condition
// will be added.
//
// [DynamoDB DeleteItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DeleteItem.html
func (m *Mapper[T]) Delete(ctx context.Context, item *T, optFns ...func(opts *DeleteOptions)) (*dynamodb.DeleteItemOutput, error) {
	return m.Mapper.Delete(ctx, item, internal.ApplyOpts(&DeleteOptions{}, optFns...).CopyTo)
}

// DeleteOptions customises Delete.
//
// DeleteOptions can be modified either by changing the fields directly or via chaining With methods.
type DeleteOptions struct {
	// DisableOptimisticLocking, if true, will disable optimistic locking functionality.
	DisableOptimisticLocking bool

	tableName *string
	condition expression.ConditionBuilder
	inputFn   func(input *dynamodb.DeleteItemInput)
	optFns    []func(opts *dynamodb.Options)
}

func (opts *DeleteOptions) CopyTo(untypedOpts *untyped.DeleteOptions) {
	untypedOpts.DisableOptimisticLocking = opts.DisableOptimisticLocking
	untypedOpts.TableName = opts.tableName
	untypedOpts.Condition = opts.condition
	untypedOpts.InputFn = opts.inputFn
	untypedOpts.OptFns = opts.optFns
}

// WithTableNameOverride overrides the table name.
func (opts *DeleteOptions) WithTableNameOverride(tableName string) *DeleteOptions {
	opts.tableName = &tableName
	return opts
}

// WithInputOptions modifies the [dynamodb.DeleteItemInput] parameters right before invoking DynamoDB.
func (opts *DeleteOptions) WithInputOptions(fn func(input *dynamodb.DeleteItemInput)) *DeleteOptions {
	opts.inputFn = fn
	return opts
}

// WithClientOptions attaches options to the [dynamodb.Client.DeleteItem] invocation.
func (opts *DeleteOptions) WithClientOptions(optFns ...func(opts *dynamodb.Options)) *DeleteOptions {
	opts.optFns = optFns
	return opts
}
