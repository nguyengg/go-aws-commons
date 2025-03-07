package ddb

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// PutOptions customises Put and CreatePutItem via chainable methods.
type PutOptions struct {
	// EnableOptimisticLocking is true by default to add optimistic locking.
	EnableOptimisticLocking bool
	// EnableAutoGeneratedTimestamps is true by default to add generated timestamp attributes.
	EnableAutoGeneratedTimestamps bool

	// TableName modifies the [dynamodb.PutItemInput.TableName]
	TableName *string
	// ReturnConsumedCapacity modifies the [dynamodb.PutItemInput.ReturnConsumedCapacity]
	ReturnConsumedCapacity types.ReturnConsumedCapacity
	// ReturnItemCollectionMetrics modifies the [dynamodb.PutItemInput.ReturnItemCollectionMetrics]
	ReturnItemCollectionMetrics types.ReturnItemCollectionMetrics
	// ReturnValues modifies the [dynamodb.PutItemInput.ReturnValues]
	ReturnValues types.ReturnValue
	// ReturnValuesOnConditionCheckFailure modifies the [dynamodb.PutItemInput.ReturnValuesOnConditionCheckFailure].
	ReturnValuesOnConditionCheckFailure types.ReturnValuesOnConditionCheckFailure

	// used by CreatePutItem.
	condition expression.ConditionBuilder

	// used by Manager.Put.
	optFns                           []func(*dynamodb.Options)
	oldValues                        interface{}
	oldValuesOnConditionCheckFailure interface{}
}

// DisableOptimisticLocking disables optimistic locking logic.
func (o *PutOptions) DisableOptimisticLocking() *PutOptions {
	o.EnableOptimisticLocking = false
	return o
}

// DisableAutoGeneratedTimestamps disables auto-generated timestamps logic.
func (o *PutOptions) DisableAutoGeneratedTimestamps() *PutOptions {
	o.EnableAutoGeneratedTimestamps = false
	return o
}

// WithTableName overrides [PutOptions.TableName].
func (o *PutOptions) WithTableName(tableName string) *PutOptions {
	o.TableName = &tableName
	return o
}

// WithReturnAllOldValues sets the [dynamodb.PutItemInput.ReturnValues] to "ALL_OLD" and instructs Put to decode the
// returned attributes to the optional out argument.
//
// If out is given, it must be a struct pointer that can be passed to [attributevalue.Decoder.Decode].
func (o *PutOptions) WithReturnAllOldValues(out interface{}) *PutOptions {
	o.oldValues, o.ReturnValues = out, types.ReturnValueAllOld
	return o
}

// WithReturnAllOldValuesOnConditionCheckFailure sets the
// [dynamodb.PutItemInput.ReturnValuesOnConditionCheckFailure] to "ALL_OLD" and instructs Put to decode the returned
// attributes to the optional out argument.
//
// If out is given, it must be a struct pointer that can be passed to [attributevalue.Decoder.Decode].
func (o *PutOptions) WithReturnAllOldValuesOnConditionCheckFailure(out interface{}) *PutOptions {
	o.oldValuesOnConditionCheckFailure, o.ReturnValuesOnConditionCheckFailure = out, types.ReturnValuesOnConditionCheckFailureAllOld
	return o
}

// WithOptions adds SDK options to the PutItem call following the ones provided by NewManager.
func (o *PutOptions) WithOptions(optFns ...func(*dynamodb.Options)) *PutOptions {
	o.optFns = append(o.optFns, optFns...)
	return o
}

// And adds an expression.And to the condition expression.
func (o *PutOptions) And(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *PutOptions {
	if o.condition.IsSet() {
		o.condition = o.condition.And(right, other...)
		return o
	}

	switch len(other) {
	case 0:
		o.condition = right
	case 1:
		o.condition = right.And(other[0])
	default:
		o.condition = right.And(other[0], other[1:]...)
	}
	return o
}

// Or adds an expression.Or to the condition expression.
func (o *PutOptions) Or(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *PutOptions {
	if o.condition.IsSet() {
		o.condition = o.condition.Or(right, other...)
		return o
	}

	switch len(other) {
	case 0:
		o.condition = right
	case 1:
		o.condition = right.Or(other[0])
	default:
		o.condition = right.Or(other[0], other[1:]...)
	}
	return o
}
