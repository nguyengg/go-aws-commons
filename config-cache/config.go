// Package configcache provides a singleton cache to retrieve a default aws.Config instance for creating AWS SDK clients.
//
// The main method should explicitly create and cache a config with LoadDefaultConfig, LoadSharedConfigProfile, Set, or
// Update. Whenever an aws.Config instance is needed, call Get or MustGet. Most libraries in go-aws-commons that can
// create a default SDK client will use Get to do so.
package configcache

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	cfg   aws.Config
	err   error
	set   bool
	hooks []func(*aws.Config)
	lock  sync.Mutex
)

// Get returns the cached config or any error from creating it.
//
// If the cache has no config, config.LoadDefaultConfig will be used to create and cache one.
func Get(ctx context.Context) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	if !set {
		if cfg, err = config.LoadDefaultConfig(ctx); err == nil {
			applyHooks(&cfg)
		}
		set = true
	}

	return cfg, err
}

// MustGet is a panicky variant of Get.
func MustGet(ctx context.Context) aws.Config {
	cfg, err := Get(ctx)
	if err != nil {
		panic(err)
	}

	return cfg
}

// Set changes the cached config to the given instance.
func Set(c aws.Config) {
	lock.Lock()
	defer lock.Unlock()

	cfg, err, set = c, nil, true
	applyHooks(&cfg)
}

// Update modifies the cached config.
//
// Similar to Get, if the cache has no config, config.LoadDefaultConfig will be used to create one. The updated config
// is returned.
func Update(ctx context.Context, fn func(*aws.Config)) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	if !set {
		cfg, err = config.LoadDefaultConfig(ctx)
		set = true
	}

	if err == nil {
		fn(&cfg)
		applyHooks(&cfg)
	}

	return cfg, err
}

// MustUpdate is a panicky variant of Update.
func MustUpdate(ctx context.Context, fn func(*aws.Config)) aws.Config {
	cfg, err := Update(ctx, fn)
	if err != nil {
		panic(err)
	}

	return cfg
}

// LoadDefaultConfig uses config.LoadDefaultConfig to set the cached config.
//
// The optFns argument modifies the config prior to caching it.
func LoadDefaultConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	if cfg, err = config.LoadDefaultConfig(ctx, optFns...); err == nil {
		applyHooks(&cfg)
	}
	set = true

	return cfg, err
}

// LoadSharedConfigProfile uses config.LoadDefaultConfig to set the cached config to using shared config profile.
//
// Specifically, config.WithSharedConfigProfile is appended to the optFn argument.
func LoadSharedConfigProfile(ctx context.Context, profile string, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return LoadDefaultConfig(ctx, append(optFns, config.WithSharedConfigProfile(profile))...)
}

// AssumeRole creates a derived config that assumes the given role.
//
// Similar to Get, if the cache has no config, config.LoadDefaultConfig will be used to create one. This cached config
// is used to create the original STS client to assume the specified role. If you need to configure this STS client,
// explicitly call LoadDefaultConfig, LoadSharedConfigProfile, Set, or Update first.
//
// This method does not update the cache otherwise.
func AssumeRole(ctx context.Context, roleArn string, optFns ...func(*stscreds.AssumeRoleOptions)) (aws.Config, error) {
	cfg, err := Get(ctx)
	if err != nil {
		return cfg, err
	}

	WithAssumeRole(roleArn, optFns...)(&cfg)
	return cfg, nil
}

// AddHook adds a hook to the cache.
//
// If the cache already has a config, fn is applied right away to the cached config similar to Update. If the cache does
// not have a config, or if a new config is being cached (via Set, Update, LoadDefaultConfig, or
// LoadSharedConfigProfile), the hooks will be applied on the config prior to caching.
//
// Hooks are useful when you have modifiers you want to apply to the aws.Config instances regardless of how they were
// created. A good hook is to add metrics.NewClientSideMetrics for example.
func AddHook(hook func(cfg *aws.Config)) {
	lock.Lock()
	defer lock.Unlock()

	hooks = append(hooks, hook)
}

// WithAssumeRole creates a hook to modify the config to assume a specific role.
func WithAssumeRole(roleArn string, optFns ...func(*stscreds.AssumeRoleOptions)) func(cfg *aws.Config) {
	return func(cfg *aws.Config) {
		// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/credentials/stscreds#hdr-Assume_Role
		cfg.Credentials = aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg.Copy()), roleArn, optFns...))
	}
}

func applyHooks(cfg *aws.Config) {
	for _, fn := range hooks {
		fn(cfg)
	}
}
