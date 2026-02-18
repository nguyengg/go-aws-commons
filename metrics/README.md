# Logging latency metrics and other custom counters

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

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/metrics"
)

func main() {
	// this will attach a middleware that logs all AWS calls for latency.
	cfg, _ := config.LoadDefaultConfig(context.Background(), metrics.WithClientSideMetrics())

	// in your handler, before making the AWS calls, you must attach a metrics instance to the context that will be
	// passed to the clients. Creating the Metrics instance automatically set the start time to time.Now.
	ctx, m := metrics.NewWithContext(context.Background())

	// you can use the metrics instance to add more metrics.
	// metrics.Ctx can also be used to retrieve the instance from context.
	m.AddCounter("userDidSomethingCool", 1)

	// do some intensive work such as making DynamoDB service calls.
	client := dynamodb.NewFromConfig(cfg)
	client.GetItem(ctx, &dynamodb.GetItemInput{})

	// closing the Metrics instance will automatically log to os.Stderr an entry like this.
	_ = m.CloseContext(ctx)

	/*
		{
		  "counters": {
		    "DynamoDB.GetItem.ClientFault": 0,
		    "DynamoDB.GetItem.ServerFault": 0,
		    "DynamoDB.GetItem.UnknownFault": 0,
		    "fault": 0,
		    "panicked": 0,
		    "userDidSomethingCool": 1
		  },
		  "endTime": "Sat, 01 Jan 2000 00:00:03 UTC",
		  "latency": "3s",
		  "startTime": 946684800000,
		  "timings": {
		    "DynamoDB.GetItem": {
		      "sum": "74.255ms",
		      "min": "74.255ms",
		      "max": "74.255ms",
		      "n": 1,
		      "avg": "74.255ms"
		    }
		  }
		}
	*/

	// in CloudWatch Logs, you can put a filter like this { $.['DynamoDB.GetItem.ServerFault'] = * } to measure 500s
	// from DynamoDB.
	//
	// Similarly, you can now put a filter on { $.fault = * } to measure your own 500s.
	// See https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html#matching-terms-json-log-events.
}

```
