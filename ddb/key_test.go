package ddb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateKey_S(t *testing.T) {
	type Test struct {
		Id string `dynamodbav:"hash,hashkey" tableName:"my-table"`
	}

	key, tableName, err := CreateKey[Test]("my-id")
	assert.NoErrorf(t, err, "CreateKey[Test](...) error = %v", err)
	assert.Equal(t, "my-table", tableName)
	assert.Equal(t, map[string]types.AttributeValue{"hash": &types.AttributeValueMemberS{Value: "my-id"}}, key)

	_, _, err = CreateKey[Test](1234)
	assert.Error(t, err)
}

func TestCreateCompositeKey_N(t *testing.T) {
	type Test struct {
		Id   string `dynamodbav:"hash,hashkey" tableName:"my-table"`
		Sort int64  `dynamodbav:"sort,sortkey"`
	}

	key, tableName, err := CreateCompositeKey[Test]("my-id", 3)
	assert.NoErrorf(t, err, "CreateCompositeKey[Test](...) error = %v", err)
	assert.Equal(t, "my-table", tableName)
	assert.Equal(t, map[string]types.AttributeValue{
		"hash": &types.AttributeValueMemberS{Value: "my-id"},
		"sort": &types.AttributeValueMemberN{Value: "3"},
	}, key)

	_, _, err = CreateCompositeKey[Test](1234, "my-id")
	assert.Error(t, err)
}
