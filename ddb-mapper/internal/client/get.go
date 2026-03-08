package client

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/ddb/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// Getter is the client for executing [DynamoDB GetItem].
//
// [DynamoDB GetItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_GetItem.html
type Getter struct {
	config.Config
	model.TableModel

	TableNameOverride *string
	InputFn           func(input *dynamodb.GetItemInput)
	OptFns            []func(opts *dynamodb.Options)
}

func (c *Getter) Execute(ctx context.Context, item any) (getItemOutput *dynamodb.GetItemOutput, err error) {
	if err = initConfig(ctx, &c.Config); err != nil {
		return nil, err
	}

	v, ptr, err := internal.IndirectValueIsStruct(item, true, c.StructType)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.GetItemInput{TableName: aws.String(c.TableName)}
	if input.Key, err = c.EncodeKeys(item, c.CopyTo); err != nil {
		return nil, err
	}

	if c.TableNameOverride != nil {
		input.TableName = c.TableNameOverride
	}
	if c.InputFn != nil {
		c.InputFn(input)
	}

	if getItemOutput, err = c.Client.GetItem(ctx, input, c.OptFns...); err != nil {
		return getItemOutput, fmt.Errorf("dynamodb GetItem error: %w", err)
	}

	if len(getItemOutput.Item) != 0 {
		v.Set(reflect.Zero(c.StructType))

		if err = c.Decoder.Decode(&types.AttributeValueMemberM{Value: getItemOutput.Item}, ptr.Interface()); err != nil {
			return nil, fmt.Errorf("unmarshal item to type %T error: %w", item, err)
		}
	}

	return getItemOutput, nil
}
