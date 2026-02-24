# Implements io.ReadSeeker, io.ReaderAt, and io.WriterTo using S3 ranged GetObject

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/s3reader.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/s3reader)

This module provides implementations of `io.ReadSeeker`, `io.ReaderAt`, and `io.WriterTo` for S3 downloading needs.

**Note:** [feature/s3/transfermanager](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager)
(replacing [feature/s3/manager](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager)) is excellent (and
probably better-tested with the resources available at Amazon) than my library so give that a shot first.

```shell
go get github.com/nguyengg/go-aws-commons/s3reader
```

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nguyengg/go-aws-commons/s3reader"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stop()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)

	// s3reader.Reader implements both io.ReadSeeker and io.ReaderAt so I can start streaming the
	// S3 object however I want.
	// if in interactive mode, s3reader.WithProgressBar will show a progress bar displaying progress.
	// otherwise, use s3reader.WithProgressLogger instead.
	r, err := s3reader.New(ctx, client, &s3.GetObjectInput{
		Bucket: aws.String("my-bucket"),
		Key:    aws.String("my-key"),
	}, s3reader.WithProgressBar())
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close() // close will terminate the goroutine pool but is not strictly needed.

	// if writing to file, can also use https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager.
	dst, _ := os.CreateTemp("", "")
	_, _ = r.WriteTo(dst)
	_ = dst.Close()
}

```
