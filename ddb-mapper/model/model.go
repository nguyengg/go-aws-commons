// Package model parses `dynamodbav` struct tags and collects the metadata in [TableModel].
package model

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
)

// TableModel models a DynamoDB table by parsing `dynamodbav` struct tags similar to [attributevalue.Marshal].
//
// Field tagged with "hashkey", "partitionkey", or "pk" becomes [TableModel.HashKey]. Additionally, the struct tag must
// also provide the [TableModel.TableName].
//
//	ID	string	`dynamodbav:"id,hashkey|partitionkey|pk" tablename:"Items"`
//
// Field tagged with "sortkey", "sk", or "rangekey" becomes [TableModel.SortKey].
//
//	Shard	int	`dynamodbav:"shard,sortkey|sk|rangekey"`
//
// Field tagged with "version" become [TableModel.Version].
//
//	Version	int	`dynamodbav:"version,version"`
//
// Fields tagged with "createdtime" or "modifiedtime" become [TableModel.CreatedTime] and [TableModel.ModifiedTime].
//
//	Created		time.Time	`dynamodbav:"created,createdtime,unixtime"`
//	Modified	time.Time	`dynamodbav:"modified,modifiedtime,unixtime"`
//
// All other attributes are collected in [TableModel.Others].
type TableModel struct {
	// StructType is the type of the struct that models the items in the table.
	//
	// StructType.Kind is guaranteed to be reflect.Struct.
	StructType reflect.Type
	// KeyStructType is StructType with only the key fields.
	KeyStructType reflect.Type
	// TableName is the name of the table.
	TableName string
	// HashKey is the required partition key of the table.
	//
	// Its type is AttributeModelTypeHashKey.
	HashKey *Attribute
	// SortKey is the optional range key of the table.
	SortKey *Attribute
	// Version is the attribute whose `dynamodbav` struct tag is marked as "version".
	Version *Attribute
	// CreatedTime is the attribute whose `dynamodbav` struct tag is marked as "createdtime".
	CreatedTime *Attribute
	// ModifiedTime is the attribute whose `dynamodbav` struct tag is marked as "modifiedtime".
	ModifiedTime *Attribute
	// Others contains all other attributes that have no special role in the table.
	Others map[string]Attribute
}

// Encode marshals the given item.
//
// The item argument can be a struct or struct pointer, and must have the same type as [TableModel.StructType] (you
// can't use Encode as alternative to [attributevalue.MarshalMap]).
func (m TableModel) Encode(item any) (map[string]types.AttributeValue, error) {
	return m.EncodeWith(item, attributevalue.NewEncoder())
}

// EncodeWith is a variation of [TableModel.Encode] that requires an explicit [attributevalue.Encoder].
func (m TableModel) EncodeWith(item any, e *attributevalue.Encoder) (map[string]types.AttributeValue, error) {
	if _, _, err := internal.IndirectValueIsStruct(item, false, m.StructType); err != nil {
		return nil, err
	}

	av, err := e.Encode(item)
	if err != nil {
		return nil, fmt.Errorf("marshal item (type %T) error: %w", item, err)
	}

	avM, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return nil, fmt.Errorf("type %T does not marshal to M data type", item)
	}

	return avM.Value, err
}

// EncodeKeys marshals the given item's keys.
//
// The item argument can be a struct or struct pointer, and must have the same type as [TableModel.StructType].
//
// EncodeKeys is useful when you need to create the map[string]types.AttributeValue for just the key attributes which
// is input parameter to some of these DynamoDB operations:
//   - [DeleteItem]
//   - [GetItem]
//   - [UpdateItem]
//
// [DeleteItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DeleteItem.html#DDB-DeleteItem-request-Key
// [GetItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_GetItem.html#DDB-GetItem-request-Key
// [UpdateItem]: https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_UpdateItem.html#DDB-UpdateItem-request-Key
func (m TableModel) EncodeKeys(item any) (map[string]types.AttributeValue, error) {
	return m.EncodeWith(item, attributevalue.NewEncoder())
}

// EncodeKeysWith is a variation of [TableModel.EncodeKeys] that requires an explicit [attributevalue.Encoder].
func (m TableModel) EncodeKeysWith(item any, e *attributevalue.Encoder) (map[string]types.AttributeValue, error) {
	key := reflect.New(m.KeyStructType).Elem()
	v, _, err := internal.IndirectValueIsStruct(item, false, m.StructType)
	if err != nil {
		return nil, err
	}

	key.FieldByName(m.HashKey.Name).Set(v.FieldByIndex(m.HashKey.Index))
	if m.SortKey != nil {
		key.FieldByName(m.SortKey.Name).Set(v.FieldByIndex(m.SortKey.Index))
	}

	av, err := e.Encode(key.Interface())
	if err != nil {
		return nil, fmt.Errorf("marshal key (type %T) error: %w", item, err)
	}

	avM, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return nil, fmt.Errorf("type %T does not marshal to M data type", item)
	}

	return avM.Value, nil
}

var _ attributevalue.Marshaler
