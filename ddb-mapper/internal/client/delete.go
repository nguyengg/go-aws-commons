package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// ItemDeleter is the client for executing [DynamoDB DeleteItem].
//
// [DynamoDB DeleteItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DeleteItem.html
type ItemDeleter struct {
	config.Config
	*model.TableModel // model MUST NOT be mutated.

	DisableOptimisticLocking bool
	TableNameOverride        *string
	Condition                expression.ConditionBuilder
	InputFn                  func(input *dynamodb.DeleteItemInput)
	OptFns                   []func(opts *dynamodb.Options)
}

func (c *ItemDeleter) Execute(ctx context.Context, item any) (deleteItemOutput *dynamodb.DeleteItemOutput, err error) {
	if err = initConfig(ctx, &c.Config); err != nil {
		return nil, err
	}

	v, ptr, err := internal.IndirectValueIsStruct(item, true, c.StructType)
	if err != nil {
		return nil, err
	}

	reset := internal.Chainable()
	if !c.DisableOptimisticLocking && c.Version != nil {
		undo, cond, err := updateVersion(c.TableModel, v, ptr, c.VersionUpdater)
		if err != nil {
			return nil, err
		} else if c.Condition.IsSet() {
			c.Condition = c.Condition.And(cond)
		} else {
			c.Condition = cond
		}
		reset = reset.And(undo)
	}

	input := &dynamodb.DeleteItemInput{TableName: aws.String(c.TableName)}
	if input.Key, err = c.EncodeKeys(item, c.CopyTo); err != nil {
		return nil, err
	}

	if c.Condition.IsSet() {
		expr, err := expression.NewBuilder().WithCondition(c.Condition).Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}

		input.ConditionExpression = expr.Condition()
		input.ExpressionAttributeNames = expr.Names()
		input.ExpressionAttributeValues = expr.Values()
	}
	if c.TableNameOverride != nil {
		input.TableName = c.TableNameOverride
	}
	if c.InputFn != nil {
		c.InputFn(input)
	}

	if deleteItemOutput, err = c.Client.DeleteItem(ctx, input); err != nil {
		reset()
		err = fmt.Errorf("dynamodb DeleteItem error: %w", err)
	}

	return
}
