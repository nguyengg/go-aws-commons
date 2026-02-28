# Gin Metrics Middleware

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/gin-metrics.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-metrics)

Intended to work with [github.com/nguyengg/go-aws-commons/metrics](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/metrics)
as replacement for `gin.Logger`, `gin.Recovery`, and https://github.com/gin-contrib/requestid combined.

```shell
go get github.com/nguyengg/go-aws-commons/gin-metrics
```

Usage

```shell
package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	ginmetrics "github.com/nguyengg/go-aws-commons/gin-metrics"
)

func main() {
	r := gin.New()

	// by adding the Logger middleware with request Id enabled, both the metrics instance and the logger comes with
	// request Id attached for each request.
	r.Use(ginmetrics.Logger(ginmetrics.WithRequestId()))

	r.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")

		// the metrics are logged at the end of the request when execution comes back to Logger middleware.
		m := ginmetrics.Get(c).String("id", id)

		// logger can be used to log any messages any time. it comes with "requestId" attached, and you can attach
		// additional attributes there to this way.
		logger := ginmetrics.GetLogger(c, func(logger *slog.Logger) *slog.Logger {
			return logger.With(slog.String("id", id))
		})

		m.SetCounter("userDidSomethingCool", 1)
		logger.Info("user did something really cool!")
		c.Status(http.StatusNoContent)
	})
}

```
