package ddb_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	. "github.com/nguyengg/go-aws-commons/ddb-mapper/internal/ddb-local-test"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	ddbtypes "github.com/nguyengg/go-aws-commons/ddb-mapper/types"
	local "github.com/nguyengg/go-dynamodb-local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Put(t *testing.T) {
	type Item struct {
		ID           string             `dynamodbav:"id,hashkey" tableName:"Items"`
		Data         string             `dynamodbav:"data"`
		Version      int                `dynamodbav:"version,version"`
		CreatedTime  time.Time          `dynamodbav:"createdTime,createdTime,unixtime"`
		ModifiedTime ddbtypes.UnixMilli `dynamodbav:"modifiedTime,modifiedTime"`
	}

	client := Setup(t, Item{})
	config.DefaultClientProvider = &config.StaticClientProvider{Client: client}

	item := &Item{ID: "test", Data: "my-data"}
	_, err := ddb.Put(t.Context(), item)
	require.NoError(t, err)

	// createdTime has been truncated to epoch second to match what would be decoded from ddb due to unixtime tag.
	createdTime := item.CreatedTime
	modifiedTime := item.ModifiedTime
	assert.NotEqual(t, createdTime, modifiedTime)
	assert.Equal(t, createdTime.Unix(), time.Time(modifiedTime).Unix())

	// initial Put will get version == 1.
	RequireHasAttributes(t,
		map[string]any{
			"id":      "test",
			"version": 1,
			"data":    "my-data",
			"createdTime": AttributeAssertFn(func(t *testing.T, name string, value types.AttributeValue) {
				// createdTime must be unixtime, and must be the same value in item.
				// assert.Equal won't work here because createdTime was set from time.Now which is not truncated
				assert.True(t, createdTime.Unix() == AsUnixTime(t, value).Unix())
			}),
			"modifiedTime": AttributeAssertFn(func(t *testing.T, name string, value types.AttributeValue) {
				// modifiedTime must be UnixMilli.
				// assert.Equal won't work here because modifiedTime was set from time.Now which is not truncated.
				// TODO see if we can fix this by decoding the map[string]AttributeValue back.
				assert.True(t, modifiedTime.Equal(As[ddbtypes.UnixMilli](t, value)))
			}),
		},
		local.GetItem(t, client, "Items", "id", "test"))

	// second Put will increase version to 2, and only update modified timestamp.
	item.Data = "new-data"
	_, err = ddb.Put(t.Context(), item)
	require.NoError(t, err)
	assert.Equal(t, item.CreatedTime, createdTime)
	assert.True(t, item.ModifiedTime.After(modifiedTime))
	RequireHasAttributes(t,
		map[string]any{
			"id":      "test",
			"version": 2,
			"data":    "new-data",
		},
		local.GetItem(t, client, "Items", "id", "test"))

	// now if we change the version to 3, the PutItem will fail due to ConditionalCheckFailedException.
	item.Version = 3
	_, err = ddb.Put(t.Context(), item)
	var ccfe *types.ConditionalCheckFailedException
	require.ErrorAs(t, err, &ccfe)
}

// TestMapper_Put is exactly as Test_Put but uses typed Mapper.
func TestMapper_Put(t *testing.T) {
	type Item struct {
		ID           string             `dynamodbav:"id,hashkey" tableName:"Items"`
		Data         string             `dynamodbav:"data"`
		Version      int                `dynamodbav:"version,version"`
		CreatedTime  time.Time          `dynamodbav:"createdTime,createdTime,unixtime"`
		ModifiedTime ddbtypes.UnixMilli `dynamodbav:"modifiedTime,modifiedTime"`
	}

	client := Setup(t, Item{})
	m, err := mapper.New[Item](func(cfg *config.Config) {
		cfg.Client = client
	})
	require.NoError(t, err)

	item := &Item{ID: "test", Data: "my-data"}
	_, err = m.Put(t.Context(), item)
	require.NoError(t, err)

	// createdTime has been truncated to epoch second to match what would be decoded from ddb due to unixtime tag.
	createdTime := item.CreatedTime
	modifiedTime := item.ModifiedTime
	assert.NotEqual(t, createdTime, modifiedTime)
	assert.Equal(t, createdTime.Unix(), time.Time(modifiedTime).Unix())

	// initial Put will get version == 1.
	RequireHasAttributes(t,
		map[string]any{
			"id":      "test",
			"version": 1,
			"data":    "my-data",
			"createdTime": AttributeAssertFn(func(t *testing.T, name string, value types.AttributeValue) {
				// createdTime must be unixtime, and must be the same value in item.
				// assert.Equal won't work here because createdTime was set from time.Now which is not truncated
				assert.True(t, createdTime.Unix() == AsUnixTime(t, value).Unix())
			}),
			"modifiedTime": AttributeAssertFn(func(t *testing.T, name string, value types.AttributeValue) {
				// modifiedTime must be UnixMilli.
				// assert.Equal won't work here because modifiedTime was set from time.Now which is not truncated.
				// TODO see if we can fix this by decoding the map[string]AttributeValue back.
				assert.True(t, modifiedTime.Equal(As[ddbtypes.UnixMilli](t, value)))
			}),
		},
		local.GetItem(t, client, "Items", "id", "test"))

	// second Put will increase version to 2, and only update modified timestamp.
	item.Data = "new-data"
	_, err = m.Put(t.Context(), item)
	require.NoError(t, err)
	assert.Equal(t, item.CreatedTime, createdTime)
	assert.True(t, item.ModifiedTime.After(modifiedTime))
	RequireHasAttributes(t,
		map[string]any{
			"id":      "test",
			"version": 2,
			"data":    "new-data",
		},
		local.GetItem(t, client, "Items", "id", "test"))

	// now if we change the version to 3, the PutItem will fail due to ConditionalCheckFailedException.
	item.Version = 3
	_, err = m.Put(t.Context(), item)
	var ccfe *types.ConditionalCheckFailedException
	require.ErrorAs(t, err, &ccfe)
}
