package config

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/client"
)

// CreateTableOptions customises a single [DynamoDB CreateTable] call.
//
// CreateTableOptions can be modified either by changing the fields directly or via chaining With methods.
//
// [DynamoDB CreateTable]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateTable.html
type CreateTableOptions struct {
	// Config customises these settings at the operation level.
	Config

	// MaxWait is the amount of time to wait until table exists.
	//
	// If given a non-positive amount, waiting is skipped. Defaults to 3 minutes.
	MaxWait time.Duration

	tableName *string
	inputFn   func(input *dynamodb.CreateTableInput)
	optFns    []func(*dynamodb.Options)
}

// WithTableNameOverride overrides the table name.
func (opts *CreateTableOptions) WithTableNameOverride(tableName string) *CreateTableOptions {
	opts.tableName = &tableName
	return opts
}

// WithInputOptions modifies the [dynamodb.CreateTableInput] parameters right before invoking DynamoDB.
func (opts *CreateTableOptions) WithInputOptions(fn func(input *dynamodb.CreateTableInput)) *CreateTableOptions {
	opts.inputFn = fn
	return opts
}

// WithClientOptions attaches options to the [dynamodb.Client.CreateTable] invocation.
func (opts *CreateTableOptions) WithClientOptions(optFns ...func(opts *dynamodb.Options)) *CreateTableOptions {
	opts.optFns = optFns
	return opts
}

// Resolve creates the internal [client.TableCreator].
func (opts *CreateTableOptions) Resolve() *client.TableCreator {
	return &client.TableCreator{
		Config: client.Config{
			Client:         opts.Client,
			Encoder:        opts.Encoder,
			Decoder:        opts.Decoder,
			VersionUpdater: opts.VersionUpdater,
		},
		MaxWait:           opts.MaxWait,
		TableNameOverride: opts.tableName,
		InputFn:           opts.inputFn,
		OptFns:            opts.optFns,
	}
}
