package internal

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

// both ParseFromStruct and ParseFromType should return exact same values.
func TestParse(t *testing.T) {
	type Test struct {
		Id           string    `dynamodbav:",hashkey" tableName:""`
		Sort         string    `dynamodbav:",sortkey"`
		Version      int64     `dynamodbav:",version"`
		CreatedTime  time.Time `dynamodbav:",createdTime"`
		ModifiedTime time.Time `dynamodbav:",modifiedTime,unixtime"`
	}

	a, err := ParseFromStruct(Test{})
	if err != nil {
		t.Errorf("ParseFromStruct() error: %v", err)
	}

	b, err := ParseFromType(reflect.TypeFor[Test]())
	if err != nil {
		t.Errorf("ParseFromType() error: %v", err)
	}

	assert.Equal(t, a, b)

	// can also parse from pointer value.
	c, err := ParseFromStruct(&Test{})
	if err != nil {
		t.Errorf("ParseFromStruct() error: %v", err)
	}

	assert.Equal(t, a, c)
}
