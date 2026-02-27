package configcache

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
	"github.com/nguyengg/go-aws-commons/metrics"
)

// AddHook calls configcache.AddHook to make sure all the cached config will have the metrics middleware enabled.
func AddHook() {
	configcache.AddHook(func(cfg *aws.Config) {
		cfg.APIOptions = append(cfg.APIOptions, metrics.ClientSideMetricsMiddleware())
	})
}
