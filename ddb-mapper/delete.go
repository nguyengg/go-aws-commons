package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/client"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// Delete is a wrapper around [mapper.Mapper.Delete].
//
// The item argument must be a struct or struct pointer that is parseable by [mapper.New].
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func Delete(ctx context.Context, item any, optFns ...func(opts *DeleteOptions)) (_ *dynamodb.DeleteItemOutput, err error) {
	c := internal.ApplyOpts(&DeleteOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = model.NewForTypeOf(item); err != nil {
		return nil, err
	}

	return c.Execute(ctx, item)
}

// DeleteOptions customises Delete.
//
// DeleteOptions can be modified either by changing the fields directly or via chaining With methods.
type DeleteOptions struct {
	config.Config

	// DisableOptimisticLocking, if true, will disable optimistic locking functionality.
	DisableOptimisticLocking bool

	tableName *string
	condition expression.ConditionBuilder
	inputFn   func(input *dynamodb.DeleteItemInput)
	optFns    []func(opts *dynamodb.Options)
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

// Resolve creates the internal [client.ItemDeleter].
func (opts *DeleteOptions) Resolve() *client.ItemDeleter {
	return &client.ItemDeleter{
		Config:                   opts.Config,
		DisableOptimisticLocking: opts.DisableOptimisticLocking,
		TableNameOverride:        opts.tableName,
		Condition:                opts.condition,
		InputFn:                  opts.inputFn,
		OptFns:                   opts.optFns,
	}
}
