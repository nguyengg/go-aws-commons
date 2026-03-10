package cipher

import (
	"context"

	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
	"golang.org/x/crypto/chacha20poly1305"
)

// ChaCha20Poly1305 creates a new [Codec] using ChaCha20-Poly1305 algorithm.
func ChaCha20Poly1305(keyProvider keys.Provider) Codec {
	return &chaCha20Poly1305{keyProvider}
}

// chaCha20Poly1305 implements [Codec].
type chaCha20Poly1305 struct {
	keys.Provider
}

func (c chaCha20Poly1305) Encode(ctx context.Context, plaintext []byte) ([]byte, error) {
	secret, _, err := c.Provide(ctx, nil)
	if err != nil {
		return nil, err
	}

	b, err := chacha20poly1305.New(secret)
	if err != nil {
		return nil, err
	}

	return Seal(b, plaintext)
}

func (c chaCha20Poly1305) Decode(ctx context.Context, ciphertext []byte) ([]byte, error) {
	secret, _, err := c.Provide(ctx, nil)
	if err != nil {
		return nil, err
	}

	b, err := chacha20poly1305.New(secret)
	if err != nil {
		return nil, err
	}

	return Open(b, ciphertext)
}
