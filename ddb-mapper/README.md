# DynamoDB goodies

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/ddb-mapper.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb-mapper)

This module adds optimistic locking and auto-generated timestamps by modifying the expressions being created as part of
a DynamoDB service call. There are parallels between some types of this module and AWS SDK for Java 1.x which was what
I used for the longest time in a professional capacity:
* [Mapper](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb-mapper/mapper#Mapper) which provides a
type-safe way to interact with a specific DynamoDB table, is inspired by [DynamoDBMapper](https://docs.aws.amazon.com/AWSJavaSDK/latest/javadoc/com/amazonaws/services/dynamodbv2/datamodeling/DynamoDBMapper.html).
* [TableModel](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb-mapper/model#TableModel) parses
`dynamodbav` struct tags similar to [attributevalue.Marshal](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue#Marshal)
is inspired by [DynamoDBMapperTableModel](https://docs.aws.amazon.com/AWSJavaSDK/latest/javadoc/com/amazonaws/services/dynamodbv2/datamodeling/DynamoDBMapperTableModel.html).

## Usage

Get with:
```shell
go get github.com/nguyengg/go-aws-commons/ddb-mapper
```

### Type-agnostic package-level methods to interact with DynamoDB

Package `ddb` provides package-level methods to interact with DynamoDB. This is the most convenient way to use the
module.
```go
package main

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
)

func main() {
	// Shard, Id, Version, Created, and Modified have special semantics to the module.
	type Item struct {
		Shard    string    `dynamodbav:"shard,hashkey" tablename:"Items"`
		ID       string    `dynamodbav:"id,sortkey"`
		Data     string    `dynamodbav:"data"`
		Version  int       `dynamodbav:"version,version"`
		Created  time.Time `dynamodbav:"createdTime,createdTime"`
		Modified time.Time `dynamodbav:"modifiedTime,modifiedTime"`
	}

	// if testing against DynamoDB local, you can use this to create the table and wait for it to become active.
	err := ddb.CreateTable(context.Background(), Item{})

	// use ddb.Put to write new data.
	item := &Item{
		Shard: "us-east-1",
		ID:    uuid.NewString(),
		Data:  "hello, world",
		// with zero value for version, Put will add `attribute_not_exists(#pk)` condition to prevent overwriting.
		Version: 0,
		// with zero value for createdtime, Put will use time.Now.
		Created: time.Time{},
		// Put will always update modifiedtime to time.Now.
		Modified: time.Time{},
	}
	if _, err = ddb.Put(context.Background(), item); err == nil {
		// if ddb.Put succeeds, item.Version will be 1, and item.Created and item.Modified are no longer zero values.
		// this way, item reflects the latest contents of the item in DynamoDB until another write operation.
	}

	// subsequent ddb.Put and ddb.Update will use the version value to add optimistic locking with by adding a
	// `#version = :version` condition.

	// ddb.UpdateReturnAllNewValues automatically sets ReturnValues to ALL_NEW and will unmarshal the response back to
	// the item argument on success, so that similar to Put, item reflects the latest contents of the item in DynamoDB.
	if _, err = ddb.UpdateReturnAllNewValues(context.Background(), item, func(opts *config.UpdateOptions) {
		// DynamoDB requires at least once clause in the UpdateExpression so you should use UpdateOptions to add them.
		// ddb.Update already adds a clause to update the version.
		opts.Remove("data")
	}); err == nil {
	}
}

```

### Type-safe way to work with items in a specific DynamoDB table

Package [mapper](mapper) provides a more type-safe way to work with a specific DynamoDB table. First, in the package
that defines the struct that has `dynamodbav` tags
```go
package item

import (
	"time"

	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// Item models items inside "Items" table.
type Item struct {
	Shard    string    `dynamodbav:"shard,hashkey" tablename:"Items"`
	ID       string    `dynamodbav:"id,sortkey"`
	Data     string    `dynamodbav:"data"`
	Version  int       `dynamodbav:"version,version"`
	Created  time.Time `dynamodbav:"createdTime,createdTime"`
	Modified time.Time `dynamodbav:"modifiedTime,modifiedTime"`
}

// Mapper provides type-safe access to the table modeled by Item.
var Mapper *mapper.Mapper[Item]

func init() {
	var err error
	Mapper, err = mapper.NewMustHave[Item](model.AttributeModelTypeSortKey |
		model.AttributeModelTypeVersion |
		model.AttributeModelTypeCreatedTime |
		model.AttributeModelTypeModifiedTime)
	if err != nil {
		panic(err)
	}
}

```

Now you can use the global `Mapper` to work with items from *Items* table:
```go
package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper/app"
)

func main() {
	if _, err := item.Mapper.Put(context.Background(), &item.Item{
		Shard: "us-east-1",
		ID:    uuid.NewString(),
		Data:  "hello, world",
	}); err != nil {
		panic(err)
	}
}

```
