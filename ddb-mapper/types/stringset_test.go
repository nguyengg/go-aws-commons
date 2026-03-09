package types

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestStringSet_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		s    StringSet
		want string
	}{
		{
			name: "marshal",
			s:    NewStringSet("a", "b", "c"),
			want: `["a","b","c"]`,
		},
		{
			name: "marshal sorts results",
			s:    NewStringSet("c", "b", "a"),
			want: `["a","b","c"]`,
		},
		{
			name: "marshal removes duplicates",
			s:    NewStringSet("b", "c", "b", "a"),
			want: `["a","b","c"]`,
		},
		{
			name: "NewStringSet",
			s:    NewStringSet(),
			want: `[]`,
		},
		{
			name: "StringSet{}",
			s:    StringSet{},
			want: `[]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.s)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestStringSet_MarshalJSON_StructField(t *testing.T) {
	tests := []struct {
		name string
		i    any
		want string
	}{
		{
			name: "marshal zero value",
			i: struct {
				Tags StringSet `json:"tags"`
			}{},
			want: `{"tags":[]}`,
		},
		{
			name: "marshal some values",
			i: struct {
				Tags StringSet `json:"tags"`
			}{
				Tags: NewStringSet("b", "c", "a"),
			},
			want: `{"tags":["a","b","c"]}`,
		},
		{
			name: "marshal omitempty",
			i: struct {
				Tags *StringSet `json:"tags,omitempty"`
			}{},
			want: `{}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.i)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestStringSet_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		args string
		want StringSet
	}{
		{
			name: "unmarshal",
			args: `["a","b","c"]`,
			want: NewStringSet("c", "a", "b"),
		},
		{
			name: "unmarshal removes duplicates",
			args: `["a","b","c","a"]`,
			want: NewStringSet("b", "c", "a", "b"),
		},
		{
			name: "unmarshal empty array",
			args: `[]`,
			want: StringSet{},
		},
		{
			name: "unmarshal null",
			args: `null`,
			want: StringSet{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s StringSet
			err := json.Unmarshal([]byte(tt.args), &s)
			assert.NoErrorf(t, err, "UnmarshalJSON() error = %v", err)
			assert.Equal(t, tt.want, s)
		})
	}
}

func TestStringSet_MarshalDynamoDBAttributeValue(t *testing.T) {
	type Item struct {
		Tags StringSet `dynamodbav:"tags"`
	}

	tests := []struct {
		name string
		item Item
		want types.AttributeValue
	}{
		{
			name: "marshal set with values",
			item: Item{NewStringSet("b", "c", "a")},
			want: &types.AttributeValueMemberSS{Value: []string{"a", "b", "c"}},
		},
		{
			name: "allow empty elements",
			item: Item{NewStringSet("")},
			want: &types.AttributeValueMemberSS{Value: []string{""}},
		},
		{
			name: "empty set is marshaled as NULL",
			item: Item{NewStringSet()},
			want: &types.AttributeValueMemberNULL{Value: true},
		},
		{
			name: "unmarshal zero value",
			item: Item{},
			want: &types.AttributeValueMemberNULL{Value: true},
		},
		{
			name: "unmarshal default value",
			item: Item{StringSet{}},
			want: &types.AttributeValueMemberNULL{Value: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avM, err := attributevalue.Marshal(tt.item)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, avM.(*types.AttributeValueMemberM).Value["tags"])
		})
	}
}

func TestStringSet_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name string
		av   types.AttributeValue
		want StringSet
	}{
		{
			name: "unmarshal set with values",
			av:   &types.AttributeValueMemberSS{Value: []string{"a", "b", "c"}},
			want: NewStringSet("b", "c", "a"),
		},
		{
			name: "allow empty elements",
			av:   &types.AttributeValueMemberSS{Value: []string{""}},
			want: NewStringSet(""),
		},
		{
			name: "NULL is unmarshalled as zero value",
			av:   &types.AttributeValueMemberNULL{Value: true},
			want: StringSet{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s StringSet
			assert.NoError(t, attributevalue.Unmarshal(tt.av, &s))
			assert.Equal(t, tt.want, s)
		})
	}
}

func TestStringSet_Add(t *testing.T) {
	var s StringSet
	assert.True(t, s.Add("a"))
	assert.True(t, s.Has("a"))
	assert.False(t, s.Add("a"))
}
