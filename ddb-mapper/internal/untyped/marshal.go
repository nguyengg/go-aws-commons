package untyped

import (
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/internal"
)

// encode marshals the given item to DynamoDB attribute map.
func (m *Mapper) encode(c *Context, item any) (map[string]types.AttributeValue, error) {
	av, err := c.Encoder.Encode(item)
	if err != nil {
		return nil, fmt.Errorf("marshal item (type %T) error: %w", item, err)
	}

	avM, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return nil, fmt.Errorf("type %T does not marshal to M data type", item)
	}

	return avM.Value, nil
}

// keys is a variant of encode that returns the DynamoDB attribute map containing only key attributes.
func (m *Mapper) keys(c *Context, item any) (map[string]types.AttributeValue, error) {
	key := reflect.New(m.KeyStructType).Elem()
	v, _, err := internal.IndirectValueIsStruct(item, false, m.StructType)
	if err != nil {
		return nil, err
	}

	key.FieldByName(m.HashKey.Name).Set(v.FieldByIndex(m.HashKey.Index))
	if m.SortKey != nil {
		key.FieldByName(m.SortKey.Name).Set(v.FieldByIndex(m.SortKey.Index))
	}

	av, err := c.Encoder.Encode(key.Interface())
	if err != nil {
		return nil, fmt.Errorf("marshal key (type %T) error: %w", item, err)
	}

	avM, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return nil, fmt.Errorf("type %T does not marshal to M data type", item)
	}

	return avM.Value, nil
}

// copyKeys returns a new map[string]AttributeValue that only contains key attributes.
func (m *Mapper) copyKeys(item map[string]types.AttributeValue) map[string]types.AttributeValue {
	keys := make(map[string]types.AttributeValue)

	if av, ok := item[m.HashKey.AttrName]; ok {
		keys[m.HashKey.AttrName] = av
	}

	if m.SortKey != nil {
		if av, ok := item[m.SortKey.AttrName]; ok {
			keys[m.SortKey.AttrName] = av
		}
	}

	return keys
}

// ExtractKeys extracts just the key attributes from the given item.
func (m *Mapper) ExtractKeys(item any) (map[string]types.AttributeValue, error) {
	return m.keys(DefaultContext(), item)
}
