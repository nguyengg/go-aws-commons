package ddb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/opaque-token/cipher"
	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
	"github.com/stretchr/testify/assert"
)

type raw struct {
}

func (r raw) EncodeToString(src []byte) string {
	return string(src)
}

func (r raw) DecodeString(s string) ([]byte, error) {
	return []byte(s), nil
}

func TestKeyCodec_Encode_RawEncoding(t *testing.T) {
	tests := []struct {
		name string
		key  map[string]types.AttributeValue
		want string
	}{
		{
			name: "S hash, B sort",
			key: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "hash"},
				"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
			},
			want: `{"id":{"S":"hash"},"range":{"B":"aGVsbG8sIHdvcmxkIQ=="}}`,
		},
		{
			name: "B hash, N sort",
			key: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &types.AttributeValueMemberN{Value: "42"},
			},
			want: `{"id":{"B":"aGVsbG8sIHdvcmxkIQ=="},"version":{"N":"42"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := KeyCodec{Encoding: raw{}}
			got, err := c.Encode(context.Background(), tt.key)
			assert.NoErrorf(t, err, "EncodeKey() error = %v", err)
			assert.JSONEqf(t, tt.want, got, "EncodeKey() got = %v, want = %v", got, tt.want)
		})
	}
}

func TestKeyCodec_Decode_RawEncoding(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  map[string]types.AttributeValue
	}{
		{
			name:  "S hash, B sort",
			token: `{"id":{"S":"hash"},"range":{"B":"aGVsbG8sIHdvcmxkIQ=="}}`,
			want: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "hash"},
				"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
			},
		},
		{
			name:  "B hash, N sort",
			token: `{"id":{"B":"aGVsbG8sIHdvcmxkIQ=="},"version":{"N":"42"}}`,
			want: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberB{Value: []byte("hello, world!")},
				"version": &types.AttributeValueMemberN{Value: "42"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := KeyCodec{Encoding: raw{}}
			got, err := c.Decode(context.Background(), tt.token)
			assert.NoErrorf(t, err, "DecodeToken() error = %v", err)
			assert.Equalf(t, tt.want, got, "DecodeToken() got = %v, want = %v", got, tt.want)
		})
	}
}

func TestKeyConverter_EncodeDecodeWithAES(t *testing.T) {
	secret := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")

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
			c := &KeyCodec{Codec: cipher.AES(keys.Static(secret))}

			token, err := c.Encode(context.Background(), tt.key)
			assert.NoErrorf(t, err, "EncodeKey() error = %v", err)

			got, err := c.Decode(context.Background(), token)
			assert.NoErrorf(t, err, "DecodeToken() error = %v", err)

			assert.Equalf(t, tt.key, got, "want = %v, got = %v", tt.key, got)
		})
	}
}

func TestKeyConverter_EncodeDecodeWithChaCha(t *testing.T) {
	secret := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")

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
			c := &KeyCodec{Codec: cipher.ChaCha20Poly1305(keys.Static(secret))}

			token, err := c.Encode(context.Background(), tt.key)
			assert.NoErrorf(t, err, "EncodeKey() error = %v", err)

			got, err := c.Decode(context.Background(), token)
			assert.NoErrorf(t, err, "DecodeToken() error = %v", err)

			assert.Equalf(t, tt.key, got, "want = %v, got = %v", tt.key, got)
		})
	}
}
