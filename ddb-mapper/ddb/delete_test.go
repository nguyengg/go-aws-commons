package ddb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/ddb"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/ddb-local-test"
	. "github.com/nguyengg/go-aws-commons/must"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"Items"`
		Data    string `dynamodbav:"data"`
		Version int    `dynamodbav:"version,version"`
	}

	client := Setup(t, Item{})
	DefaultClientProvider = &StaticClientProvider{Client: client}

	// deleting an item that doesn't exist do nothing.
	_, err := Delete(t.Context(), &Item{ID: "tes"})
	require.NoError(t, err)

	// put an item to test deletion.
	want := &Item{ID: "test", Data: "i'm a teapot", Version: 3}
	_, err = client.PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String("Items"),
		Item:      Must(attributevalue.Marshal(want)).(*types.AttributeValueMemberM).Value,
	})
	require.NoError(t, err)

	// delete with ALL_OLD return values.
	deleteItemOutput, err := Delete(t.Context(), &Item{ID: "test", Version: 3}, func(opts *DeleteOptions) {
		opts.WithInputOptions(func(input *dynamodb.DeleteItemInput) {
			input.ReturnValues = types.ReturnValueAllOld
		})
	})
	require.NoError(t, err)

	// the old values must be identical to Item.
	got := &Item{}
	require.NoError(t, attributevalue.UnmarshalMap(deleteItemOutput.Attributes, got))
	require.Equal(t, want, got)
}

func TestMapper_Delete(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"Items"`
		Data    string `dynamodbav:"data"`
		Version int    `dynamodbav:"version,version"`
	}

	client := Setup(t, Item{})
	m, err := mapper.New[Item](func(m *mapper.Mapper[Item]) {
		m.Client = client
	})
	require.NoError(t, err)

	// deleting an item that doesn't exist do nothing.
	_, err = m.Delete(t.Context(), &Item{ID: "tes"})
	require.NoError(t, err)

	// put an item to test deletion.
	want := &Item{ID: "test", Data: "i'm a teapot", Version: 3}
	_, err = client.PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String("Items"),
		Item:      Must(attributevalue.Marshal(want)).(*types.AttributeValueMemberM).Value,
	})
	require.NoError(t, err)

	// delete with ALL_OLD return values.
	deleteItemOutput, err := m.Delete(t.Context(), &Item{ID: "test", Version: 3}, func(opts *DeleteOptions) {
		opts.WithInputOptions(func(input *dynamodb.DeleteItemInput) {
			input.ReturnValues = types.ReturnValueAllOld
		})
	})
	require.NoError(t, err)

	// the old values must be identical to Item.
	got := &Item{}
	require.NoError(t, attributevalue.UnmarshalMap(deleteItemOutput.Attributes, got))
	require.Equal(t, want, got)
}
