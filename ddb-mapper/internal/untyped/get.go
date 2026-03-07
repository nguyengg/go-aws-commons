package untyped

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
)

func (m *Mapper) Get(ctx context.Context, item any, optFns ...func(opts *GetOptions)) (*dynamodb.GetItemOutput, error) {
	opts := internal.ApplyOpts(&GetOptions{}, optFns...)
	c, err := Merge(ctx, m.Options, opts.Options)
	if err != nil {
		return nil, err
	}

	v, ptr, err := internal.IndirectValueIsStruct(item, true, m.StructType)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.GetItemInput{TableName: aws.String(m.TableName)}
	if input.Key, err = m.keys(c, item); err != nil {
		return nil, err
	}

	if opts.TableName != nil {
		input.TableName = opts.TableName
	}
	if opts.InputFn != nil {
		opts.InputFn(input)
	}

	getItemOutput, err := c.Client.GetItem(ctx, input, opts.OptFns...)
	if err != nil {
		return getItemOutput, fmt.Errorf("dynamodb GetItem error: %w", err)
	}

	if len(getItemOutput.Item) != 0 {
		v.Set(reflect.Zero(m.StructType))

		if err = c.Decoder.Decode(&types.AttributeValueMemberM{Value: getItemOutput.Item}, ptr.Interface()); err != nil {
			return nil, fmt.Errorf("unmarshal item to type %T error: %w", item, err)
		}
	}

	return getItemOutput, nil
}

type GetOptions struct {
	Options

	TableName *string
	InputFn   func(input *dynamodb.GetItemInput)
	OptFns    []func(opts *dynamodb.Options)
}
