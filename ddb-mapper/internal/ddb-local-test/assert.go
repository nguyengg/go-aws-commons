package localtest

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RequireHasAttributes is a require-variation of AssertHasAttributes.
func RequireHasAttributes(t *testing.T, expected map[string]any, item map[string]types.AttributeValue) {
	enc := attributevalue.NewEncoder()
	for name, value := range expected {
		a, ok := item[name]
		if !ok {
			require.Fail(t, "item does not contain attribute %q", name)
		}

		switch e := value.(type) {
		case types.AttributeValue:
			require.Equalf(t, e, a, "mismatched attribute %q: got %v, want %v", name, a, e)
		case AttributeAssertFn:
			e(t, name, a)
		default:
			attr, err := enc.Encode(e)
			require.NoErrorf(t, err, "encode %v error", e)
			require.Equalf(t, attr, a, "mismatched attribute %q: got %v, want %v", name, a, e)
		}
	}
}

// AttributeAssertFn is used by assertHasAttributes.
type AttributeAssertFn func(t *testing.T, name string, value types.AttributeValue)

// As expects the attribute value to be attributevalue-decodable to type T.
func As[T any](t *testing.T, value types.AttributeValue) T {
	tType := reflect.TypeFor[T]()
	p := reflect.New(tType)
	assert.NoErrorf(t, attributevalue.Unmarshal(value, p.Interface()), "unmarshal value (%v) as type %s error", value, tType)
	return p.Elem().Interface().(T)
}

// AsUnixTime expects the attribute value to be a unixtime (N data type).
func AsUnixTime(t *testing.T, value types.AttributeValue) time.Time {
	asN, ok := value.(*types.AttributeValueMemberN)
	require.Truef(t, ok, "value (%v) is not N data type", value)
	epochSecond, err := strconv.ParseInt(asN.Value, 10, 64)
	require.NoErrorf(t, err, "parse value.N (%q) as integer error", asN.Value)
	return time.Unix(epochSecond, 0)
}

// MustMarshalToM requires that the given item marshals successfully to M data type.
func MustMarshalToM(t *testing.T, item any) map[string]types.AttributeValue {
	av, err := attributevalue.Marshal(item)
	require.NoErrorf(t, err, "Marshal(%T)", item)
	avM, ok := av.(*types.AttributeValueMemberM)
	require.Truef(t, ok, "item of type %T did not marshal to M data type; got %T instead", item, av)
	return avM.Value
}
