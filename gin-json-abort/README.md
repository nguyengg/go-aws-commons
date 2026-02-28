# Abort gin request with JSON response 

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/gin-json-abort.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-json-abort)

```shell
go get github.com/nguyengg/go-aws-commons/gin-json-abort
```

Usage

```shell
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
	"github.com/nguyengg/go-aws-commons/errors"
	ginmetrics "github.com/nguyengg/go-aws-commons/gin-metrics"
	preconditions "github.com/nguyengg/go-aws-commons/gin-preconditions"
	"github.com/nguyengg/go-aws-commons/gin-preconditions/s3headers"
	"github.com/nguyengg/go-aws-commons/must"
)

func main() {
	ctx := context.Background()
	client := s3.NewFromConfig(configcache.MustGet(ctx))
	r := gin.New()
	r.Use(ginmetrics.Logger(ginmetrics.WithRequestId()))
	r.GET("/v1/:bucket/:key", func(c *gin.Context) {
		output := must.Must(client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(c.Param("bucket")),
			Key:    aws.String(c.Param("key")),
		}))

		// if user sets If-Modified-Since, check it so that we can avoid having to return the entire content.
		switch ignored, isModifiedSince := preconditions.IfModifiedSince(c, aws.ToTime(output.LastModified)); {
		case !ignored && !isModifiedSince:
			c.Status(http.StatusNotModified)
			return
		}

		// If-Match is also supported, which is supposed to be "stronger" than If-Modified-Since.
		switch _, matches, err := preconditions.IfMatch(c, preconditions.NewStrongETag(*output.ETag)); {
		case !matches && err == nil:
			c.Status(http.StatusNotModified)
			return
		case err != nil:
			_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid If-Match: %v", err))
			return
		}

		c.DataFromReader(200, *output.ContentLength, *output.ContentType, output.Body, nil)
		return
	})

	// you can also use s3headers module to skip the GetObject call as well like this.
	// this makes the handler almost a proxy.
	r.GET("/v2/:bucket/:key", func(c *gin.Context) {
		input := &s3.GetObjectInput{
			Bucket: aws.String(c.Param("bucket")),
			Key:    aws.String(c.Param("key")),
		}

		// this will copy any If-Match, If-None-Match, etc. headers.
		s3headers.CopyToGetObjectInput(c, input)

		output, err := client.GetObject(ctx, input)
		if errors.StatusCode(err) == http.StatusNotModified {
			c.Status(http.StatusNotFound)
			return
		}

		// now copy the headers back to the response.
		s3headers.CopyFromGetObjectOutput(c, output)

		c.DataFromReader(200, *output.ContentLength, *output.ContentType, output.Body, nil)
		return
	})
}

```
