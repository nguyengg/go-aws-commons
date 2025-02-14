package endec

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
)

// aesWithGCM uses AES with Galois/Counter Mode (GCM).
type aesWithGCM struct {
	cipher.Block
}

// NewAESWithKey returns a new Endec using the given AES key.
//
// The key must be either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewAESWithKey(key []byte) (Endec, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return aesWithGCM{b}, nil
}

// NewAES returns a new EndecWithKey using AES key given at encryption/decryption time.
func NewAES() EndecWithKey {
	return aesWithGCM{}
}

func (a aesWithGCM) Encode(_ context.Context, plaintext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(a)
	if err != nil {
		return nil, err
	}

	return Seal(gcm, plaintext)
}

func (a aesWithGCM) EncodeWithKey(_ context.Context, key, plaintext []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(b)
	if err != nil {
		return nil, err
	}

	return Seal(gcm, plaintext)
}

func (a aesWithGCM) Decode(_ context.Context, ciphertext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(a)
	if err != nil {
		return nil, err
	}

	return Open(gcm, ciphertext)
}

func (a aesWithGCM) DecodeWithKey(_ context.Context, key, ciphertext []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(b)
	if err != nil {
		return nil, err
	}

	return Open(gcm, ciphertext)
}
