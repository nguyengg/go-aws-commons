package ddb

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// both NewTableFromStruct and NewTable should return exact same values.
func TestNewTable(t *testing.T) {
	type Test struct {
		Id           string    `dynamodbav:"id,hashkey" tableName:""`
		Sort         string    `dynamodbav:"-,sortkey"`
		Version      int64     `dynamodbav:"-,version"`
		CreatedTime  time.Time `dynamodbav:"-,createdTime"`
		ModifiedTime time.Time `dynamodbav:"-,modifiedTime,unixtime"`
	}

	a, err := NewTableFromStruct(Test{})
	assert.NoErrorf(t, err, "NewTableFromStruct() error: %v", err)

	b, err := NewTable(reflect.TypeFor[Test]())
	assert.NoErrorf(t, err, "NewTable() error: %v", err)

	assert.Equal(t, a, b)

	// can also parse from pointer value.
	c, err := NewTableFromStruct(&Test{})
	assert.NoErrorf(t, err, "NewTableFromStruct() error: %v", err)
	_, err = NewTable(reflect.TypeFor[*Test]())
	assert.NoErrorf(t, err, "NewTable() error: %v", err)

	assert.Equal(t, a, c)

	// Get can be called on both struct and pointer.
	assert.Equal(t, "hello", a.MustGet(&Test{Id: "hello"}, "id"))
	assert.Equal(t, "hello", a.MustGet(Test{Id: "hello"}, "id"))
}
