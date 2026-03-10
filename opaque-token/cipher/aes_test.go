package cipher

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
	"github.com/stretchr/testify/assert"
)

func Test_NewAESWithGCM_EncodeDecode(t *testing.T) {
	secret := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	plaintext := []byte("hello, world!")
	ctx := context.Background()

	ed := AES(keys.Static(secret))

	ciphertext, err := ed.Encode(ctx, plaintext)
	assert.NoErrorf(t, err, "Encode() error = %v", err)

	decrypted, err := ed.Decode(ctx, ciphertext)
	assert.NoErrorf(t, err, "Decode() error = %v", err)

	assert.Equalf(t, plaintext, decrypted, "Decode() got = %s, want = %s", base64.RawStdEncoding.EncodeToString(decrypted), base64.RawStdEncoding.EncodeToString(plaintext))
}
