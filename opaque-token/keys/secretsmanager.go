package keys

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// FromSecretsManager creates a new [Provider] backed by AWS Secrets Manager.
func FromSecretsManager(client GetSecretValueAPIClient, secretId string, optFns ...func(*SecretsManagerOptions)) Provider {
	opts := &SecretsManagerOptions{}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.SecretStringDecoder == nil {
		opts.SecretStringDecoder = decodeString
	}

	return &secretsManagerKeyProvider{
		client:              client,
		secretId:            secretId,
		versionStage:        opts.VersionStage,
		secretStringDecoder: opts.SecretStringDecoder,
	}
}

// GetSecretValueAPIClient abstracts the Secrets Manager API [secretsmanager.Client.GetSecretValue].
type GetSecretValueAPIClient interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// SecretsManagerOptions customises [FromSecretsManager].
type SecretsManagerOptions struct {
	// VersionStage overrides [secretsmanager.GetSecretValueInput.VersionStage].
	VersionStage *string

	// SecretStringDecoder can be used to control how the secret value is decoded into a key.
	//
	// By default, the [secretsmanager.GetSecretValueOutput.SecretBinary] is used as the secret key. If this is not
	// available because the secret was provided as a string instead, this function controls how the
	// [secretsmanager.GetSecretValueOutput.SecretString] is transformed into the secret key. If not given, the
	// default function will try these two in order:
	//  1. base64.RawStdEncoding.DecodeString
	//  2. hex.DecodeString
	SecretStringDecoder func(string) ([]byte, error)
}

type secretsManagerKeyProvider struct {
	client              GetSecretValueAPIClient
	secretId            string
	versionStage        *string
	secretStringDecoder func(string) ([]byte, error)
}

func (s secretsManagerKeyProvider) Provide(ctx context.Context, versionId *string) ([]byte, *string, error) {
	getSecretValueOutput, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     &s.secretId,
		VersionId:    versionId,
		VersionStage: s.versionStage,
	})
	if err != nil {
		return nil, nil, err
	}

	secretBinary, versionId := getSecretValueOutput.SecretBinary, getSecretValueOutput.VersionId
	if v := getSecretValueOutput.SecretString; v != nil {
		if secretBinary, err = s.secretStringDecoder(*v); err != nil {
			return nil, nil, err
		}
	}

	return secretBinary, versionId, nil
}
