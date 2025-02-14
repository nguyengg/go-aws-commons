package hmac

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fixedRand always copies from the given src slice.
//
// Unit test replacement of defaultRand.
func fixedRand(src []byte) func([]byte) error {
	return func(dst []byte) error {
		copy(dst, src)
		return nil
	}
}

func TestHasher_SignStable(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	expected, _ := base64.RawStdEncoding.DecodeString("APUBqmRxODn8AoAHhLl7lKnajZltrndrmF/5u6YeMlic")
	ctx := context.Background()

	signer := New(WithKey(key))

	actual, err := signer.Sign(ctx, payload, 0)
	assert.NoErrorf(t, err, "Sign() error = %v", err)
	assert.Equalf(t, expected, actual, "Sign() got = %v, want = %v", base64.RawStdEncoding.EncodeToString(actual), base64.RawStdEncoding.EncodeToString(expected))

	ok, err := signer.Verify(ctx, actual, payload)
	assert.NoErrorf(t, err, "Verify() error = %v", err)
	assert.True(t, ok)
}

func TestHasher_SignStableWithSha1(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	expected, _ := base64.RawStdEncoding.DecodeString("AMMAm3TBNtWG6YlbfiYkyu8v/Wyn")
	ctx := context.Background()

	signer := New(WithKey(key), WithHash(sha1.New))

	actual, err := signer.Sign(ctx, payload, 0)
	assert.NoErrorf(t, err, "Sign() error = %v", err)
	assert.Equalf(t, expected, actual, "Sign() got = %v, want = %v", base64.RawStdEncoding.EncodeToString(actual), base64.RawStdEncoding.EncodeToString(expected))

	ok, err := signer.Verify(ctx, actual, payload)
	assert.NoErrorf(t, err, "Verify() error = %v", err)
	assert.True(t, ok)
}

func TestHasher_SignWithNonce(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	ctx := context.Background()

	signer := New(WithKey(key))

	signature, err := signer.Sign(ctx, payload, 16)
	assert.NoErrorf(t, err, "Sign() error = %v", err)

	ok, err := signer.Verify(ctx, signature, payload)
	assert.NoErrorf(t, err, "Verify() error = %v", err)
	assert.True(t, ok)

	// because we're using a random nonce, signing the same payload again will not produce the same signature.
	signature2, err := signer.Sign(ctx, payload, 16)
	assert.NoErrorf(t, err, "Sign() error = %v", err)
	assert.NotEqual(t, signature, signature2)
}

func TestHasher_SignWithNonceAndFixedRand(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	nonce := []byte("123456")
	expected, _ := base64.RawStdEncoding.DecodeString("AgYxMjM0NTYAjKJl8724bPXGxlIyc+0+toZyuGcadrENOD6DFIuoPP4")
	ctx := context.Background()

	signer := &lev{
		getKey: func(_ context.Context, _ *string) ([]byte, *string, error) {
			return key, nil, nil
		},
		hashProvider: sha256.New,
		rand:         fixedRand(nonce),
	}

	actual, err := signer.Sign(ctx, payload, byte(len(nonce)))
	assert.NoErrorf(t, err, "Sign() error = %v", err)
	assert.Equalf(t, expected, actual, "Sign() got = %v, want = %v", base64.RawStdEncoding.EncodeToString(actual), base64.RawStdEncoding.EncodeToString(expected))

	ok, err := signer.Verify(ctx, actual, payload)
	assert.NoErrorf(t, err, "Verify() error = %v", err)
	assert.True(t, ok)
}
