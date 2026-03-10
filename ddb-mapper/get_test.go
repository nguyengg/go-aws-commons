package ddb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/ddb-local-test"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"Items"`
		Data    string `dynamodbav:"data"`
		Version int    `dynamodbav:"version,version"`
	}

	client := Setup(t, Item{})
	ddb.DefaultClientProvider = &ddb.StaticClientProvider{Client: client}

	want := &Item{ID: "test", Data: "i'm a teapot", Version: 3}
	_, err := client.PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String("Items"),
		Item:      MustMarshalToM(t, want),
	})
	require.NoError(t, err)

	got := &Item{ID: "test"}
	_, err = ddb.Get(t.Context(), got)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

// TestMapper_Get is exactly as TestGet but uses typed Mapper.
func TestMapper_Get(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"Items"`
		Data    string `dynamodbav:"data"`
		Version int    `dynamodbav:"version,version"`
	}

	client := Setup(t, Item{})
	m, err := mapper.New[Item](func(cfg *config.Config) {
		cfg.Client = client
	})
	require.NoError(t, err)

	want := &Item{ID: "test", Data: "i'm a teapot", Version: 3}
	_, err = client.PutItem(t.Context(), &dynamodb.PutItemInput{
		TableName: aws.String("Items"),
		Item:      MustMarshalToM(t, want),
	})
	require.NoError(t, err)

	got := &Item{ID: "test"}
	_, err = m.Get(t.Context(), got)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
