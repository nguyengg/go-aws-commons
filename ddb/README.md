# DynamoDB goodies

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/ddb.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb)

This package adds optimistic locking and auto-generated timestamps by modifying the expressions being created as part of
a DynamoDB service call.

## Usage

Get with:

```shell
go get github.com/nguyengg/go-aws-commons/ddb
```

```go
package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb"
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
	// Instead, the map[string]AttributeValue sent to DynamoDB is the one that is modified along with its
	// the request's ConditionExpression, ExpressionAttributeNames, and ExpressionAttributeValues.
	_, _ = ddb.Put(ctx, item)

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
	_, _ = ddb.Update(ctx, Item{Id: "hello", Sort: " world", Version: 3}, func(options *ddb.UpdateOptions) {
		options.
			Set("anotherField", "notes").
			// by passing a struct pointer to WithReturnValues, ddbfns will unmarshall the response to
			// this struct so that item will have updated attribute values.
			WithReturnValues(types.ReturnValueAllNew, &item)
	})

	// ddbfns.Delete will only use the version attribute, and it does not care if the attribute has zero value or not
	// (i.e. you can't attempt to delete an item that doesn't exist).
	_, _ = ddb.Delete(ctx, Item{
		Id: "hello",
		// Even if the version's value was 0, `SET #version = :old_value` is used regardless.
		Version: 3,
	})

	// ddbfns.Get accepts a key which can be struct or struct pointer, and an optional struct pointer argument to
	// unmarshal the response of the GetItem request.
	item = Item{}
	_, _ = ddb.Get(ctx, Item{Id: "hello", Sort: "world"}, &item)
}

```
