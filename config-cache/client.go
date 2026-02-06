package configcache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// NewClient is a convenient function to create a client.
//
// Usage example:
//
//	client, err := configcache.NewClient(context.Background(), s3.NewFromConfig)
func NewClient[Client any, Options any](ctx context.Context, fn func(aws.Config, ...func(Options)) *Client, optFns ...func(*aws.Config)) (*Client, error) {
	if cfg, err := Get(ctx, optFns...); err != nil {
		return nil, err
	} else {
		return fn(cfg), nil
	}
}

// MustNewClient is a convenient function to create a client.
//
// Usage example:
//
//	client := configcache.MustNewClient(context.Background(), s3.NewFromConfig)
func MustNewClient[Client any, Options any](ctx context.Context, fn func(aws.Config, ...func(Options)) *Client, optFns ...func(*aws.Config)) *Client {
	return fn(MustGet(ctx, optFns...))
}
