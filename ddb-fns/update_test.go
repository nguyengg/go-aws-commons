package ddbfns

import (
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_UpdateNewVersion(t *testing.T) {
	type Test struct {
		Id           string    `dynamodbav:"id,hashkey" tableName:""`
		Notes        string    `dynamodbav:"notes"`
		Version      int64     `dynamodbav:"version,version"`
		CreatedTime  time.Time `dynamodbav:"createdTime,createdTime"`
		ModifiedTime time.Time `dynamodbav:"modifiedTime,modifiedTime,unixtime"`
	}
	input := Test{
		Id:    "hello",
		Notes: "",
		// Version will be incremented by 1.
		Version: 0,
		// CreatedTime will not be modified so it can be time.Now.
		CreatedTime: time.Now(),
		// ModifiedTime will be updated to some time way after testTime.
		ModifiedTime: testTime,
	}

	// this is to make sure the input item is not mutated.
	before := MustToJSON(input)

	got, err := CreateUpdateItem(input, func(opts *UpdateOptions) {
		opts.Set("notes", "world!")
	})
	assert.NoErrorf(t, err, "CreateUpdateItem() err = %v", err)
	assert.JSONEq(t, before, MustToJSON(input))
	assert.Equal(t, "attribute_not_exists (#0)", *got.ConditionExpression)
	assert.Equal(t, "SET #1 = :0, #2 = :1, #3 = :2\n", *got.UpdateExpression)
	assert.Equal(t, map[string]string{"#0": "id", "#1": "notes", "#2": "version", "#3": "modifiedTime"}, got.ExpressionAttributeNames)
	assert.Equal(t, map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "hello"}}, got.Key)

	// should have :0 = "world!", :1 = 1, and :2 = new timestamp.
	assert.Equal(t, 3, len(got.ExpressionAttributeValues))
	assert.Equal(t, &types.AttributeValueMemberS{Value: "world!"}, got.ExpressionAttributeValues[":0"])
	assert.Equal(t, &types.AttributeValueMemberN{Value: "1"}, got.ExpressionAttributeValues[":1"])

	v := got.ExpressionAttributeValues[":2"].(*types.AttributeValueMemberN).Value
	epochSecond, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		t.Errorf("Parse :2 (%v) error: %v", v, err)
	}
	modifiedTime := time.Unix(epochSecond, 0)
	assert.Truef(t, modifiedTime.After(testTime), "Updated modifiedTime (%s) is not after testTime (%s)", modifiedTime, testTime)
}

func TestBuilder_UpdateNoVersion(t *testing.T) {
	type Test struct {
		Id string `dynamodbav:"id,hashkey" tableName:""`
	}
	input := Test{
		Id: "hello",
	}

	// this is to make sure the input item is not mutated.
	before := MustToJSON(input)

	got, err := CreateUpdateItem(input, func(opts *UpdateOptions) {
		opts.Set("notes", "world!")
	})
	assert.NoErrorf(t, err, "CreateUpdateItem() err = %v", err)
	assert.JSONEq(t, before, MustToJSON(input))
	assert.Nil(t, got.ConditionExpression)
	assert.Equal(t, "SET #0 = :0\n", *got.UpdateExpression)
	assert.Equal(t, map[string]string{"#0": "notes"}, got.ExpressionAttributeNames)
	assert.Equal(t, map[string]types.AttributeValue{":0": &types.AttributeValueMemberS{Value: "world!"}}, got.ExpressionAttributeValues)
}

func TestBuilder_UpdateIncrementVersion(t *testing.T) {
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
	got, err := CreateUpdateItem(&input, func(opts *UpdateOptions) {
		opts.Set("notes", "world!")
	})
	assert.NoErrorf(t, err, "CreateUpdateItem() err = %v", err)
	assert.JSONEq(t, before, MustToJSON(input))
	assert.Equal(t, "#0 = :0", *got.ConditionExpression)
	assert.Equal(t, "ADD #0 :1\nSET #1 = :2\n", *got.UpdateExpression)
	assert.Equal(t, map[string]string{"#0": "version", "#1": "notes"}, got.ExpressionAttributeNames)
	assert.Equal(t, map[string]types.AttributeValue{":0": &types.AttributeValueMemberN{Value: "3"}, ":1": &types.AttributeValueMemberN{Value: "1"}, ":2": &types.AttributeValueMemberS{Value: "world!"}}, got.ExpressionAttributeValues)
}
