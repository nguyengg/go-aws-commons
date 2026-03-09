package config

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/client"
)

// DeleteOptions customises a single [DynamoDB DeleteItem] call.
//
// DeleteOptions can be modified either by changing the fields directly or via chaining With methods.
//
// [DynamoDB DeleteItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DeleteItem.html
type DeleteOptions struct {
	// Config customises these settings at the operation level.
	Config

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
		Config: client.Config{
			Client:         opts.Client,
			Encoder:        opts.Encoder,
			Decoder:        opts.Decoder,
			VersionUpdater: opts.VersionUpdater,
		},
		DisableOptimisticLocking: opts.DisableOptimisticLocking,
		TableNameOverride:        opts.tableName,
		Condition:                opts.condition,
		InputFn:                  opts.inputFn,
		OptFns:                   opts.optFns,
	}
}
