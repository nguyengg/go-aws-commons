package endec

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Endec provides methods for encrypting and decrypting data that fit in memory.
//
// Endec instances have their own ways to retrieve the key as opposed to EndecWithKey. They can be used to encrypt/decrypt
// binary (or string) tokens to make them opaque.
type Endec interface {
	// Encode encrypts the given plaintext.
	Encode(ctx context.Context, plaintext []byte) (ciphertext []byte, err error)
	// Decode decrypts the given ciphertext.
	Decode(ctx context.Context, ciphertext []byte) (plaintext []byte, err error)
}

// EndecWithKey provides methods for encrypting and decrypting data that fit in memory.
//
// EndecWithKey instances must be given a key at encryption/decryption time as opposed to Endec.
type EndecWithKey interface {
	// EncodeWithKey encrypts the plaintext with the given key.
	EncodeWithKey(ctx context.Context, key, plaintext []byte) (ciphertext []byte, err error)
	// DecodeWithKey decrypts the ciphertext with the given key.
	DecodeWithKey(ctx context.Context, key, ciphertext []byte) (plaintext []byte, err error)
}

// Seal encrypts the given plaintext with the given cipher.AEAD.
func Seal(aead cipher.AEAD, plaintext []byte) ([]byte, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Open decrypts the given ciphertext with the given cipher.AEAD.
func Open(aead cipher.AEAD, ciphertext []byte) ([]byte, error) {
	nonceSize := aead.NonceSize()
	if n := len(ciphertext); n <= nonceSize {
		return nil, fmt.Errorf("ciphertext's size (%d) is less than minimal (%d)", n, nonceSize)
	}

	return aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
}
