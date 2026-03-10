// Package keys defines [Provider] which is the common interface to retrieve a secret key []byte from several sources.
//   - [Static] and [FromEnv] are appropriate for testing environments.
//   - [FromSecretsManager] is a good production option and will get you secret rotation for free.
//   - [FromParameterStore] if you're using AWS Systems Manager Parameter Store.
//   - If running in Lambda with [AWS Parameters and Secrets Lambda Extension] enabled, consider
//     [FromLambdaExtensionSecrets] or [FromLambdaExtensionParameter] as well.
//
// [AWS Parameters and Secrets Lambda Extension]: https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html
package keys

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
)

// Provider contains single method Provide.
type Provider interface {
	// Provide returns the secret key []byte along with its optional Id.
	//
	// Users of Provider will save the returned id if non-nil in some ways so that subsequent attempts to use the
	// key will pass the same Id. For AWS Secrets Manager, this is the version Id of the secret.
	Provide(context.Context, *string) (key []byte, id *string, err error)
}

// ProviderFunc is [Provider] as a function.
type ProviderFunc func(ctx context.Context, _ *string) (key []byte, id *string, err error)

func (fn ProviderFunc) Provide(ctx context.Context, id *string) ([]byte, *string, error) {
	return fn(ctx, id)
}

// FromEnv converts the environment variable with the given name to a secret key.
//
// Missing or empty environment variables will result in a non-nil error. The value will be decoded in this order:
//  1. base64.StdEncoding.DecodeString
//  2. hex.DecodeString
func FromEnv(name string) Provider {
	return ProviderFunc(func(ctx context.Context, _ *string) ([]byte, *string, error) {
		v, ok := os.LookupEnv(name)
		if !ok {
			return nil, nil, fmt.Errorf("missing environment variable %s", name)
		}
		if v == "" {
			return nil, nil, fmt.Errorf("empty environment variable %s", name)
		}

		key, err := decodeString(v)
		return key, nil, err
	})
}

// Static uses the given key argument as the secret.
func Static(key []byte) Provider {
	return ProviderFunc(func(ctx context.Context, _ *string) ([]byte, *string, error) {
		return key, nil, nil
	})
}

func decodeString(v string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return hex.DecodeString(v)
	}

	return data, nil
}
