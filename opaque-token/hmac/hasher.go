package hmac

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"hash"
	"io"

	"github.com/nguyengg/go-aws-commons/lambda"
	"github.com/nguyengg/go-aws-commons/opaque-token/endec"
)

// Hasher provides HMAC signing and validation methods.
//
// See New regarding options to create a Hasher instance.
type Hasher interface {
	// Sign creates an HMAC signature from the given payload.
	//
	// If nonce size is 0, the same payload will always produce the same signature.
	//
	// In order to use the signature as CSRF token, pass a non-zero value for the nonce size (16 is a good length).
	// According to
	// https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#pseudo-code-for-implementing-hmac-csrf-tokens,
	// the payload should include the session id and any other information you wish. Do not include a random value
	// in the payload; this method already creates the random value for you from the given nonce size.
	Sign(ctx context.Context, payload []byte, nonceSize byte) ([]byte, error)

	// Verify validates the given signature against the expected payload.
	//
	// The signature should have been created by a previous call to Sign.
	//
	// The boolean return value is true if and only if the signature has passed all validation. When the boolean
	// return value is false and there is no error, the signature passes all parsing but fails at the final
	// comparing step. Otherwise, any parsing error will be returned.
	Verify(ctx context.Context, signature, payload []byte) (bool, error)
}

type lev struct {
	getKey       func(context.Context, *string) ([]byte, *string, error)
	hashProvider func() hash.Hash
	rand         func([]byte) error
}

// KeyProvider customises how the Hasher retrieves its key.
type KeyProvider func(*lev)

// Option customises other aspects of Hasher.
type Option func(*lev)

// New returns a new Hasher for creating and validating signed tokens.
//
// If you have static key, pass WithKey.
// If you want to retrieve secret binary from AWS Secrets Hasher, pass WithKeyFromSecretsManager.
// If you are running in AWS Lambda with AWS Parameters and Secrets Lambda Extension
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html) enabled, pass
// WithKeyFromLambdaExtensionSecrets.
//
// If you want to use a specific hash function instead of sha256.New, use WithHash.
func New(keyProvider KeyProvider, optFns ...Option) Hasher {
	m := &lev{
		hashProvider: sha256.New,
		rand:         defaultRand,
	}

	keyProvider(m)

	for _, fn := range optFns {
		fn(m)
	}

	return m
}

// WithKey uses a fixed key for signing and verification.
func WithKey(key []byte) KeyProvider {
	// copy so that caller cannot mutate the key.
	dst := make([]byte, len(key))
	copy(dst, key)

	return func(m *lev) {
		m.getKey = func(ctx context.Context, s *string) ([]byte, *string, error) {
			return dst, nil, nil
		}
	}
}

// WithKeyFromSecretsManager retrieves key from AWS Secrets Hasher.
func WithKeyFromSecretsManager(client endec.GetSecretValueAPIClient, secretId string, optFns ...func(*endec.SecretsManagerEndecOptions)) KeyProvider {
	return func(m *lev) {
		m.getKey = endec.NewSecretsManagerEndec(client, secretId, optFns...).GetSecretBinary
	}
}

// WithKeyFromLambdaExtensionSecrets retrieves key from AWS Parameters and Secrets Lambda Extension
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html) using the default client
// [lambda.DefaultParameterSecretsExtensionClient].
func WithKeyFromLambdaExtensionSecrets(secretId string) KeyProvider {
	return WithKeyFromSecretsManager(lambda.DefaultParameterSecretsExtensionClient, secretId)
}

// WithHash can be used to change the hash function.
//
// By default, sha256.New is used.
func WithHash(hashProvider func() hash.Hash) Option {
	return func(l *lev) {
		l.hashProvider = hashProvider
	}
}

func (hasher lev) Sign(ctx context.Context, payload []byte, nonceSize byte) ([]byte, error) {
	var nonce []byte
	if nonceSize > 0 {
		nonce = make([]byte, nonceSize)
		if err := hasher.rand(nonce); err != nil {
			return nil, err
		}

	}

	return hasher.hash(ctx, bytes.NewReader(payload), nonce, nil)
}

func (hasher lev) hash(ctx context.Context, payload io.Reader, nonce []byte, versionId *string) ([]byte, error) {
	key, versionId, err := hasher.getKey(ctx, versionId)
	if err != nil {
		return nil, err
	}

	// token is essentially TLV (https://en.wikipedia.org/wiki/Type%E2%80%93length%E2%80%93value):
	// 0x01 is versionId
	// 0x02 is nonce
	// 0x00 indicates payload with arbitrary length. this must be the last component.
	var b bytes.Buffer

	if versionId != nil {
		v := *versionId
		b.WriteByte(0x01)
		b.WriteByte(byte(len(v)))
		b.Write([]byte(v))
	}

	if n := len(nonce); n > 0 {
		b.WriteByte(0x02)
		b.WriteByte(byte(n))
		b.Write(nonce)
	}

	w := hmac.New(hasher.hashProvider, key)
	if _, err = io.Copy(w, payload); err != nil {
		return nil, err
	}
	if _, err = w.Write(nonce); err != nil {
		return nil, err
	}

	b.WriteByte(0x00)
	b.Write(w.Sum(nil))
	return b.Bytes(), nil
}

func (hasher lev) Verify(ctx context.Context, signature, payload []byte) (ok bool, err error) {
	key, nonce, expected, err := hasher.unpack(ctx, signature)
	if err != nil {
		return false, err
	}

	w := hmac.New(hasher.hashProvider, key)
	w.Write(payload)
	w.Write(nonce)
	actual := w.Sum(nil)
	return subtle.ConstantTimeCompare(expected, actual) == 1, nil
}

func (hasher lev) unpack(ctx context.Context, rawPayload []byte) (key, nonce, payload []byte, err error) {
	var (
		versionId  *string
		code, size byte
	)

	// token is TLV-encoded (https://en.wikipedia.org/wiki/Type%E2%80%93length%E2%80%93value):
	// 0x01 is versionId
	// 0x02 is nonce
	// 0x00 indicates payload with all remaining bytes; this must be the last component.

ingBad:
	for b := bytes.NewBuffer(rawPayload); err == nil; {
		if code, err = b.ReadByte(); err != nil {
			break
		}

		switch code {
		case 0x00:
			payload = b.Bytes()
			break ingBad
		case 0x01:
			if size, err = b.ReadByte(); err == nil {
				data := make([]byte, size)
				if _, err = b.Read(data); err == nil {
					v := string(data)
					versionId = &v
				}
			}
		case 0x02:
			if size, err = b.ReadByte(); err == nil {
				nonce = make([]byte, size)
				_, err = b.Read(nonce)
			}
		}
	}

	switch {
	case err == io.EOF:
		err = fmt.Errorf("token ends too soon")
	case err != nil:
	default:
		key, _, err = hasher.getKey(ctx, versionId)
	}

	return
}

// defaultRand fills the given slice with random data from [rand.Reader].
func defaultRand(dst []byte) error {
	_, err := io.ReadFull(rand.Reader, dst)
	return err
}
