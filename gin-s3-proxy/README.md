# S3 proxy Gin middleware

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/gin-s3-proxy.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-s3-proxy)

At a high level, you can use this module as a Gin middleware to proxy GET, HEAD, DELETE, PUT, and POST requests to S3.
You can also use low-level APIs to copy request headers such as If-Match, If-Modified-Since, etc. into S3 GetObject
input, and copy S3 GetObject output back into response headers.

Get with:
```shell
go get github.com/nguyengg/go-aws-commons/gin-s3-proxy
```

Usage:
```go
package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	configcache "github.com/nguyengg/go-aws-commons/config-cache"
	proxy "github.com/nguyengg/go-aws-commons/gin-s3-proxy"
)

func main() {
	cfg := configcache.MustGet(context.Background())
	client := s3.NewFromConfig(cfg)
	r := gin.New()
	r.GET("/:bucket/:key", proxy.New(client, proxy.WithBucketKeyFromParams("bucket", "key")))
}

```
