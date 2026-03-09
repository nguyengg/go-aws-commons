package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
	ddbtypes "github.com/nguyengg/go-aws-commons/ddb-mapper/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_updateVersion_int(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version int    `dynamodbav:"version,version"`
	}

	m, err := model.NewForType[Item]()
	require.NoError(t, err)

	item := &Item{}
	ptr := reflect.ValueOf(item)
	undo, _, err := updateVersion(m, ptr.Elem(), ptr, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, item.Version)

	undo()
	assert.Equal(t, 0, item.Version)
}

func Test_updateVersion_uint(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version uint32 `dynamodbav:"version,version"`
	}

	m, err := model.NewForType[Item]()
	require.NoError(t, err)

	item := &Item{Version: 67}
	ptr := reflect.ValueOf(item)
	undo, _, err := updateVersion(m, ptr.Elem(), ptr, nil)
	require.NoError(t, err)
	assert.Equal(t, uint32(68), item.Version)

	undo()
	assert.Equal(t, uint32(67), item.Version)
}

func Test_updateVersion_string(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version string `dynamodbav:"version,version"`
	}

	m, err := model.NewForType[Item]()
	require.NoError(t, err)

	item := &Item{Version: "hello, world!"}
	ptr := reflect.ValueOf(item)
	undo, _, err := updateVersion(m, ptr.Elem(), ptr, nil)
	require.NoError(t, err)
	_, err = uuid.Parse(item.Version)
	assert.NoError(t, err)

	undo()
	assert.Equal(t, "hello, world!", item.Version)
}

func Test_updateVersion_stringAlias(t *testing.T) {
	type myString string
	type Item struct {
		ID      string   `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version myString `dynamodbav:"version,version"`
	}

	m, err := model.NewForType[Item]()
	require.NoError(t, err)

	item := &Item{}
	ptr := reflect.ValueOf(item)
	undo, _, err := updateVersion(m, ptr.Elem(), ptr, nil)
	require.NoError(t, err)
	_, err = uuid.Parse(string(item.Version))
	assert.NoError(t, err)

	undo()
	assert.Equal(t, myString(""), item.Version)
}

func Test_updateVersion_stringWithCustomNextVersion(t *testing.T) {
	type Item struct {
		ID      string `dynamodbav:"id,hashkey" tableName:"my-table"`
		Version string `dynamodbav:"version,version"`
	}

	m, err := model.NewForType[Item]()
	require.NoError(t, err)

	item := &Item{Version: "hello, world!"}
	ptr := reflect.ValueOf(item)
	undo, _, err := updateVersion(m, ptr.Elem(), ptr, func(item any) {
		item.(*Item).Version = "i'm a teapot"
	})
	require.NoError(t, err)
	assert.Equal(t, "i'm a teapot", item.Version)

	undo()
	assert.Equal(t, "hello, world!", item.Version)
}

func Test_updateTimestamps(t *testing.T) {
	type Item struct {
		ID          string    `dynamodbav:"id,hashkey" tableName:"my-table"`
		CreatedTime time.Time `dynamodbav:"createdTime,createdTime"`
	}

	m, err := model.NewForType[Item]()
	require.NoError(t, err)

	item := &Item{}
	now := time.Now()
	undo := updateTimestamps(m, reflect.ValueOf(item).Elem(), now, true)
	require.NoError(t, err)
	assert.Equal(t, now, item.CreatedTime)
	undo()
	assert.Equal(t, time.Time{}, item.CreatedTime)

	// because CreatedTime is non-zero, it will be ignored.
	item.CreatedTime = now
	undo = updateTimestamps(m, reflect.ValueOf(item).Elem(), now, true)
	require.NoError(t, err)
	assert.Equal(t, now, item.CreatedTime)
	assert.Nil(t, undo)
}

func Test_updateTimestamps_CustomTime(t *testing.T) {
	type Item struct {
		ID           string             `dynamodbav:"id,hashkey" tableName:"my-table"`
		ModifiedTime ddbtypes.UnixMilli `dynamodbav:"modifiedTime,modifiedTime"`
	}

	m, err := model.NewForType[Item]()
	require.NoError(t, err)

	item := &Item{}
	now := time.Now()
	undo := updateTimestamps(m, reflect.ValueOf(item).Elem(), now, false)
	require.NoError(t, err)
	assert.Equal(t, ddbtypes.UnixMilli(now), item.ModifiedTime)
	undo()
	assert.Equal(t, ddbtypes.UnixMilli{}, item.ModifiedTime)
}
