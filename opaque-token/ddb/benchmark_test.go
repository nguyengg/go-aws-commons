package ddb

import (
	"context"
	"io"
	"math/rand"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/opaque-token/cipher"
	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
)

func BenchmarkAESWithStaticKey_EncodeDecode(b *testing.B) {
	var seed int64 = 12345
	r := rand.New(rand.NewSource(seed))

	item := map[string]types.AttributeValue{
		"id":    &types.AttributeValueMemberS{Value: "hash"},
		"range": &types.AttributeValueMemberB{Value: []byte("hello, world!")},
	}

	for range b.N {
		secret := make([]byte, 32)
		if _, err := io.ReadFull(r, secret); err != nil {
			panic(err)
		}

		c := &KeyCodec{Codec: cipher.AES(keys.Static(secret))}
		token, err := c.Encode(context.Background(), item)
		if err != nil {
			panic(err)
		}

		if _, err = c.Decode(context.Background(), token); err != nil {
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
		secret := make([]byte, 32)
		if _, err := io.ReadFull(r, secret); err != nil {
			panic(err)
		}

		c := &KeyCodec{Codec: cipher.ChaCha20Poly1305(keys.Static(secret))}
		token, err := c.Encode(context.Background(), item)
		if err != nil {
			panic(err)
		}

		if _, err = c.Decode(context.Background(), token); err != nil {
			panic(err)
		}
	}
}
