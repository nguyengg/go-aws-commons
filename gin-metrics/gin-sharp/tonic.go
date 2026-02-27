// Package ginsharp provides an alternative to gin.Default while also toggling sensible default flags via init.
//
// Specifically, init will call gin.EnableJsonDecoderDisallowUnknownFields while also adding
// metrics.ClientSideMetricsMiddleware hook to the configcache.Get.
package ginsharp

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gin-gonic/gin"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
	ginmetrics "github.com/nguyengg/go-aws-commons/gin-metrics"
	"github.com/nguyengg/go-aws-commons/metrics"
)

// Default creates a gin.Engine with sensible defaults.
func Default() *gin.Engine {
	r := gin.New()
	r.Use(ginmetrics.Logger(ginmetrics.WithRequestId()))
	return r
}

func init() {
	gin.EnableJsonDecoderDisallowUnknownFields()
	configcache.AddHook(func(cfg *aws.Config) {
		cfg.APIOptions = append(cfg.APIOptions, metrics.ClientSideMetricsMiddleware())
	})
}
