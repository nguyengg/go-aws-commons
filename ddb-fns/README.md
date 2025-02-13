# ddbfns

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-ddb-fns.svg)](https://pkg.go.dev/github.com/nguyengg/go-ddb-fns)

Inspired by date-fns but for Go AWS SDK v2, this package adds optimistic locking and auto-generated timestamps by 
modifying the expressions being created as part of a DynamoDB service call.

## Usage

Get with:
```shell
go get "github.com/nguyengg/go-ddb-fns"
```

Example usage (name of the module is `ddbfns`)

```go
package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	ddbfns "github.com/nguyengg/go-ddb-fns"
)

type Item struct {
	Id           string    `dynamodbav:"id,hashkey" tableName:"my-table"`
	Sort         string    `dynamodbav:"sort,sortkey"`
	Version      int64     `dynamodbav:"version,version"`
	CreatedTime  time.Time `dynamodbav:"createdTime,createdTime,unixtime"`
	ModifiedTime time.Time `dynamodbav:"modifiedTime,modifiedTime,unixtime"`
}

func main() {
	ctx := context.TODO()
	cfg, _ := config.LoadDefaultConfig(ctx)
	client := dynamodb.NewFromConfig(cfg)

	// ddbfns.Put and ddbfns.Update is smart enough to add condition expression for me.
	item := Item{
		Id:   "hello",
		Sort: "world",
		// Since version has zero value, ddbfns.Put and ddbfns.Put will add a condition expression that's basically
		// `attribute_not_exists (id)`, and increment the version's value in the request for me.
		Version: 0,
		// Since these timestamps have zero value, they will be updated to the same time.Now in the request for me.
		CreatedTime:  time.Time{},
		ModifiedTime: time.Time{},
	}

	// my original item is never mutated by ddbfns.
	// Instead, putItemRequest.Item is the one that is modified, along with its ConditionExpression,
	// ExpressionAttributeNames, and ExpressionAttributeValues.
	putItemInput, _ := ddbfns.Put(item)
	_, _ = client.PutItem(ctx, putItemInput)

	// If the version is not at zero value, ddbfns.Put and ddbfns.Update knows to add `#version = :old_value` instead.
	item = Item{
		Id:   "hello",
		Sort: "world",
		// Since version has non-zero value, ddbfns.Put and ddbfns.Put will add `#version = 3` instead, and increment
		// the version's value in the request for me.
		Version: 3,
		// In ddbfns.Update, only ModifiedTime is updated with a `SET #modifiedTime = :now`.
		ModifiedTime: time.Time{},
	}

	// Update requires me to specify at least one update expression. Here's an example of how to return updated values
	// as well.
	updateItemInput, _ := ddbfns.Update(item, func(opts *ddbfns.UpdateOpts) {
		opts.Set("anotherField", "notes")
		opts.ReturnValues = types.ReturnValueAllNew
	})
	_, _ = client.UpdateItem(ctx, updateItemInput)

	// ddbfns.Delete will only use the version attribute, and it does not care if the attribute has zero value or not
	// (i.e. you can't attempt to delete an item that doesn't exist).
	deleteItemInput, _ := ddbfns.Delete(Item{
		Id: "hello",
		// Even if the version's value was 0, `SET #version = :old_value` is used regardless.
		Version: 3,
	})
	_, _ = client.DeleteItem(ctx, deleteItemInput)

	// all methods above come with a DoXyz version that executes the request for you.
	// for example, DoGet is used with Decode here in order to perform unmarshalling of the response.
	key := Item{Id: "hello", Sort: "world"}
	if getItemOutput, _ := ddbfns.DoGet(ctx, client, key, func(opts *ddbfns.GetOpts) {
		// must pass a pointer to a struct here.
		// if it's not a pointer, attributevalue.UnmarshalMap will fail.
		opts.Decode(&item)
	}); len(getItemOutput.Item) == 0 {
		// item with key above doesn't exist.
	}

	// similarly, you can DoUpdate and return new values like this.
	_, _ = ddbfns.DoUpdate(ctx, client, key, func(opts *ddbfns.UpdateOpts) {
		opts.
			Decode(&item).
			Set("notes", "hello, world!").
			WithReturnValues(types.ReturnValueAllNew).
			WithReturnValuesOnConditionCheckFailure(types.ReturnValuesOnConditionCheckFailureAllOld)
	})
}

```
