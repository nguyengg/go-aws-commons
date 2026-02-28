# Lambda handler wrappers with sensible defaults and metrics integration

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/lambda.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/lambda)

## Convenient handler wrappers

The various `StartABC` functions wrap your Lambda handler so that a `Metrics` instance (from
[`github.com/nguyengg/go-aws-commons/metrics`](../metrics)) is available from context and will be logged with sensible
default metrics (start and end time, latency, fault, etc.) upon return of your Lambda handler.

```go
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nguyengg/go-aws-commons/lambda"
	"github.com/nguyengg/go-aws-commons/metrics"
)

func main() {
	// start Lambda loop like this.
	lambda.StartHandlerFunc(func(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
		m := metrics.Get(ctx)
		m.AddCounter("userDidSomethingCool", 1)

		// when your handler returns, the Metrics instance will be logged to standard error stream.
		// see https://pkg.go.dev/github.com/nguyengg/go-aws-commons/metrics for examples of what is logged.
		return events.DynamoDBEventResponse{}, nil
	})

	// there are custom wrappers for SQS messages and DynamoDB stream events with batch item failure reporting.
	lambda.StartSQSMessageHandler(func(ctx context.Context, message events.SQSMessage) error {
		// if a non-nil error is returned, only that message should be retried.
		// see https://docs.aws.amazon.com/lambda/latest/dg/example_serverless_SQS_Lambda_batch_item_failures_section.html
		return nil
	})
	lambda.StartDynamodbEventHandler(func(ctx context.Context, record events.DynamoDBEventRecord) error {
		// similarly, if a non-nil error is returned, only that record should be retried.
		// see https://docs.aws.amazon.com/lambda/latest/dg/example_serverless_DynamoDB_Lambda_batch_item_failures_section.html
		return nil
	})
}

```

## Gin adapter for Function URL

A Gin adapter for API Gateway V1 and V2 are already available from [github.com/awslabs/aws-lambda-go-api-proxy](https://pkg.go.dev/github.com/awslabs/aws-lambda-go-api-proxy).
The [function-url](function-url) module (`functionurl`) provides an adapter specifically for Function URL events with
both BUFFERED (which, technically, is no different from API Gateway V2/HTTP events) and RESPONSE_STREAM mode
which uses [`github.com/aws/aws-lambda-go/lambdaurl`](https://pkg.go.dev/github.com/aws/aws-lambda-go/lambdaurl) under the hood.

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	functionurl "github.com/nguyengg/go-aws-commons/lambda/function-url"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.Render(http.StatusOK, render.String{
			Format: "hello, world!",
		})
	})

	// start the Lambda handler either in BUFFERED or STREAM_RESPONSE mode.
	functionurl.StartBuffered(r)
	functionurl.StartStream(r)
}

```

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
	. "github.com/nguyengg/go-aws-commons/must"
)

func main() {
	// lambda.ParameterSecretsExtensionClient implements GetSecretValue and GetParameter so I can substitute the
	// client to any code that needs it. The zero-value struct is ready for use so might as well use the package-level
	// methods unless you want to customise the http.Client.
	c := lambda.ParameterSecretsExtensionClient{}

	// in my Lambda handler, instead of invoking Secrets Manager SDK directly, I can use the client from the
	// extension package which makes use of the AWS Parameter and Secrets Lambda extension.
	getSecretValueOutput := Must(c.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String("my-secret"),
		VersionId:    nil,
		VersionStage: nil,
	}))

	// I can also use the package-level methods which will use the default client.
	getParameterOutput := Must(lambda.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           aws.String("my-parameter"),
		WithDecryption: nil,
	}))
}

```

[getenv](getenv) adds abstraction on top of this so that I can easily swap out how the variable is retrieved.

```go
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/nguyengg/go-aws-commons/lambda/getenv"
)

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
