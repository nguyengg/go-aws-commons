package endec

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGetSecretValueAPIClient struct {
	mock.Mock
}

func (m *mockGetSecretValueAPIClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(ctx, input, optFns)
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}

func TestSecretsManagerEndec_WithSecretsManager(t *testing.T) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	plaintext := []byte("hello, world!")
	secretId := "my-secret"
	versionId := "123456"

	ctx := context.Background()

	client := &mockGetSecretValueAPIClient{}
	client.
		// the first invocation will have nil for versionId.
		On("GetSecretValue", ctx, mock.MatchedBy(func(input *secretsmanager.GetSecretValueInput) bool {
			assert.Equal(t, secretId, *input.SecretId)
			return input.VersionId == nil
		}), mock.FunctionalOptions()).
		Return(&secretsmanager.GetSecretValueOutput{SecretBinary: key, VersionId: &versionId}, nil).
		Once()
	client.
		// the second invocation will have the expected version Id.
		On("GetSecretValue", ctx, mock.MatchedBy(func(input *secretsmanager.GetSecretValueInput) bool {
			assert.Equal(t, secretId, *input.SecretId)
			return *input.VersionId == versionId
		}), mock.FunctionalOptions()).
		Return(&secretsmanager.GetSecretValueOutput{SecretBinary: key}, nil).
		Once()

	c := NewSecretsManagerEndec(client, secretId)

	ciphertext, err := c.Encode(context.Background(), plaintext)
	assert.NoErrorf(t, err, "Encode() error = %v", err)

	got, err := c.Decode(context.Background(), ciphertext)
	assert.NoErrorf(t, err, "Decode() error = %v", err)

	assert.Equalf(t, plaintext, got, "want = %v, got = %v", plaintext, got)

	client.AssertNumberOfCalls(t, "GetSecretValue", 2)
}
