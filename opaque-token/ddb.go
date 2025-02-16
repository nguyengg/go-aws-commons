package token

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/go-aws-commons/lambda"
	"github.com/nguyengg/go-aws-commons/opaque-token/endec"
)

// DynamoDBKeyConverter converts from DynamoDB's last evaluated key to pagination token and vice versa, intended to be
// used for query and scan operations.
//
// Per specifications, only three data types (S, N, or B) can be partition key or sort key. The pagination token will
// be the DynamoDB JSON blob of the evaluated key, which should have no more than 2 entries.
//
// The zero value struct is ready for use which will encode/decode keys without any encryption.
// Prefer NewDynamoDBKeyConverter instead which provides ways to encrypt/decrypt the token, making it the token opaque.
type DynamoDBKeyConverter struct {
	// Endec controls how the tokens are encrypted/decrypted.
	//
	// By default, there is no encryption. Prefer NewDynamoDBKeyConverter instead.
	Endec endec.Endec

	// EncodeToString controls how the decrypted binary token is encoded to string.
	//
	// If Endec is not nil, [base64.RawURLEncoding.EncodeToString] will be used as the default EncodeToString.
	// If Endec is nil, EncodeToString is used only if EncodeToString is non-nil.
	EncodeToString func([]byte) string

	// DecodeString controls how the encrypted string token is decoded.
	//
	// If Endec is not nil, [base64.RawURLEncoding.DecodeString] will be used as the default DecodeString.
	// If Endec is nil, DecodeString is used only if DecodeString is non-nil.
	DecodeString func(string) ([]byte, error)
}

// EncodeKey encodes the given last evaluated key to an opaque token.
func (c DynamoDBKeyConverter) EncodeKey(ctx context.Context, key map[string]types.AttributeValue) (string, error) {
	switch n := len(key); n {
	case 1, 2:
	default:
		return "", fmt.Errorf("invalid number of attributes in key: expected 1 or 2, got (%d)", n)
	}

	item := make(map[string]map[string]string)
	for k, v := range key {
		item[k] = make(map[string]string)

		avS, ok := v.(*types.AttributeValueMemberS)
		if ok {
			item[k]["S"] = avS.Value
			continue
		}

		avN, ok := v.(*types.AttributeValueMemberN)
		if ok {
			item[k]["N"] = avN.Value
			continue
		}

		avB, ok := v.(*types.AttributeValueMemberB)
		if ok {
			item[k]["B"] = base64.RawStdEncoding.EncodeToString(avB.Value)
			continue
		}

		return "", fmt.Errorf("key named %s has unknown type %T", k, v)
	}

	plaintext, err := json.Marshal(item)
	if err != nil {
		return "", fmt.Errorf("marshal token as JSON error: %w", err)
	}

	if c.Endec == nil {
		if f := c.EncodeToString; f == nil {
			return string(plaintext), nil
		} else {
			return f(plaintext), nil
		}
	}

	ciphertext, err := c.Endec.Encode(ctx, plaintext)
	if err != nil {
		return "", nil
	}

	if c.EncodeToString == nil {
		c.EncodeToString = base64.RawURLEncoding.EncodeToString
	}
	return c.EncodeToString(ciphertext), nil
}

// DecodeToken decodes the given opaque token to an exclusive start key.
func (c DynamoDBKeyConverter) DecodeToken(ctx context.Context, token string) (key map[string]types.AttributeValue, err error) {
	plaintext := []byte(token)

	if c.DecodeString != nil {
		plaintext, err = c.DecodeString(token)
		if err != nil {
			return nil, err
		}
	}

	if c.Endec != nil {
		plaintext, err = c.Endec.Decode(ctx, plaintext)
		if err != nil {
			return nil, err
		}
	}

	item := make(map[string]map[string]string)
	if err = json.Unmarshal(plaintext, &item); err != nil {
		return nil, fmt.Errorf("unmarshal token as JSON error: %w", err)
	}

	switch n := len(item); n {
	case 1, 2:
	default:
		return nil, fmt.Errorf("invalid number of attributes in key: expected 1 or 2, got (%d)", n)
	}

	key = make(map[string]types.AttributeValue)
	for k, v := range item {
		if n := len(v); n != 1 {
			return nil, fmt.Errorf("invalid number of attributes in key named %s: expected 1 or 2, got (%d)", k, n)
		}

		avS, ok := v["S"]
		if ok {
			key[k] = &types.AttributeValueMemberS{Value: avS}
			continue
		}

		avN, ok := v["N"]
		if ok {
			key[k] = &types.AttributeValueMemberN{Value: avN}
			continue
		}

		avB, ok := v["B"]
		if ok {
			data, err := base64.RawStdEncoding.DecodeString(avB)
			if err != nil {
				return nil, fmt.Errorf("decode attribute named %s as B error: %w", k, err)
			}

			key[k] = &types.AttributeValueMemberB{Value: data}
			continue
		}
	}

	return key, nil
}

