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
// Default to using [config.LoadDefaultConfig] if no [aws.Config] instance has been cached.
func Get(ctx context.Context) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	if set {
		return cfg, err
	}

	cfg, err = LoadDefaultConfig(ctx)
	set = true
	return cfg, err
}

// MustGet is a panicky variant of Get.
func MustGet(ctx context.Context) aws.Config {
	lock.Lock()
	defer lock.Unlock()

	if set {
		if err != nil {
			panic(err)
		}
		return cfg
	}

	if cfg, err = config.LoadDefaultConfig(ctx); err != nil {
		panic(err)
	}
	set = true
	return cfg
}

// LoadDefaultConfig create and caches a new [aws.Config] instance from [config.LoadDefaultConfig].
func LoadDefaultConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	cfg, err = config.LoadDefaultConfig(ctx, optFns...)
	set = true
	return cfg, err
}

// Profile will attach [config.WithSharedConfigProfile] as the last optFn argument, equivalent to having set AWS_PROFILE
// environment variable.
func Profile(ctx context.Context, profile string, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	cfg, err = config.LoadDefaultConfig(ctx, append(optFns, config.WithSharedConfigProfile(profile))...)
	set = true
	return cfg, err
}

// AssumeRole will create and store a new [aws.Config] instance that assumes the given role.
//
// If the cache has no [aws.Config] instance prior to this call, a default instance will be created with
// [config.LoadDefaultConfig]. The prior instance will provide the STS client to assume the specified role. If you need
// to configure the original STS client, explicitly call LoadDefaultConfig first.
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
