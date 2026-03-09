package types

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// StringSet is a set of strings encoded as DynamoDB SS data type.
//
// The zero value StringSet is ready for use.
type StringSet struct {
	// m is embedded instead of alias to allow zero value StringSet to work.
	m map[string]struct{}
}

// NewStringSet creates a new StringSet.
func NewStringSet(s ...string) StringSet {
	set := StringSet{m: make(map[string]struct{})}

	for _, v := range s {
		set.m[v] = struct{}{}
	}

	return set
}

// Add returns true only if the value hasn't existed in the set.
func (set *StringSet) Add(v string) bool {
	if _, ok := set.m[v]; !ok {
		if set.m == nil {
			set.m = map[string]struct{}{v: {}}
		} else {
			set.m[v] = struct{}{}
		}

		return true
	}

	return false
}

// Has returns true only if the value exists in the set.
func (set StringSet) Has(v string) bool {
	_, ok := set.m[v]
	return ok
}

// Delete removes the value from the set and returns true only if the value existed in the set.
func (set *StringSet) Delete(v string) bool {
	if set == nil {
		return false
	}

	if _, ok := set.m[v]; ok {
		delete(set.m, v)
		return true
	}

	return false
}

// Clear empties the set.
func (set *StringSet) Clear() {
	if set != nil {
		set.m = make(map[string]struct{})
	}
}

// Size returns the number of entries in the set.
func (set StringSet) Size() int {
	return len(set.m)
}

// Equal returns true only if every element in this set exists in the other set and vice versa.
func (set StringSet) Equal(other StringSet) bool {
	return maps.Equal(set.m, other.m)
}

func (set StringSet) sortedKeys() (keys []string) {
	for k := range set.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

func (set StringSet) String() string {
	return fmt.Sprintf("%s", set.sortedKeys())
}

var _ attributevalue.Marshaler = StringSet{}
var _ attributevalue.Unmarshaler = (*StringSet)(nil)

func (set StringSet) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	keys := set.sortedKeys()
	if len(keys) == 0 {
		return &types.AttributeValueMemberNULL{Value: true}, nil
	}

	return &types.AttributeValueMemberSS{Value: set.sortedKeys()}, nil
}

func (set *StringSet) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	avSS, ok := value.(*types.AttributeValueMemberSS)
	if !ok {
		return nil
	}

	clear(set.m)

	if set.m == nil {
		set.m = make(map[string]struct{})
	}

	for _, k := range avSS.Value {
		set.m[k] = struct{}{}
	}

	return nil
}

var _ json.Marshaler = StringSet{}
var _ json.Unmarshaler = &StringSet{}

func (set StringSet) MarshalJSON() ([]byte, error) {
	keys := set.sortedKeys()
	if len(keys) == 0 {
		return []byte("[]"), nil
	}

	// the rest is just json.Marshal(keys) with premature optimisation :D
	var b strings.Builder
	b.WriteRune('[')
	for i, k := range keys {
		if i != 0 {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(k)
		b.WriteRune('"')
	}
	b.WriteRune(']')

	return []byte(b.String()), nil
}

func (set *StringSet) UnmarshalJSON(bytes []byte) error {
	set.m = nil
	switch string(bytes) {
	case "null", "[]":
		return nil
	}

	keys := make([]string, 0)
	if err := json.Unmarshal(bytes, &keys); err != nil {
		return err
	}

	if n := len(keys); n != 0 {
		set.m = make(map[string]struct{}, n)
	}

	for _, k := range keys {
		set.m[k] = struct{}{}
	}

	return nil
}
