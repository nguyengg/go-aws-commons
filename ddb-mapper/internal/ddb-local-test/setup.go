package localtest

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	mapperddb "github.com/nguyengg/go-aws-commons/ddb-mapper/ddb"
	local "github.com/nguyengg/go-dynamodb-local"
	"github.com/stretchr/testify/require"
)

// Setup starts the DynamoDB local instance and create the given tables.
func Setup(t *testing.T, items ...any) *dynamodb.Client {
	client := local.DefaultSkippable(t)

	mapperddb.DefaultClientProvider = mapperddb.StaticClientProvider{Client: client}

	for _, item := range items {
		require.NoErrorf(t, mapperddb.CreateTable(t.Context(), item), "create table for type %T error", item)
	}

	return client
}
