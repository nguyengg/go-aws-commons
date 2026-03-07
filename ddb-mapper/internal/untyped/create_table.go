package untyped

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
)

func (m *Mapper) CreateTable(ctx context.Context, optFns ...func(opts *CreateTableOptions)) error {
	opts := internal.ApplyOpts(&CreateTableOptions{MaxWait: 3 * time.Minute}, optFns...)
	c, err := Merge(ctx, m.Options, opts.Options)
	if err != nil {
		return err
	}

	input := &dynamodb.CreateTableInput{
		TableName:   aws.String(m.TableName),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(m.HashKey.AttrName),
				AttributeType: m.HashKey.ScalarAttributeType,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(m.HashKey.AttrName),
				KeyType:       types.KeyTypeHash,
			},
		},
	}

	if m.SortKey != nil {
		input.AttributeDefinitions = append(input.AttributeDefinitions, types.AttributeDefinition{
			AttributeName: aws.String(m.SortKey.AttrName),
			AttributeType: m.SortKey.ScalarAttributeType,
		})
		input.KeySchema = append(input.KeySchema, types.KeySchemaElement{
			AttributeName: aws.String(m.SortKey.AttrName),
			KeyType:       types.KeyTypeRange,
		})
	}

	if opts.TableName != nil {
		input.TableName = opts.TableName
	}
	if opts.InputFn != nil {
		opts.InputFn(input)
	}

	if _, err = c.Client.CreateTable(ctx, input, opts.OptFns...); err != nil {
		return fmt.Errorf("dynamodb CreateTable error: %w", err)
	}

	if opts.MaxWait > 0 {
		if err := dynamodb.NewTableExistsWaiter(c.Client).
			Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(m.TableName)}, opts.MaxWait); err != nil {
			return fmt.Errorf("wait until table exists error: %w", err)
		}
	}

	return nil
}

type CreateTableOptions struct {
	Options

	MaxWait   time.Duration
	TableName *string
	InputFn   func(input *dynamodb.CreateTableInput)
	OptFns    []func(*dynamodb.Options)
}
