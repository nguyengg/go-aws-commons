# `log/slog` goodies

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/slogging.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/slogging)

Get with:
```shell
go get github.com/nguyengg/go-aws-commons/slogging
```

Usage:
```shell
package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	. "github.com/nguyengg/go-aws-commons/must"
	"github.com/nguyengg/go-aws-commons/slogging"
)

func main() {
	var (
		ctx = context.Background()
		err = errors.New("this is an error")
	)

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// this will log the error as JSON with stack trace.
	logger.LogAttrs(ctx, slog.LevelError, "to err or not to err", slogging.AnError("error", err))

	// you can also provide additional context to the error.
	logger.LogAttrs(ctx, slog.LevelError, "to err or not to err", slogging.Wrapf("error", err, "get s3://%s/%s error", "my-bucket", "my-key"))

	// if you have []byte or string that you believe are valid JSON (such as a GET response), you can do this:
	res := Must(http.Get("https://some.website.com/file.json"))
	defer res.Body.Close()
	data := Must(io.ReadAll(res.Body))
	logger.LogAttrs(ctx, slog.LevelInfo, "i got some json", slogging.JSON("payload", data))
}

```
