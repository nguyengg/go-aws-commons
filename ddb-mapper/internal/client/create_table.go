package client

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/ddb/config"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// TableCreator is the client for executing [DynamoDB CreateTable].
//
// [DynamoDB CreateTable]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_CreateTable.html
type TableCreator struct {
	config.Config
	model.TableModel

	MaxWait           time.Duration
	TableNameOverride *string
	InputFn           func(input *dynamodb.CreateTableInput)
	OptFns            []func(opts *dynamodb.Options)
}

func (m *TableCreator) Execute(ctx context.Context) (err error) {
	if err = initConfig(ctx, &m.Config); err != nil {
		return err
	}

	// try to extract the hash key type
	scalarType := guessScalarType(m.HashKey.Type)
	if scalarType == "" {
		return fmt.Errorf("cannot determine hashkey's scalar type: %s{%q: %s}", m.StructType, m.HashKey.Name, m.HashKey.Type)
	}

	input := &dynamodb.CreateTableInput{
		TableName:   aws.String(m.TableName),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(m.HashKey.AttrName),
				AttributeType: scalarType,
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
		if scalarType = guessScalarType(m.SortKey.Type); scalarType == "" {
			return fmt.Errorf("cannot determine sortkey's scalar type: %s{%q: %s}", m.StructType, m.HashKey.Name, m.HashKey.Type)
		}

		input.AttributeDefinitions = append(input.AttributeDefinitions, types.AttributeDefinition{
			AttributeName: aws.String(m.SortKey.AttrName),
			AttributeType: scalarType,
		})
		input.KeySchema = append(input.KeySchema, types.KeySchemaElement{
			AttributeName: aws.String(m.SortKey.AttrName),
			KeyType:       types.KeyTypeRange,
		})
	}

	if m.TableNameOverride != nil {
		input.TableName = m.TableNameOverride
	}
	if m.InputFn != nil {
		m.InputFn(input)
	}

	if _, err = m.Client.CreateTable(ctx, input, m.OptFns...); err != nil {
		return fmt.Errorf("dynamodb CreateTable error: %w", err)
	}

	if m.MaxWait > 0 {
		if err := dynamodb.NewTableExistsWaiter(m.Client).
			Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(m.TableName)}, m.MaxWait); err != nil {
			return fmt.Errorf("wait until table exists error: %w", err)
		}
	}

	return nil
}

// guessScalarType uses the Go type to determine its DynamoDB scalar type.
func guessScalarType(t reflect.Type) types.ScalarAttributeType {
	// if t implements attributevalue.Marshaler then we can't rely on its kind.
	if !t.Implements(reflect.TypeFor[attributevalue.Marshaler]()) {
		switch k := t.Kind(); k {

		case reflect.String:
			return types.ScalarAttributeTypeS

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return types.ScalarAttributeTypeN

		case reflect.Array, reflect.Slice:
			if t == byteSliceType || t.Elem().Kind() == reflect.Uint8 {
				return types.ScalarAttributeTypeB
			}

		default:
			// fallthrough
		}
	}

	// let's try to marshal the zero value to see if we can still get something useful.
	v := reflect.Zero(t)
	if av, err := attributevalue.Marshal(v.Interface()); err == nil {
		switch av.(type) {
		case *types.AttributeValueMemberS:
			return types.ScalarAttributeTypeS
		case *types.AttributeValueMemberN:
			return types.ScalarAttributeTypeN
		case *types.AttributeValueMemberB:
			return types.ScalarAttributeTypeB
		}
	}

	return ""
}

var (
	byteSliceType = reflect.TypeFor[[]byte]()
)
