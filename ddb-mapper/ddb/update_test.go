package ddb_test

import (
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/ddb"
	ddbtypes "github.com/nguyengg/go-aws-commons/ddb-mapper/ddb/types"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/ddb-local-test"
	mappertypes "github.com/nguyengg/go-aws-commons/ddb-mapper/types"
	local "github.com/nguyengg/go-dynamodb-local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	type Item struct {
		ID           string             `dynamodbav:"id,hashkey" tableName:"Items"`
		Data         string             `dynamodbav:"data"`
		Version      int                `dynamodbav:"version,version"`
		CreatedTime  ddbtypes.UnixTime  `dynamodbav:"createdTime,createdTime"`
		ModifiedTime ddbtypes.UnixMilli `dynamodbav:"modifiedTime,modifiedTime"`
	}

	client := Setup(t, Item{})
	DefaultClientProvider = StaticClientProvider{Client: client}

	// first UpdateItem will create it with only Version, CreatedTime, and ModifiedTime filled out.
	_, err := Update(t.Context(), &Item{ID: "test"})
	require.NoError(t, err)
	avM := local.GetItem(t, client, "Items", "id", "test")
	assert.ElementsMatch(t, []string{"id", "version", "modifiedTime"}, slices.Collect(maps.Keys(avM)))
	item := &Item{}
	require.NoError(t, attributevalue.UnmarshalMap(avM, item))
	assert.Equal(t, 1, item.Version)

	// second UpdateItem, I'll manually update data and createdTime to old modifiedTime.
	oldModifiedTime := item.ModifiedTime
	_, err = Update(t.Context(), item, func(opts *mapper.UpdateOptions) {
		opts.
			Set("data", "i'm a teapot").
			Set("createdTime", ddbtypes.UnixTime(oldModifiedTime))
	})
	avM = local.GetItem(t, client, "Items", "id", "test")
	assert.ElementsMatch(t, []string{"id", "data", "version", "createdTime", "modifiedTime"}, slices.Collect(maps.Keys(avM)))
	item = &Item{}
	require.NoError(t, attributevalue.UnmarshalMap(avM, item))
	assert.Equal(t, 2, item.Version)
	assert.Equal(t, "i'm a teapot", item.Data)
	assert.True(t, item.CreatedTime.Equal(ddbtypes.UnixTime(oldModifiedTime)))
	assert.True(t, item.ModifiedTime.After(oldModifiedTime))
}

func TestUpdate_UpdateBehaviourDefault(t *testing.T) {
	type Item struct {
		ID           string             `dynamodbav:"id,hashkey" tableName:"Items"`
		Data         string             `dynamodbav:"data"`
		Tag          ddbtypes.StringSet `dynamodbav:"tags"`
		Version      int                `dynamodbav:"version,version"`
		CreatedTime  ddbtypes.UnixTime  `dynamodbav:"createdTime,createdTime"`
		ModifiedTime ddbtypes.UnixMilli `dynamodbav:"modifiedTime,modifiedTime"`
	}

	client := Setup(t, Item{})
	DefaultClientProvider = StaticClientProvider{Client: client}

	_, err := Update(t.Context(), &Item{
		ID:          "test",
		Data:        "hello, world!",
		Version:     0,
		CreatedTime: ddbtypes.UnixTime(time.Now()),
	}, func(opts *mapper.UpdateOptions) {
		opts.WithUpdateBehaviour(mappertypes.UpdateBehaviourDefault)
	})
	require.NoError(t, err)

	avM := local.GetItem(t, client, "Items", "id", "test")
	// UpdateBehaviourDefault cause tags to be omitted form the write because it's a zero value.
	assert.ElementsMatch(t, []string{"id", "data", "version", "createdTime", "modifiedTime"}, slices.Collect(maps.Keys(avM)))
}

func TestUpdate_UpdateBehaviourAsTagged(t *testing.T) {
	type Item struct {
		ID           string             `dynamodbav:"id,hashkey" tableName:"Items"`
		Data         string             `dynamodbav:"data"`
		Tag          ddbtypes.StringSet `dynamodbav:"tags"`
		Version      int                `dynamodbav:"version,version"`
		CreatedTime  ddbtypes.UnixTime  `dynamodbav:"createdTime,createdTime"`
		ModifiedTime ddbtypes.UnixMilli `dynamodbav:"modifiedTime,modifiedTime"`
	}

	client := Setup(t, Item{})
	DefaultClientProvider = StaticClientProvider{Client: client}

	_, err := Update(t.Context(), &Item{
		ID:          "test",
		Data:        "hello, world!",
		Version:     0,
		CreatedTime: ddbtypes.UnixTime(time.Now()),
	}, func(opts *mapper.UpdateOptions) {
		opts.WithUpdateBehaviour(mappertypes.UpdateBehaviourAsTagged)
	})
	require.NoError(t, err)

	avM := local.GetItem(t, client, "Items", "id", "test")
	// UpdateBehaviourAsTagged will cause "tags" to be written as NULL data type.
	assert.ElementsMatch(t, []string{"id", "data", "tags", "version", "createdTime", "modifiedTime"}, slices.Collect(maps.Keys(avM)))
	assert.IsType(t, &types.AttributeValueMemberNULL{}, avM["tags"])
}

func TestUpdateReturnAllNewValues(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"Items"`
		Data    string `dynamodbav:"data"`
		Version int    `dynamodbav:"version,version"`
	}

	client := Setup(t, Item{})
	DefaultClientProvider = StaticClientProvider{Client: client}

	// the first Update has non-empty data.
	_, err := Update(t.Context(), &Item{ID: "test"}, func(opts *mapper.UpdateOptions) {
		opts.Set("data", "hello, world!")
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"id", "data", "version"}, slices.Collect(maps.Keys(local.GetItem(t, client, "Items", "id", "test"))))

	// the second Update only changes the version, but because of ALL_NEW return values, item is updated with correct data.
	item := &Item{ID: "test", Version: 1}
	_, err = UpdateReturnAllNewValues(t.Context(), item)
	assert.Equal(t, &Item{
		ID:      "test",
		Data:    "hello, world!",
		Version: 2,
	}, item)
}
