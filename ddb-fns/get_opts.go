package ddbfns

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// GetOptions customises Get and CreateGetItem via chainable methods.
type GetOptions struct {
	// TableName modifies the [dynamodb.GetItemInput.TableName]
	TableName *string
	// ConsistentRead modifies the [dynamodb.GetItemInput.ConsistentRead]
	ConsistentRead *bool
	// ReturnConsumedCapacity modifies the [dynamodb.GetItemInput.ReturnConsumedCapacity]
	ReturnConsumedCapacity types.ReturnConsumedCapacity

	// used by CreateGetItem.
	names []string

	// used by Manager.Get.
	optFns []func(*dynamodb.Options)
}

// WithTableName overrides [GetOptions.TableName].
func (o *GetOptions) WithTableName(tableName string) *GetOptions {
	o.TableName = &tableName
	return o
}

// WithProjectionExpression replaces the current projection expression with the given names.
func (o *GetOptions) WithProjectionExpression(name string, names ...string) *GetOptions {
	o.names = append([]string{name}, names...)
	return o
}

// WithOptions adds SDK options to the GetItem call following the ones provided by NewManager.
func (o *GetOptions) WithOptions(optFns ...func(*dynamodb.Options)) *GetOptions {
	o.optFns = append(o.optFns, optFns...)
	return o
}