// EncryptionOption makes it easy to specify both the secret key and the encryption algorithm in a user-friendly manner.
type EncryptionOption func(*options) error

// options is not exported to prevent users from accidentally writing their own EncryptionOption.
type options struct {
	c *DynamoDBKeyConverter
}

// NewDynamoDBKeyConverter returns a new DynamoDBKeyConverter that uses encryption/decryption to produce opaque tokens.
//
// If you have static key, pass WithAES or WithChaCha20Poly1305.
// If you want to retrieve secret binary from AWS Secrets Hasher, pass WithKeyFromSecretsManager.
// If you are running in AWS Lambda with AWS Parameters and Secrets Lambda Extension
// (https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html) enabled, pass
// WithKeyFromLambdaExtensionSecrets.
func NewDynamoDBKeyConverter(opt EncryptionOption, optFns ...func(*DynamoDBKeyConverter)) (*DynamoDBKeyConverter, error) {
	c := &DynamoDBKeyConverter{
		EncodeToString: base64.RawURLEncoding.EncodeToString,
		DecodeString:   base64.RawURLEncoding.DecodeString,
	}

	if err := opt(&options{c}); err != nil {
		return nil, err
	}

	for _, fn := range optFns {
		fn(c)
	}

	return c, nil
}

// WithAES makes the DynamoDBKeyConverter uses WithAES encryption with the given key.
func WithAES(key []byte) EncryptionOption {
	return func(opts *options) (err error) {
		opts.c.Endec, err = endec.NewAESWithKey(key)
		return
	}
}

// WithChaCha20Poly1305 makes the DynamoDBKeyConverter uses ChaCha20-Poly1305 encryption with the given key.
func WithChaCha20Poly1305(key []byte) EncryptionOption {
	return func(opts *options) (err error) {
		opts.c.Endec, err = endec.NewChaCha20Poly1305WithKey(key)
		return
	}
}

// WithKeyFromSecretsManager makes the DynamoDBKeyConverter uses key from AWS Secrets Manager.
//
// If you want to change the encryption suite or customises the [endec.SecretsManagerEndec] further, see
// [endec.SecretsManagerEndecOptions].
func WithKeyFromSecretsManager(client endec.GetSecretValueAPIClient, secretId string, optFns ...func(*endec.SecretsManagerEndecOptions)) EncryptionOption {
	return func(opts *options) error {
		opts.c.Endec = endec.NewSecretsManagerEndec(client, secretId, optFns...)
		return nil
	}
}

// WithKeyFromLambdaExtensionSecrets makes the DynamoDBKeyConverter uses key from AWS Parameters and Secrets Lambda
// Extension (https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html) using the
// default client [lambda.DefaultParameterSecretsExtensionClient].
//
// If you want to change the encryption suite or customises the [endec.SecretsManagerEndec] further, see
// [endec.SecretsManagerEndecOptions].
func WithKeyFromLambdaExtensionSecrets(secretId string, optFns ...func(*endec.SecretsManagerEndecOptions)) EncryptionOption {
	return func(opts *options) error {
		opts.c.Endec = endec.NewSecretsManagerEndec(lambda.DefaultParameterSecretsExtensionClient, secretId, optFns...)
		return nil
	}
}
