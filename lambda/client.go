package lambda

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

// GetParameterClient abstracts the GetParameter API that has an implementation using AWS Parameter and Secrets Lambda
// extension (ParameterSecretsExtensionClient).
type GetParameterClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// GetSecretValueClient abstracts the GetSecretValue API that has an implementation using AWS Parameter and Secrets
// Lambda extension (ParameterSecretsExtensionClient).
type GetSecretValueClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// ParameterSecretsExtensionClient implements both GetParameterClient and GetSecretValueClient using the
// AWS Parameter and Secrets Lambda extension.
//
// See https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html and
// https://docs.aws.amazon.com/systems-manager/latest/userguide/ps-integration-lambda-extensions.html.
//
// The zero-value DefaultParameterSecretsExtensionClient is ready for use.
type ParameterSecretsExtensionClient struct {
	// Client is the HTTP client to use for making HTTP requests.
	//
	// If nil, http.DefaultClient is used.
	Client *http.Client

	init sync.Once
}

// DefaultParameterSecretsExtensionClient is the client used by package-level GetSecretValue and GetParameter.
var DefaultParameterSecretsExtensionClient = &ParameterSecretsExtensionClient{Client: http.DefaultClient}

// GetSecretValue is a wrapper around [DefaultClient.GetSecretValue].
func GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	return DefaultParameterSecretsExtensionClient.GetSecretValue(ctx, input)
}

func (l *ParameterSecretsExtensionClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	l.init.Do(l.initFn)

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

	res, err := l.Client.Do(req.WithContext(ctx))
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

// GetParameter is a wrapper around [DefaultClient.GetParameter].
func GetParameter(ctx context.Context, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return DefaultParameterSecretsExtensionClient.GetParameter(ctx, input)
}

func (l *ParameterSecretsExtensionClient) GetParameter(ctx context.Context, input *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	l.init.Do(l.initFn)

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

	res, err := l.Client.Do(req.WithContext(ctx))
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

func (l *ParameterSecretsExtensionClient) initFn() {
	if l.Client == nil {
		l.Client = http.DefaultClient
	}
}
