package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	ini "github.com/nguyengg/init-once"
)

// DefaultClientProvider provides the default DynamoDB client used by package-level methods.
//
// By default, it uses [configcache.Get] to create a client.
var DefaultClientProvider ClientProvider = &defaultClientProvider{}

// ClientProvider defines a single method Provide to return a DynamoDB client.
type ClientProvider interface {
	// Provide returns the DynamoDB client to use.
	Provide(ctx context.Context) (*dynamodb.Client, error)
}

// ClientProviderFunc is the type of [DefaultClientProvider].
type ClientProviderFunc func(ctx context.Context) (*dynamodb.Client, error)

func (fn ClientProviderFunc) Provide(ctx context.Context) (*dynamodb.Client, error) {
	return fn(ctx)
}

// defaultClientProvider's provide is cached when success.
type defaultClientProvider struct {
	c    *dynamodb.Client
	once ini.SuccessOnce
}

func (p *defaultClientProvider) Provide(ctx context.Context) (*dynamodb.Client, error) {
	if err := p.once.Do(func() error {
		cfg, err := configcache.Get(ctx)
		if err != nil {
			return err
		}

		p.c = dynamodb.NewFromConfig(cfg)
		return nil
	}); err != nil {
		return nil, err
	}

	return p.c, nil
}

// StaticClientProvider implements [ClientProvider] for the given client.
//
// Useful if you already have a client for all the package-level methods to use.
//
// Usage:
//
//	cfg, _ := config.LoadDefaultConfig(context.Background())
//	client := dynamodb.NewFromConfig(cfg)
//	ddb.DefaultClientProvider = &StaticClientProvider{Client: client}
type StaticClientProvider struct {
	// Client is returned by Provide.
	Client *dynamodb.Client
}

// Provide returns [StaticClientProvider.Client].
func (p StaticClientProvider) Provide(_ context.Context) (*dynamodb.Client, error) {
	return p.Client, nil
}

// defaultConfig creates a [config.Config] with its [config.Config.Client] being set to DefaultClientProvider's return
// value.
func defaultConfig(ctx context.Context) (cfg config.Config) {
	if c, err := DefaultClientProvider.Provide(ctx); err == nil {
		cfg.Client = c
	}

	return
}
