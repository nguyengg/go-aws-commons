package configcache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// NewClient is a convenient function to create a client.
//
// Usage example:
//
//	client, err := configcache.NewClient(context.Background(), s3.NewFromConfig, s3.WithPresignExpires(time.Hour))
func NewClient[Client any, Options any](ctx context.Context, fn func(aws.Config, ...func(*Options)) *Client, clientOptions ...func(*Options)) (*Client, error) {
	if cfg, err := Get(ctx); err != nil {
		return nil, err
	} else {
		return fn(cfg, clientOptions...), nil
	}
}

// MustNewClient is a panicky variant of NewClient.
//
// Usage example:
//
//	client := configcache.MustNewClient(context.Background(), s3.NewFromConfig, s3.WithPresignExpires(1 * time.Hour))
func MustNewClient[Client any, Options any](ctx context.Context, fn func(aws.Config, ...func(*Options)) *Client, clientOptions ...func(*Options)) *Client {
	return fn(MustGet(ctx), clientOptions...)
}
