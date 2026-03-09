package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/client"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// Get is a wrapper around [mapper.Mapper.Get].
//
// The item argument must be parseable by [mapper.New], and must be a struct pointer since the struct's fields may be
// modified on success.
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func Get(ctx context.Context, item any, optFns ...func(opts *GetOptions)) (_ *dynamodb.GetItemOutput, err error) {
	c := internal.ApplyOpts(&GetOptions{Config: defaultConfig(ctx)}, optFns...).Resolve()
	if c.TableModel, err = model.NewForTypeOf(item); err != nil {
		return nil, err
	}

	return c.Execute(ctx, item)
}

// GetOptions customises Get.
//
// GetOptions can be modified either by changing the fields directly or via chaining With methods.
type GetOptions struct {
	config.Config

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
		Config:            opts.Config,
		TableNameOverride: opts.tableName,
		InputFn:           opts.inputFn,
		OptFns:            opts.optFns,
	}
}
