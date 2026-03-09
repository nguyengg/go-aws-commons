package ddb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/model"
)

// EncodeKeys extracts and marshals the key attributes from the given item.
//
// The item argument must be a struct or struct pointer that is parseable by [model.NewForTypeOf]. The table name is also
// returned for convenience. Use this method if you need to extract the key to execute DynamoDB calls that have no
// out-of-the-box support in this package (e.g. Query, Scan).
func EncodeKeys(item any) (key map[string]types.AttributeValue, tableName string, err error) {
	m, err := model.NewForTypeOf(item)
	if err != nil {
		return key, "", err
	}

	tableName = m.TableName
	key, err = m.EncodeKeys(item)
	return
}
