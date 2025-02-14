package endec

import (
	"context"
	"crypto/cipher"

	"golang.org/x/crypto/chacha20poly1305"
)

// chaCha20Poly1305 uses ChaCha20-Poly1305 and requires 256-bit (32-byte) key.
type chaCha20Poly1305 struct {
	cipher.AEAD
}

// NewChaCha20Poly1305WithKey returns a ChaCha20-Poly1305 Endec that uses the given 256-bit key.
func NewChaCha20Poly1305WithKey(key []byte) (Endec, error) {
	b, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	return chaCha20Poly1305{b}, nil
}

// NewChaCha20Poly1305 returns a new ChaCha20-Poly1305 EndecWithKey using 256-bit key given at encryption/decryption time.
func NewChaCha20Poly1305() EndecWithKey {
	return chaCha20Poly1305{}
}

func (c chaCha20Poly1305) Encode(_ context.Context, plaintext []byte) ([]byte, error) {
	return Seal(c, plaintext)
}

func (c chaCha20Poly1305) Decode(_ context.Context, ciphertext []byte) ([]byte, error) {
	return Open(c, ciphertext)
}

func (c chaCha20Poly1305) EncodeWithKey(_ context.Context, key, plaintext []byte) ([]byte, error) {
	b, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	return Seal(b, plaintext)
}

func (c chaCha20Poly1305) DecodeWithKey(_ context.Context, key, ciphertext []byte) ([]byte, error) {
	b, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	return Open(b, ciphertext)
}
