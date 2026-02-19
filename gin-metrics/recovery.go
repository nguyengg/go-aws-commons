package ginmetrics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/metrics"
)

// WithRecovery is a replacement for gin.Recovery.
func WithRecovery() func(cfg *LoggerConfig) {
	return func(cfg *LoggerConfig) {
		cfg.recovery = true
	}
}

// RecoveryFunc is a custom recovery function that increases metrics.Metrics's "panicked" counter by 1.
//
// Intended to be used with gin.CustomRecovery; this middleware should go after Logger since it requires metrics.Metrics
// from context. WithRecovery is intended to be a replacement.
var RecoveryFunc gin.RecoveryFunc = func(c *gin.Context, err any) {
	m := metrics.Get(c)
	m.AddCounter("panicked", 1).Any("error", err)
	c.AbortWithStatus(http.StatusInternalServerError)
}
