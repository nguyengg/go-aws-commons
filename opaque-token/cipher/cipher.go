// Package cipher provides utilities to encrypt and decrypt data that fit in memory.
//
// Do not use this package for encrypting and decrypting streams of data; they won't work.
package cipher

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Codec provides methods for encrypting and decrypting data that fit in memory.
type Codec interface {
	// Encode encrypts the given plaintext.
	Encode(ctx context.Context, plaintext []byte) (ciphertext []byte, err error)
	// Decode decrypts the given ciphertext.
	Decode(ctx context.Context, ciphertext []byte) (plaintext []byte, err error)
}

// Seal encrypts the given plaintext with the given [cipher.AEAD].
func Seal(aead cipher.AEAD, plaintext []byte) ([]byte, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Open decrypts the given ciphertext with the given [cipher.AEAD].
func Open(aead cipher.AEAD, ciphertext []byte) ([]byte, error) {
	nonceSize := aead.NonceSize()
	if n := len(ciphertext); n <= nonceSize {
		return nil, fmt.Errorf("ciphertext's size (%d) is less than minimal (%d)", n, nonceSize)
	}

	return aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
}
