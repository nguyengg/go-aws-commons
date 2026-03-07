package untyped

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	ddbtypes "github.com/nguyengg/go-aws-commons/ddb-mapper/ddb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapper_keys_WithHashKeyOnly(t *testing.T) {
	type Item struct {
		ID   string `dynamodbav:"id,hashkey" tableName:"Items"`
		Data string `dynamodbav:"data"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	key, err := m.keys(DefaultContext(), Item{
		ID:   "my-id",
		Data: "data", // Data must not show up in key.
	})
	require.NoError(t, err)
	require.Equal(t, map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "my-id"}}, key)
}

func TestMapperr_keys_WithCompositeKey(t *testing.T) {
	type Item struct {
		ID   string `dynamodbav:"id,hashkey" tableName:"Items"`
		Sort int64  `dynamodbav:"sort,rangekey"`
		Data string `dynamodbav:"data"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	key, err := m.keys(DefaultContext(), Item{
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

func TestMapper_updateVersion_int(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version int    `dynamodbav:"version,version"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	item := &Item{}
	ptr := reflect.ValueOf(item)
	undo, _ := m.updateVersion(DefaultContext(), ptr.Elem(), ptr)
	assert.Equal(t, 1, item.Version)

	undo()
	assert.Equal(t, 0, item.Version)
}

func TestMapper_updateVersion_uint(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version uint32 `dynamodbav:"version,version"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	item := &Item{Version: 67}
	ptr := reflect.ValueOf(item)
	undo, _ := m.updateVersion(DefaultContext(), ptr.Elem(), ptr)
	assert.Equal(t, uint32(68), item.Version)

	undo()
	assert.Equal(t, uint32(67), item.Version)
}

func TestMapper_updateVersion_string(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version string `dynamodbav:"version,version"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	item := &Item{Version: "hello, world!"}
	ptr := reflect.ValueOf(item)
	undo, _ := m.updateVersion(DefaultContext(), ptr.Elem(), ptr)
	_, err = uuid.Parse(item.Version)
	assert.NoError(t, err)

	undo()
	assert.Equal(t, "hello, world!", item.Version)
}

func TestMapper_updateVersion_stringAlias(t *testing.T) {
	type myString string
	type Item struct {
		ID      string   `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version myString `dynamodbav:"version,version"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	item := &Item{}
	ptr := reflect.ValueOf(item)
	undo, _ := m.updateVersion(DefaultContext(), ptr.Elem(), ptr)
	_, err = uuid.Parse(string(item.Version))
	assert.NoError(t, err)

	undo()
	assert.Equal(t, myString(""), item.Version)
}

func TestMapper_updateVersion_stringWithCustomNextVersion(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version string `dynamodbav:"version,version"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	item := &Item{Version: "hello, world!"}
	ptr := reflect.ValueOf(item)
	undo, _ := m.updateVersion(&Context{Options: Options{VersionUpdater: func(item any) {
		item.(*Item).Version = "i'm a teapot"
	}}}, ptr.Elem(), ptr)
	assert.Equal(t, "i'm a teapot", item.Version)

	undo()
	assert.Equal(t, "hello, world!", item.Version)
}

func TestMapper_updateTimestamps(t *testing.T) {
	type Item struct {
		ID          string    `dynamodbav:"id,hashkey" tableName:"my-table"`
		CreatedTime time.Time `dynamodbav:"createdTime,createdTime"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	item := &Item{}
	now := time.Now()
	undo := m.updateTimestamps(reflect.ValueOf(item).Elem(), now, true)
	require.NoError(t, err)
	assert.Equal(t, now, item.CreatedTime)
	undo()
	assert.Equal(t, time.Time{}, item.CreatedTime)

	// because CreatedTime is non-zero, it will be ignored.
	item.CreatedTime = now
	undo = m.updateTimestamps(reflect.ValueOf(item).Elem(), now, true)
	require.NoError(t, err)
	assert.Equal(t, now, item.CreatedTime)
	assert.Nil(t, undo)
}

func TestMapper_updateTimestamps_CustomTime(t *testing.T) {
	type Item struct {
		ID           string             `dynamodbav:"id,hashkey" tableName:"my-table"`
		ModifiedTime ddbtypes.UnixMilli `dynamodbav:"modifiedTime,modifiedTime"`
	}

	m, err := NewFromType(reflect.TypeFor[Item]())
	require.NoError(t, err)

	item := &Item{}
	now := time.Now()
	undo := m.updateTimestamps(reflect.ValueOf(item).Elem(), now, false)
	require.NoError(t, err)
	assert.Equal(t, ddbtypes.UnixMilli(now), item.ModifiedTime)
	undo()
	assert.Equal(t, ddbtypes.UnixMilli{}, item.ModifiedTime)
}
