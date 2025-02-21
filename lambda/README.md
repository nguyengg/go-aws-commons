# Lambda handler wrappers with sensible defaults

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/lambda.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/lambda)

## Convenient handler wrappers

The various `StartABC` functions wrap your Lambda handler so that a [Metrics](../metrics) instance is available from
context and will be logged with sensible default metrics (start and end time, latency, fault, etc.) upon return of your
Lambda handler.

```go
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nguyengg/go-aws-commons/lambda"
	"github.com/nguyengg/go-aws-commons/metrics"
)

func main() {
	// you can use a specific specialisation for your handler like DynamoDB stream event below.
	lambda.StartDynamoDBEventHandleFunc(func(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
		m := metrics.Ctx(ctx)
		m.IncrementCount("myMetric")
		return events.DynamoDBEventResponse{}, nil
	})

	// or you can use the generic StartHandlerFunc template if there isn't a specialisation.
	lambda.StartHandlerFunc(func(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
		m := metrics.Ctx(ctx)
		m.IncrementCount("myMetric")
		return events.DynamoDBEventResponse{}, nil
	})

	// when your handler returns, the Metrics instance will be logged to standard error stream.
}

```

## Gin adapter for Function URL

A Gin adapter for API Gateway V1 and V2 are already available from https://github.com/awslabs/aws-lambda-go-api-proxy.
This [module](functionurl/gin) provides an adapter specifically for Function URL events with both BUFFERED (which,
technically, is no different from API Gateway V2/HTTP events) and RESPONSE_STREAM mode which uses
https://github.com/aws/aws-lambda-go/tree/main/lambdaurl.

Furthermore, there's a neat middleware that can be used to add authorisation check.

```go
package main

import (
	"github.com/gin-gonic/gin"
	sessions "github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb"
	ginadapter "github.com/nguyengg/go-aws-commons/lambda/functionurl/gin"
	"github.com/nguyengg/go-aws-commons/lambda/functionurl/gin/rules"
)

type Session struct {
	SessionId string `dynamodbav:"sessionId" tableName:"session"`
	User      *User  `dynamodbav:"user,omitempty"`
}

type User struct {
	Sub    string   `dynamodbav:"sub"`
	Groups []string `dynamodbav:"groups,stringset"`
}

func main() {
	r := gin.Default()

	// this example uses github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb to provide session management.
	r.GET("/",
		sessions.Sessions[Session]("sid"),
		ginadapter.RequireGroupMembership(func(c *gin.Context) (authenticated bool, groups rules.Groups) {
			var s *Session = sessions.Get[Session](c)
			if s.User == nil {
				return false, nil
			}
			return true, s.User.Groups
		}, rules.AllOf("a", "b"), rules.OneOf("b", "c")))
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
