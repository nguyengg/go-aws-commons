# Logging SDK latency metrics and other custom metrics

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/metrics.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/metrics)

## Usage

Get with:

```shell
go get github.com/nguyengg/go-aws-commons/metrics
```

```go
package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/metrics"
	"github.com/rs/zerolog"
)

func main() {
	// this will attach a middleware that logs all AWS calls for latency.
	cfg, _ := config.LoadDefaultConfig(context.Background(), metrics.WithClientSideMetrics())

	// just use the cfg to create the AWS clients normally, for example with DynamoDB:
	_ = dynamodb.NewFromConfig(cfg)

	// in your handler, before making the AWS calls, you must attach a metrics instance to the context that will be
	// passed to the clients.
	ctx := metrics.WithContext(context.Background(), metrics.New())

	// you can use the metrics instance to add more metrics.
	m := metrics.Ctx(ctx).AddCount("userDidSomething", 1)

	// then, at the end of the request, print the metrics.
	m.Log(zerolog.Ctx(ctx))

	// here's a real log message from my production website.
	/**
	{
	    "startTime": 1739504515510,
	    "endTime": "Fri, 14 Feb 2025 03:41:57 GMT",
	    "time": "1602.040 ms",
	    "statusCode": 200,
	    "requestId": "0401b979-2355-4413-a4e8-ea8e6c798491",
	    "path": "/api/payments/issaquah/2025/02/13",
	    "method": "GET",
	    "sessionIdHash": "03eb5405e8cc3926",
	    "user.sub": "50407e6f-f34d-4762-9070-2bc26a011fc5",
	    "site": "issaquah",
	    "counters": {
	        "fault": 0,
	        "sessionCacheHit": 0,
	        "isAuthenticated": 1,
	        "availableFromS3": 0,
	        "S3.GetObject.ServerFault": 0,
	        "S3.GetObject.UnknownFault": 0,
	        "DynamoDB.Query.ClientFault": 0,
	        "DynamoDB.Query.ServerFault": 0,
	        "panicked": 0,
	        "S3.GetObject.ClientFault": 0,
	        "DynamoDB.Query.UnknownFault": 0,
	        "2xx": 1,
	        "4xx": 0,
	        "5xx": 0
	    },
	    "timings": {
	        "S3.GetObject": {
	            "sum": "64.680 ms",
	            "min": "64.680 ms",
	            "max": "64.680 ms",
	            "n": 1,
	            "avg": "64.680 ms"
	        },
	        "DynamoDB.Query": {
	            "sum": "74.255 ms",
	            "min": "74.255 ms",
	            "max": "74.255 ms",
	            "n": 1,
	            "avg": "74.255 ms"
	        }
	    }
	}
	*/
}

```
