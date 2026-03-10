// Package ddb provides [KeyCodec] to encode DynamoDB last evaluated keys an opaque token that can be returned to
// API caller without exposing the internal details of the table.
package ddb

import (
	"context"
	"encoding/base64"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/opaque-token/cipher"
	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
)

// KeyCodec encodes last evaluated key (map[string]types.AttributeValue) to an opaque token string and vice versa.
//
// The zero value is ready to use without any encryption or decryption.
type KeyCodec struct {
	// Encoding must be given to encode the ciphertext to string, and decode the string back to ciphertext.
	//
	// By default, [base64.RawURLEncoding] is used since the intention is to return the opaque token in API usage.
	Encoding Encoding

	// Codec is optionally given to encrypt and decrypt the token, making the token truly opaque.
	//
	// Without a Codec, KeyCodec simply applies [Encoding] on top of the JSON encoding of the keys.
	Codec cipher.Codec
}

// Encode encodes the last evaluated key and optionally encrypts the returned string token.
func (c KeyCodec) Encode(ctx context.Context, key map[string]types.AttributeValue) (string, error) {
	var (
		plaintext, ciphertext []byte
		err                   error
	)

	if plaintext, err = attributevalue.MarshalMapJSON(key); err != nil {
		return "", err
	}

	encoder := c.Encoding
	if encoder == nil {
		encoder = base64.RawURLEncoding
	}

	if encrypter := c.Codec; encrypter != nil {
		if ciphertext, err = encrypter.Encode(ctx, plaintext); err != nil {
			return "", err
		}

		return encoder.EncodeToString(ciphertext), nil
	}

	return encoder.EncodeToString(plaintext), nil
}

// Decode decodes the given string to an exclusive start key.
func (c KeyCodec) Decode(ctx context.Context, token string) (map[string]types.AttributeValue, error) {
	var (
		plaintext, ciphertext []byte
		err                   error
	)

	decoder := c.Encoding
	if decoder == nil {
		decoder = base64.RawURLEncoding
	}

	if decrypter := c.Codec; decrypter != nil {
		if ciphertext, err = decoder.DecodeString(token); err != nil {
			return nil, err
		}

		if plaintext, err = decrypter.Decode(ctx, ciphertext); err != nil {
			return nil, err
		}
	} else if plaintext, err = decoder.DecodeString(token); err != nil {
		return nil, err
	}

	avM, err := attributevalue.UnmarshalMapJSON(plaintext)
	if err != nil {
		return nil, err
	}

	return avM, nil
}

// New returns a new [KeyCodec] that uses ChaCha20-Poly1305.
//
// See [keys] package for several options to construct an [Engine]:
//   - [keys.Static] and [keys.FromEnv] are appropriate for testing.
//   - [keys.FromSecretsManager] is good in production and gets you secret rotation for free.
//   - Consider [keys.FromLambdaExtensionSecrets] if running in Lambda with [AWS Parameters and Secrets Lambda Extension]
//     enabled.
//
// If you want to use AES or some other algorithm instead, use the optFns argument to modify [KeyCodec.Codec].
//
// [AWS Parameters and Secrets Lambda Extension]: https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html
func New(keyProvider keys.Provider, optFns ...func(c *KeyCodec)) *KeyCodec {
	c := &KeyCodec{
		Encoding: base64.RawURLEncoding,
		Codec:    cipher.ChaCha20Poly1305(keyProvider),
	}
	for _, fn := range optFns {
		fn(c)
	}

	return c
}
