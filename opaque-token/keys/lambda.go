package keys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// FromLambdaExtensionParameter retrieves from the [Parameter Store component] of the
// [AWS Parameter and Secrets Lambda extension].
//
// [Parameter Store component]: https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html
// [AWS Parameter and Secrets Lambda extension]: https://aws.amazon.com/blogs/compute/using-the-aws-parameter-and-secrets-lambda-extension-to-cache-parameters-and-secrets/
func FromLambdaExtensionParameter(name string, optFns ...func(opts *ParameterStoreOptions)) Provider {
	return FromParameterStore(&ParamSecretsLambdaExtClient{}, name, optFns...)
}

// FromLambdaExtensionSecrets retrieves from the [AWS Secrets Manager component] of the
// [AWS Parameter and Secrets Lambda extension].
//
// [AWS Secrets Manager component]: https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html
// [AWS Parameter and Secrets Lambda extension]: https://aws.amazon.com/blogs/compute/using-the-aws-parameter-and-secrets-lambda-extension-to-cache-parameters-and-secrets/
func FromLambdaExtensionSecrets(secretId string, optFns ...func(*SecretsManagerOptions)) Provider {
	return FromSecretsManager(&ParamSecretsLambdaExtClient{}, secretId, optFns...)
}

// ParamSecretsLambdaExtClient implements both [GetParameterAPIClient] and [GetSecretValueAPIClient] using the
// [AWS Parameter and Secrets Lambda extension].
//
// The zero value is ready for use. You can explicitly construct ParamSecretsLambdaExtClient if you need to modify
// the Client.
//
// [AWS Parameter and Secrets Lambda extension]: https://aws.amazon.com/blogs/compute/using-the-aws-parameter-and-secrets-lambda-extension-to-cache-parameters-and-secrets/
type ParamSecretsLambdaExtClient struct {
	// Client is the HTTP client to use for making HTTP requests to the AWS Parameter and Secrets Lambda extension.
	//
	// If nil, http.DefaultClient is used.
	Client *http.Client

	once sync.Once
}

var _ GetParameterAPIClient = &ParamSecretsLambdaExtClient{}
var _ GetSecretValueAPIClient = &ParamSecretsLambdaExtClient{}

func (c *ParamSecretsLambdaExtClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}

	// https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html
	port := os.Getenv("PARAMETERS_SECRETS_EXTENSION_HTTP_PORT")
	if port == "" {
		port = "2773"
	}

	req, err := http.NewRequest("GET", "http://localhost:"+port+"/secretsmanager/get", nil)
	if err != nil {
		return nil, fmt.Errorf("create GET secrets request error: %w", err)
	}

	token := os.Getenv("AWS_SESSION_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("no AWS_SESSION_TOKEN")
	}

	req.Header.Add("X-Aws-Parameters-Secrets-Token", token)

	q := url.Values{}
	q.Add("secretId", aws.ToString(input.SecretId))
	if input.VersionId != nil {
		q.Add("versionId", *input.VersionId)
	}
	if input.VersionStage != nil {
		q.Add("versionStage", *input.VersionStage)
	}
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("do GET secret error: %w", err)
	}

	output := &secretsmanager.GetSecretValueOutput{}
	err = json.NewDecoder(res.Body).Decode(output)
	_ = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("decode GET secret response error: %w", err)
	}

	return output, nil
}

func (c *ParamSecretsLambdaExtClient) GetParameter(ctx context.Context, input *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}

	// https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html
	port := os.Getenv("PARAMETERS_SECRETS_EXTENSION_HTTP_PORT")
	if port == "" {
		port = "2773"
	}

	req, err := http.NewRequest("GET", "http://localhost:"+port+"/systemsmanager/parameters/get", nil)
	if err != nil {
		return nil, fmt.Errorf("create GET secrets request error: %w", err)
	}

	token := os.Getenv("AWS_SESSION_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("no AWS_SESSION_TOKEN")
	}

	req.Header.Add("X-Aws-Parameters-Secrets-Token", token)

	q := url.Values{}

	// we have to parse the name to see if it has version or label in it.
	if input.Name != nil {
		name := *input.Name
		parts := strings.SplitN(name, ":", 3)
		if len(parts) == 2 {
			q.Add("name", parts[0])
			if _, err = strconv.ParseInt(parts[1], 10, 64); err == nil {
				q.Add("version", parts[1])
			} else {
				q.Add("label", parts[1])
			}
		} else {
			q.Add("name", name)
		}
	}

	if input.WithDecryption != nil {
		q.Add("withDecryption", fmt.Sprintf("%t", *input.WithDecryption))
	}
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("do GET parameter error: %w", err)
	}

	output := &ssm.GetParameterOutput{}
	err = json.NewDecoder(res.Body).Decode(output)
	_ = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("decode GET parameter response error: %w", err)
	}

	return output, nil
}
