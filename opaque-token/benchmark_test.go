package token

import (
	"context"
	"io"
	"math/rand"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func BenchmarkAESWithStaticKey_EncodeDecode(b *testing.B) {
	var seed int64 = 12345
	r := rand.New(rand.NewSource(seed))

	item := map[string]types.AttributeValue{
		"id":    &types.AttributeValueMemberS{Value: "hash"},
		"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
	}

	for range b.N {
		key := make([]byte, 32)
		if _, err := io.ReadFull(r, key); err != nil {
			panic(err)
		}

		c, _ := NewDynamoDBKeyConverter(WithAES(key))
		token, err := c.EncodeKey(context.Background(), item)
		if err != nil {
			panic(err)
		}

		if _, err = c.DecodeToken(context.Background(), token); err != nil {
			panic(err)
		}
	}
}

func BenchmarkChaCha20Poly1305WithStaticKey_EncodeDecode(b *testing.B) {
	var seed int64 = 12345
	r := rand.New(rand.NewSource(seed))

	item := map[string]types.AttributeValue{
		"id":    &types.AttributeValueMemberS{Value: "hash"},
		"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
	}

	for range b.N {
		key := make([]byte, 32)
		if _, err := io.ReadFull(r, key); err != nil {
			panic(err)
		}

		c, err := NewDynamoDBKeyConverter(WithChaCha20Poly1305(key))
		token, err := c.EncodeKey(context.Background(), item)
		if err != nil {
			panic(err)
		}

		if _, err = c.DecodeToken(context.Background(), token); err != nil {
			panic(err)
		}
	}
}
