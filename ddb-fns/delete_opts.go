package ddbfns

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DeleteOptions customises Delete and CreateDeleteItem via chainable methods.
type DeleteOptions struct {
	// EnableOptimisticLocking is true by default to add optimistic locking.
	EnableOptimisticLocking bool

	// TableName modifies the [dynamodb.DeleteItemInput.TableName]
	TableName *string
	// ReturnConsumedCapacity modifies the [dynamodb.DeleteItemInput.ReturnConsumedCapacity]
	ReturnConsumedCapacity types.ReturnConsumedCapacity
	// ReturnItemCollectionMetrics modifies the [dynamodb.DeleteItemInput.ReturnItemCollectionMetrics]
	ReturnItemCollectionMetrics types.ReturnItemCollectionMetrics
	// ReturnValues modifies the [dynamodb.DeleteItemInput.ReturnValues]
	ReturnValues types.ReturnValue
	// ReturnValuesOnConditionCheckFailure modifies the [dynamodb.DeleteItemInput.ReturnValuesOnConditionCheckFailure].
	ReturnValuesOnConditionCheckFailure types.ReturnValuesOnConditionCheckFailure

	// used by CreateDeleteItem.
	condition expression.ConditionBuilder

	// used by Manager.Delete.
	optFns                           []func(*dynamodb.Options)
	oldValues                        interface{}
	oldValuesOnConditionCheckFailure interface{}
}

// DisableOptimisticLocking disables optimistic locking logic.
func (o *DeleteOptions) DisableOptimisticLocking() *DeleteOptions {
	o.EnableOptimisticLocking = false
	return o
}

// WithTableName overrides [DeleteOptions.TableName].
func (o *DeleteOptions) WithTableName(tableName string) *DeleteOptions {
	o.TableName = &tableName
	return o
}

// WithReturnAllOldValues sets the [dynamodb.DeleteItemInput.ReturnValues] to "ALL_OLD" and instructs Delete to decode
// the returned attributes to the optional out argument.
//
// If out is given, it must be a struct pointer that can be passed to [attributevalue.Decoder.Decode].
func (o *DeleteOptions) WithReturnAllOldValues(out interface{}) *DeleteOptions {
	o.oldValues, o.ReturnValues = out, types.ReturnValueAllOld
	return o
}

// WitReturnAllOldValuesOnConditionCheckFailure sets the
// [dynamodb.DeleteItemInput.ReturnValuesOnConditionCheckFailure] to "ALL_OLD" and instructs Delete to decode the
// returned attributes to the optional out argument.
//
// If out is given, it must be a struct pointer that can be passed to [attributevalue.Decoder.Decode].
func (o *DeleteOptions) WitReturnAllOldValuesOnConditionCheckFailure(out interface{}) *DeleteOptions {
	o.oldValuesOnConditionCheckFailure, o.ReturnValuesOnConditionCheckFailure = out, types.ReturnValuesOnConditionCheckFailureAllOld
	return o
}

// WithOptions adds SDK options to the DeleteItem call following the ones provided by NewManager.
func (o *DeleteOptions) WithOptions(optFns ...func(*dynamodb.Options)) *DeleteOptions {
	o.optFns = append(o.optFns, optFns...)
	return o
}

// And adds an expression.And to the condition expression.
func (o *DeleteOptions) And(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *DeleteOptions {
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
func (o *DeleteOptions) Or(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *DeleteOptions {
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
