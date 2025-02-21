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
		Id           string    `dynamodbav:",hashkey" tableName:""`
		Sort         string    `dynamodbav:",sortkey"`
		Version      int64     `dynamodbav:",version"`
		CreatedTime  time.Time `dynamodbav:",createdTime"`
		ModifiedTime time.Time `dynamodbav:",modifiedTime,unixtime"`
	}

	a, err := NewTableFromStruct(Test{})
	if err != nil {
		t.Errorf("NewTableFromStruct() error: %v", err)
	}

	b, err := NewTable(reflect.TypeFor[Test]())
	if err != nil {
		t.Errorf("NewTableFromType() error: %v", err)
	}

	assert.Equal(t, a, b)

	// can also parse from pointer value.
	c, err := NewTableFromStruct(&Test{})
	if err != nil {
		t.Errorf("NewTableFromStruct() error: %v", err)
	}

	assert.Equal(t, a, c)
}
