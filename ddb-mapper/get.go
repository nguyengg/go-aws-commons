package mapper

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/untyped"
)

// Get uses [DynamoDB GetItem] to retrieve attributes for the specified item.
//
// The item argument must have its key attributes filled out to produce the [dynamodb.GetItemInput.Key] field. On
// success, all attributes from response ([dynamodb.GetItemOutput.Item]) are unmarshalled with [Options.Decoder] and set
// to the respective item's fields. Any field that doesn't exist in the response, possibly due to projection expression,
// will be set to their zero values. If the request did not succeed, or if the response is empty
// (`len(getItemOutput.Item) == 0` indicates item doesn't exist), the item will not be modified.
//
// [DynamoDB GetItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_GetItem.html
func (m *Mapper[T]) Get(ctx context.Context, item *T, optFns ...func(opts *GetOptions)) (*dynamodb.GetItemOutput, error) {
	return m.Mapper.Get(ctx, item, internal.ApplyOpts(&GetOptions{}, optFns...).CopyTo)
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Common options.
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
