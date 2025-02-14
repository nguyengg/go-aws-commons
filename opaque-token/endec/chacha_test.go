package endec

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ChaCha20Poly1305_EncodeDecode(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	plaintext := []byte("hello, world!")
	ctx := context.Background()

	ed, err := NewChaCha20Poly1305WithKey(key)
	assert.NoErrorf(t, err, "NewChaCha20Poly1305WithKey() error = %v", err)

	ciphertext, err := ed.Encode(ctx, plaintext)
	assert.NoErrorf(t, err, "Encode() error = %v", err)

	decrypted, err := ed.Decode(ctx, ciphertext)
	assert.NoErrorf(t, err, "Decode() error = %v", err)

	assert.Equalf(t, plaintext, decrypted, "Decode() got = %s, want = %s", base64.RawStdEncoding.EncodeToString(decrypted), base64.RawStdEncoding.EncodeToString(plaintext))
}
