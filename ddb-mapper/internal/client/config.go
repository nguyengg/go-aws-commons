package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/ddb/config"
)

// initConfig makes sure all the nil fields in [config.Config] that have default values are non-nil.
func initConfig(ctx context.Context, c *config.Config) error {
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
