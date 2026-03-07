package mapper

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/untyped"
)

// CreateTable uses [DynamoDB CreateTable] to add a new table and wait for the table to become active.
//
// By default, [dynamodb.CreateTableInput.BillingMode] is set to [types.BillingModePayPerRequest] so that optFns is
// optional; otherwise, you'll have to explicitly provide [dynamodb.CreateTableInput.ProvisionedThroughput] for the
// input parameters to be valid.
//
// [DynamoDB CreateTable]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateTable.html
func (m *Mapper[T]) CreateTable(ctx context.Context, optFns ...func(opts *CreateTableOptions)) error {
	return m.Mapper.CreateTable(ctx, internal.ApplyOpts(&CreateTableOptions{}, optFns...).CopyTo)
}

// CreateTableOptions customises [CreateTable].
//
// CreateTableOptions can be modified either by changing the fields directly or via chaining With methods.
type CreateTableOptions struct {
	// MaxWait is the amount of time to wait until table exists.
	//
	// If given a non-positive amount, waiting is skipped. Defaults to 3 minutes.
	MaxWait time.Duration

	tableName *string
	inputFn   func(input *dynamodb.CreateTableInput)
	optFns    []func(*dynamodb.Options)
}

func (opts *CreateTableOptions) CopyTo(untypedOpts *untyped.CreateTableOptions) {
	untypedOpts.MaxWait = opts.MaxWait
	untypedOpts.TableName = opts.tableName
	untypedOpts.InputFn = opts.inputFn
	untypedOpts.OptFns = opts.optFns
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
