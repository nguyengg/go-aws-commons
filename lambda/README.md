# Lambda handler wrappers with sensible defaults

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/lambda.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/lambda)

## GetParameter and GetSecretValue using AWS Parameter and Secrets Lambda extension

When running in Lambda, if you need to retrieve parameters from Parameter Store or secrets from Secrets Manager, you can
use the AWS Parameter and Secrets Lambda extension to cache the values. The extension was first described in detail in
blog post https://aws.amazon.com/blogs/compute/using-the-aws-parameter-and-secrets-lambda-extension-to-cache-parameters-and-secrets/.

```go
package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/nguyengg/go-aws-commons/lambda"
)

func main() {
	// lambda.ParameterSecretsExtensionClient implements GetSecretValue and GetParameter so I can substitute the
	// client to any code that needs it. the zero-value struct is ready for use.
	c := lambda.ParameterSecretsExtensionClient{}

	// in my Lambda handler, instead of invoking Secrets Manager SDK directly, I can use the client from the
	// extension package which makes use of the AWS Parameter and Secrets Lambda extension.
	_, err := c.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String("my-secret"),
		VersionId:    nil,
		VersionStage: nil,
	})

	// I can also use the package-level methods which will use the default client.
	_, err = lambda.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           aws.String("my-parameter"),
		WithDecryption: nil,
	})
}

```

[getenv](getenv) adds abstraction on top of this so that I can easily swap out how the variable is retrieved.

```go
func main() {
   // while prototyping, you can retrieve from environment variable
   v := getenv.Env("TEST")

   // now you want to retrieve from Parameter Store instead
   v = getenv.ParameterString(&ssm.GetParameterInput{
      Name:           aws.String("my-parameter-name"),
      WithDecryption: aws.Bool(true),
   })

   // in the next example, the key is retrieved and then used as secret key for HMAC verification.
   key := getenv.SecretBinary(&secretsmanager.GetSecretValueInput{
      SecretId:     aws.String("my-secret-id"),
      VersionId:    nil,
      VersionStage: nil,
   })
   h := hmac.New(sha256.New, key.MustGetWithContext(context.Background()))
   h.Write( /* some data */ )
   h.Sum(nil)
}

```
