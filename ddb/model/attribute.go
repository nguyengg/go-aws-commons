package model

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Attribute contains metadata about a reflect.StructField that represents a DynamoDB attribute.
type Attribute struct {
	// Field is the original struct field that was parsed.
	Field reflect.StructField
	// Name is the first tag value in the `dynamodbav` struct tag.
	Name string
	// DataType describes the type of data that this attribute stores.
	DataType DataType
	// OmitEmpty is true only if the `dynamodbav` struct tag also includes `omitempty`.
	OmitEmpty bool
	// UnixTime is true only if the `dynamodbav` struct tag also includes `unixtime`.
	UnixTime bool
}

// Get returns the reflected value from the given struct value.
func (a *Attribute) Get(value reflect.Value) (reflect.Value, error) {
	return value.FieldByIndexErr(a.Field.Index)
}

// DataType models DynamoDB data type.
type DataType int

const (
	DataTypeUnknown DataType = iota
	DataTypeS
	DataTypeN
	DataTypeB
)

func (t DataType) String() string {
	switch t {
	case DataTypeS:
		return "S"
	case DataTypeN:
		return "N"
	case DataTypeB:
		return "B"
	default:
		return "?"
	}
}

// EncodeKey encodes the given value as a DynamoDB key attribute value (can only be S, N, or B type).
func (a *Attribute) EncodeKey(value interface{}) (types.AttributeValue, error) {
	switch a.DataType {
	case DataTypeS:
		switch v := value.(type) {
		case string:
			return &types.AttributeValueMemberS{Value: v}, nil
		default:
			return nil, fmt.Errorf("expect key of type string, got %T", value)
		}
	case DataTypeN:
		switch v := value.(type) {
		// https://go.dev/ref/spec#Type_switches requires multiple cases for the code to compile naturally.
		case int:
			return &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(v), 10)}, nil
		case int8:
			return &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(v), 10)}, nil
		case int16:
			return &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(v), 10)}, nil
		case int32:
			return &types.AttributeValueMemberN{Value: strconv.FormatInt(int64(v), 10)}, nil
		case int64:
			return &types.AttributeValueMemberN{Value: strconv.FormatInt(v, 10)}, nil
		case uint:
			return &types.AttributeValueMemberN{Value: strconv.FormatUint(uint64(v), 10)}, nil
		case uint8:
			return &types.AttributeValueMemberN{Value: strconv.FormatUint(uint64(v), 10)}, nil
		case uint16:
			return &types.AttributeValueMemberN{Value: strconv.FormatUint(uint64(v), 10)}, nil
		case uint32:
			return &types.AttributeValueMemberN{Value: strconv.FormatUint(uint64(v), 10)}, nil
		case uint64:
			return &types.AttributeValueMemberN{Value: strconv.FormatUint(v, 10)}, nil
		case float32, float64:
			// strconv.FormatFloat() requires precision and I don't know what to give.
			// besides, using reflect.String matches what attributevalue.Encoder does.
			return &types.AttributeValueMemberN{Value: fmt.Sprintf("%s", v)}, nil
		default:
			return nil, fmt.Errorf("expect key of numeric type, got %T", value)
		}
	case DataTypeB:
		switch v := value.(type) {
		case []byte:
			return &types.AttributeValueMemberB{Value: v}, nil
		default:
			return nil, fmt.Errorf("expect key of byte slice, got %T", value)
		}
	default:
		return nil, fmt.Errorf("attribute's data type is not valid for key")
	}
}
