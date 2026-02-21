package ginmetrics

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/metrics"
	"github.com/rotisserie/eris"
)

// Logger is a replacement for gin.Logger that uses metrics.Metrics instead.
//
// You must use Get to retrieve the metrics.Metrics instance from a gin.Context.
func Logger(options ...func(cfg *LoggerConfig)) gin.HandlerFunc {
	cfg := &LoggerConfig{}
	for _, optFn := range options {
		optFn(cfg)
	}

	return func(c *gin.Context) {
		var (
			ctx = c.Request.Context()
			m   *metrics.Metrics
			ok  bool
		)

		m, ok = metrics.TryGet(ctx)
		if !ok {
			if newFn := cfg.newMetrics; newFn == nil {
				ctx, m = metrics.NewWithContext(ctx)
				c.Request = c.Request.WithContext(ctx)
			} else {
				m = newFn(c)
				c.Request = c.Request.WithContext(metrics.WithContext(ctx, m))
			}
		}

		var logger = cfg.Parent
		if logger == nil {
			logger = slog.Default()
		}

		if cfg.requestId != nil {
			rid := cfg.requestId()
			m.String("requestId", rid)
			c.Header("X-Request-Id", rid)
			c.Set(requestIdKey, rid)
			c.Set(slogLoggerKey, logger.With(slog.String("requestId", rid)))
		}

		m.
			String("ip", c.ClientIP()).
			String("method", c.Request.Method).
			String("path", c.Request.URL.Path).
			String("referrer", c.Request.Referer()).
			String("userAgent", c.Request.UserAgent())

		defer func() {
			if !cfg.DisableRecovery {
				if r := recover(); r != nil {
					m.Panicked()

					switch v := r.(type) {
					case error:
						m.Error(v)
					default:
						m.Error(eris.Wrapf(fmt.Errorf("%+v", v), "recover non-error %T: %#v", v, v))
					}

					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}

			if errors := c.Errors.ByType(gin.ErrorTypePrivate); len(errors) != 0 {
				for _, err := range errors {
					m.PushError(err, true)
				}
			}

			if err := c.Errors.Last(); err != nil && !m.HasError() {
				m.Error(err)
			}

			w := c.Writer

			m.
				Int64("status", int64(w.Status())).
				Int64("size", int64(w.Size()))

			if !ok {
				_ = m.CloseContext(c)
			}
		}()

		c.Next()
	}
}

// LoggerConfig contains customisations for Logger middleware.
type LoggerConfig struct {
	// Skip indicates which kind of requests to skip logging.
	//
	// Combine both [gin.LoggerConfig.SkipPath] and [gin.LoggerConfig.Skip].
	Skip func(ctx context.Context, req *http.Request) bool

	// DisableRecovery, if specified, disable gin.Recovery replacement.
	DisableRecovery bool

	// Parent is the slog.Logger instance that is used to derive loggers for specific requests.
	//
	// If nil, slog.Default will be used. The child loggers can be retrieved with Slog.
	Parent *slog.Logger

	newMetrics func(c *gin.Context) *metrics.Metrics
	requestId  func() string
}

// SkipPath is a convenient method to replace LoggerConfig.Skip with one that will skip logging for any request to the
// given paths.
func SkipPath(paths ...string) func(cfg *LoggerConfig) {
	m := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		m[path] = struct{}{}
	}

	return func(cfg *LoggerConfig) {
		cfg.Skip = func(_ context.Context, req *http.Request) bool {
			_, ok := m[req.URL.Path]
			return ok
		}
	}
}

// WithCustomMetrics can be used to customise how the metrics.Metrics instance is created and attached to gin.Context.
//
// Useful if you need to populate the metrics.Metrics instance with additional properties, or you want to change how
// the metrics.Metrics instance is logged. Note: if a metrics.Metrics instance is already available from context, the
// middleware will not create a new one (hence it will not trigger fn), and it will not be responsible for closing the
// instance either since the instance may have been created by an earlier middleware, and that middleware should be
// responsible for closing and logging the metrics.
func WithCustomMetrics(fn func(c *gin.Context) *metrics.Metrics) func(cfg *LoggerConfig) {
	return func(cfg *LoggerConfig) {
		cfg.newMetrics = fn
	}
}

// Get correctly retrieves the metrics.Metrics instance from the underlying request's context.
func Get(c *gin.Context) *metrics.Metrics {
	return metrics.Get(c.Request.Context())
}
