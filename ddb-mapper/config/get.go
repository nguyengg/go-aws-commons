package config

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/client"
)

// GetOptions customises a single [DynamoDB GetItem] call.
//
// GetOptions can be modified either by changing the fields directly or via chaining With methods.
//
// [DynamoDB GetItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_GetItem.html
type GetOptions struct {
	// Config customises these settings at the operation level.
	Config

	tableName *string
	inputFn   func(input *dynamodb.GetItemInput)
	optFns    []func(opts *dynamodb.Options)
}

// WithTableNameOverride overrides the table name.
func (opts *GetOptions) WithTableNameOverride(tableName string) *GetOptions {
	opts.tableName = &tableName
	return opts
}

// WithInputOptions modifies the [dynamodb.GetItemInput] parameters right before invoking DynamoDB.
func (opts *GetOptions) WithInputOptions(fn func(input *dynamodb.GetItemInput)) *GetOptions {
	opts.inputFn = fn
	return opts
}

// WithClientOptions attaches options to the [dynamodb.Client.GetItem] invocation.
func (opts *GetOptions) WithClientOptions(optFns ...func(opts *dynamodb.Options)) *GetOptions {
	opts.optFns = optFns
	return opts
}

// Resolve creates the internal [client.ItemGetter].
func (opts *GetOptions) Resolve() *client.ItemGetter {
	return &client.ItemGetter{
		Config: client.Config{
			Client:         opts.Client,
			Encoder:        opts.Encoder,
			Decoder:        opts.Decoder,
			VersionUpdater: opts.VersionUpdater,
		},
		TableNameOverride: opts.tableName,
		InputFn:           opts.inputFn,
		OptFns:            opts.optFns,
	}
}
