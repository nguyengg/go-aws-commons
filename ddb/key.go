package ddb

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// CreateKey creates the `map[string]types.AttributeValue` containing just the partition key for items of type T.
//
// The table name is also returned for convenience. Use this method if you need to extract the key to execute DynamoDB
// calls that have no out-of-the-box support in this package (e.g. Query, Scan). The method does not validate whether
// the type of partitionKey argument does not match or does not encode to same DynamoDB data type as the actual key
// modeled in struct type T. The method does return an error if struct T does not model a hashkey.
//
// [DefaultBuilder.Encoder] is used to produce the returned values.
func CreateKey[T interface{}](partitionKey interface{}) (key map[string]types.AttributeValue, tableName string, err error) {
	return createKey[T](partitionKey, nil)
}

// CreateCompositeKey creates the `map[string]types.AttributeValue` containing both the partition and sort keys for
// items of type T.
//
// The table name is also returned for convenience. Use this method if you need to extract the key to execute DynamoDB
// calls that have no out-of-the-box support in this package (e.g. Query, Scan). The method does not validate whether
// if the type of key arguments do not match or do not encode to same DynamoDB data type as the actual keys modeled in
// struct type T. The method does return an error if struct T does not model both a hashkey and sortkey.
//
// [DefaultBuilder.Encoder] is used to produce the returned values.
func CreateCompositeKey[T interface{}](partitionKey, sortKey interface{}) (key map[string]types.AttributeValue, tableName string, err error) {
	return createKey[T](partitionKey, sortKey)
}

// createKey allows sortKey to be nil.
func createKey[T interface{}](partitionKey, sortKey interface{}) (key map[string]types.AttributeValue, tableName string, err error) {
	t := reflect.TypeFor[T]()
	m, err := DefaultBuilder.loadOrParse(t)
	if err != nil {
		return nil, "", err
	}

	tableName = m.TableName
	key = make(map[string]types.AttributeValue)

	if k := m.HashKey; k == nil {
		return nil, "", fmt.Errorf(`no hashkey field in type "%s"`, t)
	} else if key[k.AttributeName], err = DefaultBuilder.Encoder.Encode(partitionKey); err != nil {
		return nil, "", err
	}

	if sortKey != nil {
		if k := m.SortKey; k == nil {
			return nil, "", fmt.Errorf(`no sortkey field in type "%s"`, t)
		} else if key[k.AttributeName], err = DefaultBuilder.Encoder.Encode(sortKey); err != nil {
			return nil, "", err
		}
	}

	return
}
