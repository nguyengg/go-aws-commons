package endec

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGetParameterAPIClient struct {
	mock.Mock
}

func (m *mockGetParameterAPIClient) GetParameter(ctx context.Context, input *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	args := m.Called(ctx, input, optFns)
	return args.Get(0).(*ssm.GetParameterOutput), args.Error(1)
}

func TestParameterStoreEndec_WithSecretsManager(t *testing.T) {
	key := base64.RawStdEncoding.EncodeToString([]byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e"))
	plaintext := []byte("hello, world!")
	name := "my-parameter"
	var version int64 = 123456

	ctx := context.Background()

	client := &mockGetParameterAPIClient{}
	client.
		// the first invocation will have name == "name".
		On("GetParameter", ctx, mock.MatchedBy(func(input *ssm.GetParameterInput) bool {
			return *input.Name == name
		}), mock.FunctionalOptions()).
		Return(&ssm.GetParameterOutput{Parameter: &types.Parameter{Value: &key, Version: version}}, nil).
		Once()
	client.
		// the second invocation will have name == "name:version".
		On("GetParameter", ctx, mock.MatchedBy(func(input *ssm.GetParameterInput) bool {
			return *input.Name == fmt.Sprintf("%s:%d", name, version)
		}), mock.FunctionalOptions()).
		Return(&ssm.GetParameterOutput{Parameter: &types.Parameter{Value: &key}}, nil).
		Once()

	c := NewParameterStoreEndec(client, name)

	ciphertext, err := c.Encode(context.Background(), plaintext)
	assert.NoErrorf(t, err, "Encode() error = %v", err)

	got, err := c.Decode(context.Background(), ciphertext)
	assert.NoErrorf(t, err, "Decode() error = %v", err)

	assert.Equalf(t, plaintext, got, "want = %v, got = %v", plaintext, got)

	client.AssertNumberOfCalls(t, "GetParameter", 2)
}
