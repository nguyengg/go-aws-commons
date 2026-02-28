# Set response caching headers in Gin handlers

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/gin-caching-response-headers.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-caching-response-headers)

```shell
go get github.com/nguyengg/go-aws-commons/gin-caching-response-headers
```

Usage:
```shell
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	cachingheaders "github.com/nguyengg/go-aws-commons/gin-caching-response-headers"
	cc "github.com/nguyengg/go-aws-commons/gin-caching-response-headers/cachecontrol"
)

// Response implements all three Cache-Control, ETag, and Last-Modified.
type Response struct {
	Data         []byte    `json:"data"`
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"lastModified"`
}

var _ cachingheaders.HasCacheControl = &Response{}
var _ cachingheaders.HasCacheControl = (*Response)(nil)

func (r Response) GetCacheControl() string {
	return cc.Join(cc.SMaxAge(3*time.Hour), cc.Private)
}

var _ cachingheaders.HasETag = &Response{}
var _ cachingheaders.HasETag = (*Response)(nil)

func (r Response) GetETag() string {
	// must be wrapped by quotes such as "xyyyz", and if it's a weak ETag, it must be prefixed with W/ like W/"xyyyz".
	return fmt.Sprintf(`"%s"`, strings.Trim(r.ETag, `"`))
}

var _ cachingheaders.HasLastModified = &Response{}
var _ cachingheaders.HasLastModified = (*Response)(nil)

func (r Response) GetLastModified() time.Time {
	return r.LastModified
}

func main() {
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		res := Response{
			ETag:         `W/"xyyyz"`,
			LastModified: time.Now(),
		}

		// this will set "Cache-Control", "ETag", and "Last-Modified".
		cachingheaders.Set(c, res)
		c.JSON(200, res)
	})
}

```
