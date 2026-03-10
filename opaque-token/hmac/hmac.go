// Package hmac builds on top of [crypto/hmac] to provide CSRF token generation and validation.
//
// For testing or simple setup, use [keys.Static] with a random 32-byte key, or use [keys.FromEnv] to retrieve from
// environment variables. In product, pass [keys.FromSecretsManager] to [New] to get secret rotation for free. If running
// in Lambda with [AWS Parameters and Secrets Lambda Extension] enabled, [keys.FromLambdaExtensionSecrets] can be used.
//
// [AWS Parameters and Secrets Lambda Extension]: https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html
package hmac

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"hash"
	"io"

	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
)

// Signer has a single method Sign.
type Signer interface {
	// Sign creates an HMAC signature from the given payload.
	//
	// If nonce size is 0, the same payload will always produce the same signature.
	//
	// In order to use the signature as CSRF token, pass a non-zero value for the nonce size (16 is a good length).
	// According to [CSRF Prevention Cheat Sheet], the payload should include the session id and any other
	// information you wish. Do not include a random value in the payload; this method already creates the random
	// value for you from the given nonce size.
	//
	// [CSRF Prevention Cheat Sheet]: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#pseudo-code-for-implementing-hmac-csrf-tokens
	Sign(ctx context.Context, payload []byte, nonceSize byte) ([]byte, error)
}

// Verifier has a single method Verify.
type Verifier interface {
	// Verify validates the given signature against the expected payload.
	//
	// The signature should have been created by a previous call to [Signer.Sign].
	//
	// The boolean return value is true if and only if the signature has passed all validation. When the boolean
	// return value is false and there is no error, the signature passes all parsing but fails at the final
	// comparing step. Otherwise, any parsing error will be returned.
	Verify(ctx context.Context, signature, payload []byte) (bool, error)
}

// Engine implements both [Signer] and [Verifier].
//
// Engine implementations are stateless and are safe to use from concurrent goroutines.
type Engine interface {
	Signer
	Verifier
}

// New creates a new [Engine] with the given [keys.Provider].
//
// See [keys] package for several options to construct an [Engine]:
//   - [keys.Static] and [keys.FromEnv] are appropriate for testing.
//   - [keys.FromSecretsManager] is good in production and gets you secret rotation for free.
//   - Consider [keys.FromLambdaExtensionSecrets] if running in Lambda with [AWS Parameters and Secrets Lambda Extension]
//     enabled.
//
// If you want to use a specific hash function instead of [sha256.New], use [WithHash].
//
// [AWS Parameters and Secrets Lambda Extension]: https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html
func New(keyProvider keys.Provider, optFns ...Option) Engine {
	e := &engine{
		keyProvider:  keyProvider,
		hashProvider: sha256.New,
		rand:         defaultRand,
	}

	for _, fn := range optFns {
		fn(e)
	}

	return e
}

// WithHash can be used to change the hash function.
//
// By default, [sha256.New] is used.
func WithHash(hashProvider func() hash.Hash) Option {
	return func(e *engine) {
		e.hashProvider = hashProvider
	}
}

// Option customises [Engine].
type Option func(e *engine)

// defaultRand fills the given slice with random data from [rand.Reader].
func defaultRand(dst []byte) error {
	_, err := io.ReadFull(rand.Reader, dst)
	return err
}
