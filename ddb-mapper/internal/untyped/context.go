package untyped

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
)

// Options contains the fields that are common to all options.
type Options struct {
	Client         *dynamodb.Client
	Encoder        *attributevalue.Encoder
	Decoder        *attributevalue.Decoder
	VersionUpdater func(item any)
}

// Context embeds both context.Context and Options.
type Context struct {
	context.Context
	Options
}

// Merge merges the given options into the returned context, and makes sure all fields that have default values will be
// non-nil.
func Merge(parent context.Context, opts ...Options) (*Context, error) {
	c := &Context{Context: parent}
	for _, opt := range opts {
		if opt.Client != nil {
			c.Client = opt.Client
		}
		if opt.Encoder != nil {
			c.Encoder = opt.Encoder
		}
		if opt.Decoder != nil {
			c.Decoder = opt.Decoder
		}
		if opt.VersionUpdater != nil {
			c.VersionUpdater = opt.VersionUpdater
		}
	}

	if err := c.init(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Context) init() error {
	if c.Client == nil {
		cfg, err := configcache.Get(c)
		if err != nil {
			return err
		}

		c.Client = dynamodb.NewFromConfig(cfg)
	}

	if c.Encoder == nil {
		c.Encoder = attributevalue.NewEncoder()
	}

	if c.Decoder == nil {
		c.Decoder = attributevalue.NewDecoder()
	}

	return nil
}

// DefaultContext creates a context that has no client.
func DefaultContext() *Context {
	return &Context{
		Context: context.Background(),
		Options: Options{
			Client:  nil,
			Encoder: attributevalue.NewEncoder(),
			Decoder: attributevalue.NewDecoder(),
		},
	}
}
