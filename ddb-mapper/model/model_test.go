package model

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
)

func TestTableModel_EncodeKeys_WithHashKeyOnly(t *testing.T) {
	type Item struct {
		ID   string `dynamodbav:"id,hashkey" tableName:"Items"`
		Data string `dynamodbav:"data"`
	}

	m, err := NewForType[Item]()
	require.NoError(t, err)

	key, err := m.EncodeKeys(Item{
		ID:   "my-id",
		Data: "data", // Data must not show up in key.
	})
	require.NoError(t, err)
	require.Equal(t, map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "my-id"}}, key)
}

func TestTableModel_EncodeKeys_WithCompositeKey(t *testing.T) {
	type Item struct {
		ID   string `dynamodbav:"id,hashkey" tableName:"Items"`
		Sort int64  `dynamodbav:"sort,rangekey"`
		Data string `dynamodbav:"data"`
	}

	m, err := NewForType[*Item]()
	require.NoError(t, err)

	key, err := m.EncodeKeys(Item{
		ID:   "my-id",
		Sort: 7,
		Data: "data", // Data must not show up in key.
	})
	require.NoError(t, err)
	require.Equal(t, map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: "my-id"},
		"sort": &types.AttributeValueMemberN{Value: "7"},
	}, key)
}
