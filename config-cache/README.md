# Global AWS Config Cache

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commonds/config-cache.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/config-cache)

Golang prefers zero-value structs to be usable out of the box. As a result, a common pattern is for the structs to
perform one-time initialisation (via `sync.Once`) to set up sensible defaults, including AWS clients like this:

```go
package app

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
)

type MyApp struct {
	DynamoDBClient *dynamodb.Client

	// once guards init.
	once sync.Once
}

func (a *MyApp) init(ctx context.Context) (err error) {
	a.once.Do(func() {
		if a.DynamodbClient == nil {
			var cfg config.Config
			if cfg, err = config.LoadDefaultConfig(context.TODO()); err != nil {
				return
			}

			a.DynamodbClient = dynamodb.NewFromConfig(cfg)
		}
	})
}

var DefaultApp = &MyApp{}

func (a *MyApp) DoSomething(ctx context.Context) error {
	if err := a.init(); err != nil {
		return err
	}

	// do some work
	return nil
}

func DoSomething(ctx context.Context) error {
	return DefaultApp.DoSomething(ctx)
}

func main() {
	// you would use the package-level methods (like http.Get) for convenience.
	app.DoSomething(context.TODO())
}

```

What if you need to assume a different role? What if you have ten structs that use similar pattern to `MyApp`?

This package solves that problem by providing a different way to retrieve the "default" `config.Config` like this:
```go
package app

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/nguyengg/go-aws-commons/config-cache"
)

type MyApp struct {
	DynamoDBClient *dynamodb.Client

	// once guards init.
	once sync.Once
}

func (a *MyApp) init(ctx context.Context) (err error) {
	a.once.Do(func() {
		if a.DynamodbClient == nil {
			var cfg config.Config
			if cfg, err = configcache.Get(context.TODO()); err != nil {
				return
			}

			a.DynamodbClient = dynamodb.NewFromConfig(cfg)
		}
	})
}

func main() {
	// equivalent to having set AWS_PROFILE=my-profile
	configcache.Profile("my-profile")

	// use the current aws.Config instance to assume another role.
	configcache.AssumeRole("my-role-arn")

	// you would use the package-level methods (like http.Get) for convenience.
	app.DoSomething(context.TODO())
}

```
