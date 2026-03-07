// Package ddb provides type-agnostic accessors to interact with DynamoDB tables.
package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal/untyped"
)

// CreateTable is a wrapper around [mapper.Mapper.CreateTable].
//
// The item argument must be a struct or struct pointer that is parseable by [mapper.New].
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func CreateTable(ctx context.Context, item any, optFns ...func(opts *mapper.CreateTableOptions)) error {
	client, err := DefaultClientProvider.Provide(ctx)
	if err != nil {
		return err
	}

	m, err := untyped.NewFromItem(item, func(opts *untyped.Options) { opts.Client = client })
	if err != nil {
		return err
	}

	return m.CreateTable(ctx, internal.ApplyOpts(&mapper.CreateTableOptions{}, optFns...).CopyTo)
}

// Delete is a wrapper around [mapper.Mapper.Delete].
//
// The item argument must be a struct or struct pointer that is parseable by [mapper.New].
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func Delete(ctx context.Context, item any, optFns ...func(opts *mapper.DeleteOptions)) (*dynamodb.DeleteItemOutput, error) {
	client, err := DefaultClientProvider.Provide(ctx)
	if err != nil {
		return nil, err
	}

	m, err := untyped.NewFromItem(item, func(opts *untyped.Options) { opts.Client = client })
	if err != nil {
		return nil, err
	}

	return m.Delete(ctx, item, internal.ApplyOpts(&mapper.DeleteOptions{}, optFns...).CopyTo)
}

// Get is a wrapper around [mapper.Mapper.Get].
//
// The item argument must be parseable by [mapper.New], and must be a struct pointer since the struct's fields may be
// modified on success.
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func Get(ctx context.Context, item any, optFns ...func(opts *mapper.GetOptions)) (*dynamodb.GetItemOutput, error) {
	client, err := DefaultClientProvider.Provide(ctx)
	if err != nil {
		return nil, err
	}

	m, err := untyped.NewFromItem(item, func(opts *untyped.Options) { opts.Client = client })
	if err != nil {
		return nil, err
	}

	return m.Get(ctx, item, internal.ApplyOpts(&mapper.GetOptions{}, optFns...).CopyTo)
}

// Put is a wrapper around [mapper.Mapper.Put].
//
// The item argument must be parseable by [mapper.New], and must be a struct pointer since the struct's fields may be
// modified on success.
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func Put(ctx context.Context, item any, optFns ...func(opts *mapper.PutOptions)) (*dynamodb.PutItemOutput, error) {
	client, err := DefaultClientProvider.Provide(ctx)
	if err != nil {
		return nil, err
	}

	m, err := untyped.NewFromItem(item, func(opts *untyped.Options) { opts.Client = client })
	if err != nil {
		return nil, err
	}

	return m.Put(ctx, item, internal.ApplyOpts(&mapper.PutOptions{}, optFns...).CopyTo)
}

// Update is a wrapper around [mapper.Mapper.Update].
//
// The item argument must be parseable by [mapper.New], and must be a struct pointer since the struct's fields may be
// modified on success.
//
// Because [DynamoDB UpdateItem] only requires the key of the item ([dynamodb.UpdateItemInput.Key]) while the
// [update expression] ([dynamodb.UpdateItemOutput.UpdateExpression]) effects changes to the item's attributes, you will
// want to use optFns to add at least one update clause.
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
//
// [DynamoDB UpdateItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html
// [update expression]: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.UpdateExpressions.html
func Update(ctx context.Context, item any, optFns ...func(opts *mapper.UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	client, err := DefaultClientProvider.Provide(ctx)
	if err != nil {
		return nil, err
	}

	m, err := untyped.NewFromItem(item, func(opts *untyped.Options) { opts.Client = client })
	if err != nil {
		return nil, err
	}

	return m.Update(ctx, item, internal.ApplyOpts(&mapper.UpdateOptions{}, optFns...).CopyTo)
}

// UpdateReturnAllNewValues is a wrapper around [mapper.Mapper.UpdateReturnAllNewValues].
//
// The item argument must be parseable by [mapper.New], and must be a struct pointer since the struct's fields may be
// modified on success.
//
// [DefaultClientProvider] is used to retrieve the DynamoDB client to make the service calls.
func UpdateReturnAllNewValues(ctx context.Context, item any, optFns ...func(opts *mapper.UpdateOptions)) (*dynamodb.UpdateItemOutput, error) {
	client, err := DefaultClientProvider.Provide(ctx)
	if err != nil {
		return nil, err
	}

	m, err := untyped.NewFromItem(item, func(opts *untyped.Options) { opts.Client = client })
	if err != nil {
		return nil, err
	}

	return m.UpdateReturnAllNewValues(ctx, item, internal.ApplyOpts(&mapper.UpdateOptions{}, optFns...).CopyTo)
}
