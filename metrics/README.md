# Logging SDK latency metrics and other custom metrics

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/metrics.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/metrics)

AWS SDK Go v2 middleware to measure and emit latency and fault metrics on the AWS requests. Additionally, you can also
emit custom metrics in JSON format which can then be parsed in CloudWatch Logs or turned into CloudWatch metrics.

## Usage

Get with:

```shell
go get github.com/nguyengg/go-aws-commons/metrics
```

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/metrics"
)

func main() {
	// this will attach a middleware that logs all AWS calls for latency.
	cfg, _ := config.LoadDefaultConfig(context.Background(), metrics.WithClientSideMetrics())

	// just use the cfg to create the AWS clients normally, for example with DynamoDB.
	_ = dynamodb.NewFromConfig(cfg)

	// in your handler, before making the AWS calls, you must attach a metrics instance to the context that will be
	// passed to the clients.
	ctx := metrics.WithContext(context.Background(), metrics.NewMetrics())

	// you can use the metrics instance to add more metrics.
	m := metrics.Get(ctx).AddCounter("userDidSomething", 1)

	// finally, you must log the metrics instance.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	slog.LogAttrs(context.Background(), slog.LevelInfo, "done", m.Attrs()...)
	// {
	//    "time": "2026-02-16T23:08:32.441543-08:00",
	//    "level": "INFO",
	//    "msg": "done",
	//    "startTime": 1771312112441,
	//    "endTime": "Mon, 16 Feb 2026 23:08:32 PST",
	//    "latency": "0.000ms",
	//    "counters": {
	//        "userDidSomething": 1
	//    }
	// }

	// if you want to put the metrics inside a "metrics" property:
	slog.LogAttrs(context.Background(), slog.LevelInfo, "done", slog.Any("metrics", m))
	// {
	//    "time": "2026-02-16T23:14:32.241391-08:00",
	//    "level": "INFO",
	//    "msg": "done",
	//    "metrics": {
	//        "startTime": 1771312472240,
	//        "endTime": "Mon, 16 Feb 2026 23:14:32 PST",
	//        "latency": "0.000ms",
	//        "counters": {
	//            "userDidSomething": 1
	//        }
	//    }
	// }
}

```
