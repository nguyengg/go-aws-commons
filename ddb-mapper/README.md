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

### Use `dynamodbav` struct tags to model your table

```go
type Item struct {
	PK	 string    `dynamodbav:"pk,hashkey" tablename:"Items"`
	SK       string    `dynamodbav:"sk,sortkey"`
	Data     string    `dynamodbav:"data"`
	Version  int       `dynamodbav:"version,version"`
	Created  time.Time `dynamodbav:"created,createdtime"`
	Modified time.Time `dynamodbav:"modified,modifiedtime"`
}
```

### Type-agnostic package-level methods to interact with DynamoDB

Use package-level methods (from `ddb` package) to interact with DynamoDB. This is the most convenient way to use the
module.

```go
// if testing against DynamoDB local, you can use this to create the table and wait for it to become active.
err := ddb.CreateTable(context.Background(), Item{})

// use Put to write new data.
item := &Item{
	PK:   "us-east-1",
	SK:   uuid.NewString(),
	Data: "hello, world",
	// with zero value for version, Put will add `attribute_not_exists(#pk)` condition to prevent overwriting.
	Version: 0,
}
if _, err = ddb.Put(context.Background(), item); err == nil {
	// if ddb.Put succeeds, item.Version will be 1, and item.Created and item.Modified are no longer zero values.
	// this way, item reflects the latest contents of the item in DynamoDB until another write operation.
}

// subsequent Put and Update will use the version value to apply optimistic locking by adding a
// `#version = :version` condition. item.Modified is also updated to time.Now for auto-generated timestamps.

if _, err = ddb.UpdateReturnAllNewValues(context.Background(), item, func(opts *config.UpdateOptions) {
	// DynamoDB requires at least once clause in the UpdateExpression so you should use UpdateOptions to add them.
	// It's optional though since the version and modified time already have their SET clauses added.
	opts.Remove("data")
}); err == nil {
	// ddb.UpdateReturnAllNewValues automatically sets ReturnValues to ALL_NEW and will unmarshal the response so
	// that, similar to Put, item reflects the latest contents of the item in DynamoDB.
}
```

### Type-safe way to work with items in a specific DynamoDB table

Alternatively, package [mapper](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb-mapper/mapper) provides a more
type-safe way to work with a specific DynamoDB table. First, in the package that defines the struct that has
`dynamodbav` tags, add a `Mapper` package variable:
```go
package item

import (
	"github.com/nguyengg/go-aws-commons/ddb-mapper/mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

type Item struct {
	// no changes.
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
_, err := item.Mapper.Put(context.Background(), &item.Item{
	PK:   "us-east-1",
	SK:   uuid.NewString(),
	Data: "hello, world",
})
```
