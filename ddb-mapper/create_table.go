package ddb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/client"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// CreateTable is a wrapper around [mapper.Mapper.CreateTable].
//
// The item argument must be a struct or struct pointer that is parseable by [mapper.New].
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func CreateTable(ctx context.Context, item any, optFns ...func(opts *CreateTableOptions)) (err error) {
	c := internal.ApplyOpts(&CreateTableOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = model.NewForTypeOf(item); err != nil {
		return err
	}

	return c.Execute(ctx)
}

// CreateTableOptions customises [CreateTable].
//
// CreateTableOptions can be modified either by changing the fields directly or via chaining With methods.
type CreateTableOptions struct {
	config.Config

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
		Config:            opts.Config,
		MaxWait:           opts.MaxWait,
		TableNameOverride: opts.tableName,
		InputFn:           opts.inputFn,
		OptFns:            opts.optFns,
	}
}
