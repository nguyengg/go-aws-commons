package ddbfns

import (
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_PutNewVersion(t *testing.T) {
	type Test struct {
		Id           string    `dynamodbav:"id,hashkey" tableName:""`
		Sort         string    `dynamodbav:"sort,sortkey"`
		Version      int64     `dynamodbav:"version,version"`
		CreatedTime  time.Time `dynamodbav:"createdTime,createdTime"`
		ModifiedTime time.Time `dynamodbav:"modifiedTime,modifiedTime,unixtime"`
	}
	input := Test{
		Id:   "hello",
		Sort: "world",
		// Version will be incremented by 1.
		Version: 0,
		// These timestamps will not be changed.
		CreatedTime:  testTime,
		ModifiedTime: testTime,
	}

	// this is to make sure the input item is not mutated.
	before := MustToJSON(input)

	got, err := CreatePutItem(input)
	assert.NoErrorf(t, err, "CreatePutItem() err = %v", err)
	assert.JSONEq(t, before, MustToJSON(input))
	assert.Equal(t, "attribute_not_exists (#0)", *got.ConditionExpression)
	assert.Equal(t, map[string]string{"#0": "id"}, got.ExpressionAttributeNames)
	assert.Equal(t, "2006-01-02T15:04:05Z", got.Item["createdTime"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "1136214245", got.Item["modifiedTime"].(*types.AttributeValueMemberN).Value)
}

func TestBuilder_PutNoVersion(t *testing.T) {
	type Test struct {
		Id          string    `dynamodbav:"id,hashkey" tableName:""`
		CreatedTime time.Time `dynamodbav:"createdTime,createdTime,unixtime"`
	}
	input := Test{
		Id: "hello",
		// CreatedTime will be updated to some time way after testTime since it is the zero value.
		CreatedTime: time.Time{},
	}

	// this is to make sure the input item is not mutated.
	before := MustToJSON(input)

	got, err := CreatePutItem(input)
	assert.NoErrorf(t, err, "CreatePutItem() err = %v", err)
	assert.JSONEq(t, before, MustToJSON(input))

	// the output's createdTime value should be after now.
	v := got.Item["createdTime"].(*types.AttributeValueMemberN).Value
	epochSecond, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		t.Errorf("Parse createdTime (%v) error: %v", v, err)
	}
	createdTime := time.Unix(epochSecond, 0)
	assert.Truef(t, createdTime.After(testTime), "Updated createdTime (%s) is not after testTime (%s)", createdTime, testTime)
}

func TestBuilder_PutIncrementVersion(t *testing.T) {
	type Test struct {
		Id      string `dynamodbav:"id,hashkey" tableName:""`
		Version int64  `dynamodbav:"version,version"`
	}
	input := Test{
		Id: "hello",
		// Version will be incremented by 1.
		Version: 3,
	}

	// this is to make sure the input item is not mutated.
	before := MustToJSON(input)

	// use pointer to input here to test pointer case as well.
	got, err := CreatePutItem(&input)
	assert.NoErrorf(t, err, "CreatePutItem() err = %v", err)
	assert.JSONEq(t, before, MustToJSON(input))
	assert.Equal(t, "#0 = :0", *got.ConditionExpression)
	assert.Equal(t, map[string]string{"#0": "version"}, got.ExpressionAttributeNames)
	assert.Equal(t, map[string]types.AttributeValue{":0": &types.AttributeValueMemberN{Value: "3"}}, got.ExpressionAttributeValues)
	assert.Equal(t, map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "hello"}, "version": &types.AttributeValueMemberN{Value: "4"}}, got.Item)
}
