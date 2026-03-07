package untyped

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
)

func (m *Mapper) Delete(ctx context.Context, item any, optFns ...func(opts *DeleteOptions)) (*dynamodb.DeleteItemOutput, error) {
	var (
		opts  = internal.ApplyOpts(&DeleteOptions{}, optFns...)
		reset = internal.Chainable()
	)

	c, err := Merge(ctx, m.Options, opts.Options)
	if err != nil {
		return nil, err
	}

	v, ptr, err := internal.IndirectValueIsStruct(item, true, m.StructType)
	if err != nil {
		return nil, err
	}

	if !opts.DisableOptimisticLocking && m.Version != nil {
		undo, cond := m.updateVersion(c, v, ptr)
		if opts.Condition.IsSet() {
			opts.Condition = opts.Condition.And(cond)
		} else {
			opts.Condition = cond
		}
		reset = reset.And(undo)
	}

	input := &dynamodb.DeleteItemInput{TableName: aws.String(m.TableName)}
	if input.Key, err = m.keys(c, item); err != nil {
		return nil, err
	}

	if opts.Condition.IsSet() {
		expr, err := expression.NewBuilder().WithCondition(opts.Condition).Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}

		input.ConditionExpression = expr.Condition()
		input.ExpressionAttributeNames = expr.Names()
		input.ExpressionAttributeValues = expr.Values()
	}
	if opts.TableName != nil {
		input.TableName = opts.TableName
	}
	if opts.InputFn != nil {
		opts.InputFn(input)
	}

	deleteItemOutput, err := c.Client.DeleteItem(ctx, input)
	if err != nil {
		reset()
		return deleteItemOutput, fmt.Errorf("dynamodb DeleteItem error: %w", err)
	}

	return deleteItemOutput, err
}

type DeleteOptions struct {
	Options

	DisableOptimisticLocking bool
	TableName                *string
	Condition                expression.ConditionBuilder
	InputFn                  func(input *dynamodb.DeleteItemInput)
	OptFns                   []func(opts *dynamodb.Options)
}
