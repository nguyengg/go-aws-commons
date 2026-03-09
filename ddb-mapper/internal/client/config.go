package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
)

// Config is clone of [ddb.config.Config] to break cyclic import.
type Config struct {
	Client         *dynamodb.Client
	Encoder        *attributevalue.Encoder
	Decoder        *attributevalue.Decoder
	VersionUpdater func(item any)
}

// init makes sure all the nil fields in [Config] that have default values are non-nil.
func (c *Config) init(ctx context.Context) error {
	if c.Client == nil {
		cfg, err := configcache.Get(ctx)
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
