package cipher

import (
	"context"
	"crypto/aes"
	"crypto/cipher"

	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
)

// AES creates a new [Codec] using AES algorithm with Galois/Counter Mode (GCM).
func AES(keyProvider keys.Provider) Codec {
	return &aesWithGCM{keyProvider}
}

// aesWithGCM implements [Codec].
type aesWithGCM struct {
	keys.Provider
}

func (a aesWithGCM) Encode(ctx context.Context, plaintext []byte) ([]byte, error) {
	secret, _, err := a.Provide(ctx, nil)
	if err != nil {
		return nil, err
	}

	b, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(b)
	if err != nil {
		return nil, err
	}

	return Seal(gcm, plaintext)
}

func (a aesWithGCM) Decode(ctx context.Context, ciphertext []byte) ([]byte, error) {
	secret, _, err := a.Provide(ctx, nil)
	if err != nil {
		return nil, err
	}

	b, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(b)
	if err != nil {
		return nil, err
	}

	return Open(gcm, ciphertext)
}
