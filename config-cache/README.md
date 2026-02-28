# Singleton AWS config cache to make using package-level methods easier

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commonds/config-cache.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/config-cache)

Golang prefers zero-value structs to be usable out of the box. As a result, a common pattern is for the structs to
perform one-time initialisation (via `sync.Once`) to set up sensible defaults, including AWS clients like this:

```go
package app

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	// it's nice to the package-level methods (like http.Get) for convenience.
	_ = DoSomething(context.TODO())
}

// DoSomething is just alias to DefaultApp.DoSomething.
func DoSomething(ctx context.Context) error {
	return DefaultApp.DoSomething(ctx)
}

// DefaultApp is the zero-value MyApp.
var DefaultApp = &MyApp{}

type MyApp struct {
	Client *dynamodb.Client

	// once guards init.
	once sync.Once
}

// DoSomething and other public methods must call init to make sure variables are initialised.
func (a *MyApp) DoSomething(ctx context.Context) error {
	if err := a.init(ctx); err != nil {
		return err
	}

	// do some work
	return nil
}

func (a *MyApp) init(ctx context.Context) (err error) {
	a.once.Do(func() {
		if a.Client == nil {
			var cfg aws.Config
			// you can pass cfg into init, but it would be difficult to use package-level methods this way.
			if cfg, err = config.LoadDefaultConfig(context.TODO()); err != nil {
				return
			}

			a.Client = dynamodb.NewFromConfig(cfg)
		}
	})

	return
}


```

What if you need to assume a different role? What if you have ten structs that use similar pattern to `MyApp`?

This package solves that problem by providing a different way to retrieve the "default" `aws.Config` like this:
```go
func main() {
	// if you want to use a specific profile, do this.
	Must(configcache.LoadSharedConfigProfile(context.Background(), "my-profile"))

	// if you want to make sure all configcache users assume a role, do this.
	Must(configcache.Update(context.Background(), func(cfg *aws.Config) {
		configcache.WithAssumeRole("arn:aws:iam::123456789012:role/ApplicationRole")(cfg)
	}))

	// now this package-level method will use configcache.Get to use the right config.
	_ = DoSomething(context.TODO())
}

func (a *MyApp) init(ctx context.Context) (err error) {
	a.once.Do(func() {
		if a.Client == nil {
			var cfg aws.Config
			// instead of config.LoadDefaultConfig, use configcache.Get instead.
			if cfg, err = configcache.Get(context.TODO()); err != nil {
				return
			}

			a.Client = dynamodb.NewFromConfig(cfg)
		}
	})

	return
}

```
