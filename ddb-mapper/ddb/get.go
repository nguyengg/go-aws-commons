package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/untyped"
)

// Get is a wrapper around [mapper.Mapper.Get].
//
// The item argument must be parseable by [mapper.New], and must be a struct pointer since the struct's fields may be
// modified on success.
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func Get(ctx context.Context, item any, optFns ...func(opts *GetOptions)) (*dynamodb.GetItemOutput, error) {
	client, err := DefaultClientProvider.Provide(ctx)
	if err != nil {
		return nil, err
	}

	m, err := untyped.NewFromItem(item, func(opts *untyped.Options) { opts.Client = client })
	if err != nil {
		return nil, err
	}

	return m.Get(ctx, item, internal.ApplyOpts(&GetOptions{}, optFns...).CopyTo)
}

// GetOptions customises Get.
//
// GetOptions can be modified either by changing the fields directly or via chaining With methods.
type GetOptions struct {
	tableName *string
	inputFn   func(input *dynamodb.GetItemInput)
	optFns    []func(opts *dynamodb.Options)
}

func (opts *GetOptions) CopyTo(untypedOpts *untyped.GetOptions) {
	untypedOpts.TableName = opts.tableName
	untypedOpts.InputFn = opts.inputFn
	untypedOpts.OptFns = opts.optFns
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
