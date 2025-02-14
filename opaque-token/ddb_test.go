package token

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestKeyConverter_Encode(t *testing.T) {
	tests := []struct {
		name string
		key  map[string]types.AttributeValue
		want string
	}{
		// TODO: Add test cases.
		{
			name: "S hash, B sort",
			key: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "hash"},
				"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
			},
			want: `{"id":{"S":"hash"},"range":{"B":"aGVsbG8sIHdvcmxkIQ"}}`,
		},
		{
			name: "B hash, N sort",
			key: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &types.AttributeValueMemberN{Value: "42"},
			},
			want: `{"id":{"B":"aGVsbG8sIHdvcmxkIQ"},"version":{"N":"42"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := DynamoDBKeyConverter{}
			got, err := c.EncodeKey(context.Background(), tt.key)
			assert.NoErrorf(t, err, "EncodeKey() error = %v", err)
			assert.Equalf(t, tt.want, got, "EncodeKey() got = %v, want = %v", got, tt.want)
		})
	}
}

func TestKeyConverter_Decode(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  map[string]types.AttributeValue
	}{
		// TODO: Add test cases.
		{
			name:  "S hash, B sort",
			token: `{"id":{"S":"hash"},"range":{"B":"aGVsbG8sIHdvcmxkIQ"}}`,
			want: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "hash"},
				"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
			},
		},
		{
			name:  "B hash, N sort",
			token: `{"id":{"B":"aGVsbG8sIHdvcmxkIQ"},"version":{"N":"42"}}`,
			want: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &types.AttributeValueMemberN{Value: "42"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := DynamoDBKeyConverter{}
			got, err := c.DecodeToken(context.Background(), tt.token)
			assert.NoErrorf(t, err, "DecodeToken() error = %v", err)
			assert.Equalf(t, tt.want, got, "DecodeToken() got = %v, want = %v", got, tt.want)
		})
	}
}

func TestKeyConverter_EncodeDecodeWithAES(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")

	tests := []struct {
		name string
		key  map[string]types.AttributeValue
	}{
		{
			name: "S hash, B sort",
			key: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "hash"},
				"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
			},
		},
		{
			name: "B hash, N sort",
			key: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &types.AttributeValueMemberN{Value: "42"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := NewDynamoDBKeyConverter(WithAES(key))

			token, err := c.EncodeKey(context.Background(), tt.key)
			assert.NoErrorf(t, err, "EncodeKey() error = %v", err)

			got, err := c.DecodeToken(context.Background(), token)
			assert.NoErrorf(t, err, "DecodeToken() error = %v", err)

			assert.Equalf(t, tt.key, got, "want = %v, got = %v", tt.key, got)
		})
	}
}

func TestKeyConverter_EncodeDecodeWithChaCha(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")

	tests := []struct {
		name string
		key  map[string]types.AttributeValue
	}{
		// TODO: Add test cases.
		{
			name: "S hash, B sort",
			key: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "hash"},
				"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
			},
		},
		{
			name: "B hash, N sort",
			key: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &types.AttributeValueMemberN{Value: "42"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := NewDynamoDBKeyConverter(WithChaCha20Poly1305(key))

			token, err := c.EncodeKey(context.Background(), tt.key)
			assert.NoErrorf(t, err, "EncodeKey() error = %v", err)

			got, err := c.DecodeToken(context.Background(), token)
			assert.NoErrorf(t, err, "DecodeToken() error = %v", err)

			assert.Equalf(t, tt.key, got, "want = %v, got = %v", tt.key, got)
		})
	}
}
