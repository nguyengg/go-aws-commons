// Package configcache provides a centralised place to retrieve default aws.Config for creating AWS clients.
//
// The main method should explicitly create and cache a config with LoadDefaultConfig, LoadSharedConfigProfile, or
// LoadConfig. Whenever an aws.Config instance is needed, call Get or MustGet. Most libraries in go-aws-commons that can
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
	cfg  aws.Config
	err  error
	set  bool
	lock sync.Mutex
)

// Get returns the current [aws.Config] and any error from creating it.
//
// If the cache has no config, config.LoadDefaultConfig will be used to create and cache one.
//
// The optFns argument modifies the aws.Config after it has been shallow-copied with [aws.Config.Copy]. As a result,
// those changes should not persist to the globally cached aws.Config in most cases. If you need to modify the globally
// cached instance, use LoadDefaultConfig, Profile, or AssumeRole.
func Get(ctx context.Context, optFns ...func(*aws.Config)) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	if !set {
		cfg, err = config.LoadDefaultConfig(ctx)
		set = true
	}

	if err == nil && len(optFns) > 0 {
		cfg = cfg.Copy()
		for _, fn := range optFns {
			fn(&cfg)
		}
	}

	return cfg, err
}

// MustGet is a panicky variant of Get.
func MustGet(ctx context.Context, optFns ...func(*aws.Config)) aws.Config {
	lock.Lock()
	defer lock.Unlock()

	if !set {
		cfg, err = config.LoadDefaultConfig(ctx)
		set = true
	}

	if err != nil {
		panic(err)
	}

	if len(optFns) > 0 {
		cfg = cfg.Copy()
		for _, fn := range optFns {
			fn(&cfg)
		}
	}

	return cfg
}

// LoadDefaultConfig creates, caches, and returns a new aws.Config instance created with config.LoadDefaultConfig.
//
// The optFns argument modifies the aws.Config prior to caching it.
func LoadDefaultConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	cfg, err = config.LoadDefaultConfig(ctx, optFns...)
	set = true
	return cfg, err
}

// LoadSharedConfigProfile creates, caches, and returns a new aws.Config with its [aws.Config.SharedConfigProfile] set
// to the given profile.
//
// It does this by attaching config.WithSharedConfigProfile as the last optFn argument; the optFns argument modifies the
// aws.Config prior to caching it.
func LoadSharedConfigProfile(ctx context.Context, profile string, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	cfg, err = config.LoadDefaultConfig(ctx, append(optFns, config.WithSharedConfigProfile(profile))...)
	set = true
	return cfg, err
}

// LoadConfig caches the given aws.Config for later usage via Get, MustGet, AssumeRole.
func LoadConfig(c aws.Config) {
	lock.Lock()
	defer lock.Unlock()

	cfg = c
	err = nil
}

// AssumeRole creates and returns a new aws.Config instance that assumes the given role.
//
// If the cache has no aws.Config instance prior to this call, a default instance will be created and cached with
// config.LoadDefaultConfig. This cached instance will provide the STS client to assume the specified role. If you need
// to configure the original STS client, explicitly call LoadDefaultConfig first.
//
// Subsequent Get will return the cached aws.Config instance, not the instance returned by AssumeRole. If you want all
// subsequent Get to use the same instance returned by AssumeRole, call LoadConfig.
func AssumeRole(ctx context.Context, roleArn string, optFns ...func(*stscreds.AssumeRoleOptions)) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	if !set {
		cfg, err = config.LoadDefaultConfig(ctx)
		set = true
	}

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/credentials/stscreds#hdr-Assume_Role
	cfg.Credentials = stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), roleArn, optFns...)
	return cfg, err
}
