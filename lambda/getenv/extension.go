package getenv

import (
	"context"
	"encoding/base64"
	"encoding/hex"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/nguyengg/go-aws-commons/lambda"
)

// ParameterAs creates a Parameter Store Variable of type T that retrieves from the AWS Parameter and Secrets Lambda
// extension.
//
// Function m is passed the Value from the [ssm.GetParameterOutput.Parameter].
//
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html#ps-integration-lambda-extensions-sample-commands.
func ParameterAs[T any](input *ssm.GetParameterInput, m func(*string) (T, error)) Variable[T] {
	c := &lambda.ParameterSecretsExtensionClient{}

	return getter[T](func(ctx context.Context) (res T, err error) {
		output, err := c.GetParameter(ctx, input)
		if err != nil {
			return res, err
		}

		return m(output.Parameter.Value)
	})
}

// ParameterString creates a string Parameter Store Variable that retrieves from the AWS Parameter and Secrets Lambda
// extension.
func ParameterString(input *ssm.GetParameterInput) Variable[string] {
	return ParameterAs[string](input, func(s *string) (string, error) {
		return aws.ToString(s), nil
	})
}

// ParameterBinary creates a ParameterAs Store binary Variable that retrieves from the AWS Parameter and Secrets Lambda
// extension.
func ParameterBinary(input *ssm.GetParameterInput, decodeString func(string) ([]byte, error)) Variable[[]byte] {
	return ParameterAs[[]byte](input, func(s *string) ([]byte, error) {
		if s == nil {
			return nil, nil
		}

		return decodeString(*s)
	})
}

// SecretAs creates a Secrets Manager Variable of type T that retrieves from the AWS Parameter and Secrets Lambda
// extension.
//
// Function m is passed both the SecretBinary and SecretString of [secretsmanager.GetSecretValueOutput].
//
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html#ps-integration-lambda-extensions-sample-commands.
func SecretAs[T any](input *secretsmanager.GetSecretValueInput, m func(secretBinary []byte, secretString *string) (T, error)) Variable[T] {
	c := &lambda.ParameterSecretsExtensionClient{}

	return getter[T](func(ctx context.Context) (res T, err error) {
		output, err := c.GetSecretValue(ctx, input)
		if err != nil {
			return res, err
		}

		return m(output.SecretBinary, output.SecretString)
	})
}

// SecretString creates a string Secrets Manager Variable that retrieves from the AWS Parameter and Secrets Lambda
// extension.
//
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html#ps-integration-lambda-extensions-sample-commands.
func SecretString(input *secretsmanager.GetSecretValueInput) Variable[string] {
	return SecretAs[string](input, func(secretBinary []byte, secretString *string) (string, error) {
		return aws.ToString(secretString), nil
	})
}

// SecretBinary creates a binary Secrets Manager Variable that retrieves from the AWS Parameter and Secrets Lambda
// extension.
//
// If the SecretBinary is not available but the SecretString is, DefaultSecretStringDecoder is used to decode the
// SecretString from [secretsmanager.GetSecretValueOutput].
//
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html#ps-integration-lambda-extensions-sample-commands.
func SecretBinary(input *secretsmanager.GetSecretValueInput) Variable[[]byte] {
	return SecretBinaryWithDecoder(input, DefaultSecretStringDecoder)
}

// SecretBinaryWithDecoder creates a binary Secrets Manager Variable that retrieves from the AWS Parameter and Secrets
// Lambda extension.
//
// If the SecretBinary is not available but the SecretString is, the given decodeString argument is used.
//
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html#ps-integration-lambda-extensions-sample-commands.
func SecretBinaryWithDecoder(input *secretsmanager.GetSecretValueInput, decodeString func(secretString string) ([]byte, error)) Variable[[]byte] {
	return SecretAs[[]byte](input, func(secretBinary []byte, secretString *string) ([]byte, error) {
		if secretBinary != nil || secretString == nil {
			return secretBinary, nil
		}

		return decodeString(*secretString)
	})
}

// DefaultSecretStringDecoder attempts to decode the string first with base64.RawStdEncoding, then with
// hex.DecodeString, and as a fallback, converts the string to []byte.
func DefaultSecretStringDecoder(v string) (data []byte, err error) {
	data, err = base64.RawStdEncoding.DecodeString(v)
	if err != nil {
		data, err = hex.DecodeString(v)
		if err != nil {
			data, err = []byte(v), nil
		}
	}

	return
}
