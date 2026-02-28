# go-aws-commons - JakartaCommons meets Go, for lack of better naming

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons)

Henry's Golang multi-module workspace containing various libraries to make using AWS just a little bit more fun.

Available as their own module:
* [config-cache](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/config-cache) (`configcache`): single AWS config cache to make using package-level methods easier.
* [ddb](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb): DynamoDB goodies to add optimistic locking and auto-generated timestamps via struct tags.
* [gin-caching-response-headers](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-caching-response-headers) (`cachingheaders`): sets caching response headers (Cache-Control, ETag, and/or Last-Modified) on the gin response.
* [gin-json-abort](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-json-abort) (`abort`): provides package-level methods to help abort a gin request using JSON response as well as logging.
* [gin-metrics](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-metrics) (`ginmetrics`): replaces gin.Logger and gin.Recovery with metrics.Metrics integration.
* [gin-preconditions](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-preconditions) (`preconditions`): provides helper methods to parse and compare conditional headers such as If-Match, If-None-Match, If-Modified-Since, and If-Unmodified-Since.
* [gin-sessions-dynamodb](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb) (`sessions`): replaces [github.com/gin-contrib/sessions](https://pkg.go.dev/github.com/gin-contrib/sessions) with [ddb](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb) integration.
* [lambda](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/lambda): Lambda handler wrappers with sensible defaults and metrics integration.
  * [function-url](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/lambda/function-url) (`functionurl`): provides Lambda wrappers for Function URL gin handlers in either BUFFERED or STREAMING mode.
  * [getenv](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/lambda/getenv): decouple how to retrieve a variable of any type (usually string or binary) via AWS Parameter Store and Secrets Lambda extension.
* [metrics](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/metrics): logging latency metrics and other custom counters.
* [opaque-token](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/opaque-token) (`token`): convert DynamoDB last evaluated key to opaque token; create and validate CSRF tokens.
* [s3reader](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/s3reader): implements `io.ReadSeeker`, `io.ReaderAt`, and `io.WriterTo` using S3 ranged GetObject.
* [s3writer](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/s3writer): implements `io.Writer` and `io.ReaderFrom` for uploading to S3.
* [scale-in-protection](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/scale-in-protection) (`sip`): protect EC2 AutoScaling instances from being scaled down while busy.
* [sri](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/sri): Subresource Integrity (SRI) computation and verification.
* [tspb](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/tspb): Terminal-Safe Progress Bar (TSPB); when you want your program to show progress bar in interactive mode (with terminal), but log normally otherwise.

Available as package in this module ([commons](https://pkg.go.dev/github.com/nguyengg/go-aws-commons)):
* [args](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/args): iterator to scan text lines from multiple sources.
* [errors](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/errors): convenient methods to extract status code and other metadata from AWS errors.
* [executor](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/executor): Java's [Executor](https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/concurrent/Executor.html) for Go.
* [fmt](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/fmt): provides [fmt.Formatter](https://pkg.go.dev/fmt#Formatter) implementations for printing/logging any data as JSON.
* [must](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/must): for when you're tired of typing `if a, err := someFunction(); err != nil` and just want to panic instead.
* [slog](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/slog): attach/retrieve to/from context; JSON and error (with stack trace) `slog.Value` implementations.
* 
