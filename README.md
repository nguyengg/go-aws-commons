# go-aws-commons - JakartaCommons meets Go, for lack of better naming

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons)

Henry's Golang multi-module workspace containing various libraries to make using AWS just a little bit more fun.

## DynamoDB goodies

This package adds optimistic locking and auto-generated timestamps by modifying the expressions being created as part of
a DynamoDB service call. Here's a snippet.

First, add new tags to your struct that can be parsed by `ddb` module like this:
```go
type Item struct {
	Id           string    `dynamodbav:"id,hashkey" tableName:"my-table"`
	Sort         string    `dynamodbav:"sort,sortkey"`
	Version      int64     `dynamodbav:"version,version"`
	CreatedTime  time.Time `dynamodbav:"createdTime,createdTime,unixtime"`
	ModifiedTime time.Time `dynamodbav:"modifiedTime,modifiedTime,unixtime"`
}
```

Then you can use the functions right off `ddb` module to execute DeleteItem, GetItem, PutItem, and UpdateItem with
optimistic locking and auto-generated timestamps working behind the scene.
```go
item := &Item{Id: "myId", Sort: "sort"}
getItemOutput, err := ddb.Get(context.Background(), item, item)
if len(getItemOutput) == 0 {
	// not found.
} else {
	// the response of GetItem should have been unmarshalled into item for me.
}
```

See [ddb](ddb) for more examples.

## Lambda handler wrappers with sensible defaults

The various `StartABC` functions wrap your Lambda handler so that a [Metrics](metrics) instance is available from
context and will be logged with sensible default metrics (start and end time, latency, fault, etc.) upon return of your
Lambda handler.

```go
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nguyengg/go-aws-commons/lambda"
	"github.com/nguyengg/go-aws-commons/metrics"
)

func main() {
	// you can use a specific specialisation for your handler like DynamoDB stream event below.
	lambda.StartDynamoDBEventHandleFunc(func(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
		m := metrics.Ctx(ctx)
		m.IncrementCount("myMetric")
		return events.DynamoDBEventResponse{}, nil
	})

	// or you can use the generic StartHandlerFunc template if there isn't a specialisation.
	lambda.StartHandlerFunc(func(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
		m := metrics.Ctx(ctx)
		m.IncrementCount("myMetric")
		return events.DynamoDBEventResponse{}, nil
	})

	// when your handler returns, the Metrics instance will be logged to standard error stream.
}

```

See [lambda](lambda) for more examples.

## Logging SDK latency metrics and other custom metrics

AWS SDK Go v2 middleware to measure and emit latency and fault metrics on the AWS requests. Additionally, you can also
emit custom metrics in JSON format which can then be parsed in CloudWatch Logs or turned into CloudWatch metrics.

The most convenient way to use `metrics` module is to attach it as a middleware to the SDK config.
```go
cfg, _ := config.LoadDefaultConfig(context.Background(), metrics.WithClientSideMetrics())
dynamodbClient := dynamodb.NewFromConfig(cfg)
```

Once processing finishes, logs the `Metrics` instance with zerolog to get JSON output like this:
```json
{
   "startTime": 1739504515510,
   "endTime": "Fri, 14 Feb 2025 03:41:57 GMT",
   "time": "1602.040 ms",
   "statusCode": 200,
   "counters": {
      "S3.GetObject.ServerFault": 0,
      "S3.GetObject.UnknownFault": 0,
      "DynamoDB.Query.ClientFault": 0,
      "DynamoDB.Query.ServerFault": 0,
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
```

See [metrics](metrics) for more examples.

## Convert DynamoDB last evaluated key to opaque token; create and validate CSRF tokens

This library was born out of my need to encrypt the `map[string]AttributeValue` last evaluated key from my DynamoDB
Query or Scan operations before passing it as the pagination token to the caller, though the library has grown to
support any `[]byte` token. ChaCha20-Poly1305 (preferred) and AES with GCM encryption are available, and you can either
provide a key statically, or from AWS Secrets Manager to get rotation support for free.

My API back-end uses Lambda Function URL so when I have to return a continuation token to user that is essentially a
DynamoDB query or scan key, I would encrypt/decrypt it like this to make it an opaque token:
```go
c, _ := token.NewDynamoDBKeyConverter(token.WithKeyFromLambdaExtensionSecrets("my-secret-id"))

// continuationToken is an opaque token that can be returned to user without leaking details about the table.
// the token includes the plaintext version id so that DecodeToken knows which key to use.
continuationToken, _ := c.EncodeKey(ctx, queryOutputItem.LastEvaluatedKey)
exclusiveStartKey, _ := c.DecodeToken(ctx, continuationToken)
_, _ = client.Query(ctx, &dynamodb.QueryInput{ExclusiveStartKey: exclusiveStartKey})
```

Because `DynamoDBKeyConverter` includes the version Id of the AWS Secrets Manager's secret in the token, I can safely
rotate the secret without impacting current users. See [opaque-token](opaque-token) for examples.

The module also has support for HMAC generation and verification which I also use for CSRF as well.
```go
signer := hmac.New(hmac.WithKeyFromLambdaExtensionSecrets("my-secret-id"))

// to use the signature as CSRF token, include session-dependent value according to
// https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#pseudo-code-for-implementing-hmac-csrf-tokens.
// don't add a random value in the payload; by passing non-zero nonce size, the generated token will already
// include a nonce for anti-collision purposes.
payload = []byte("84266fdbd31d4c2c6d0665f7e8380fa3")
signature, _ = signer.Sign(ctx, payload, 16)
ok, _ = signer.Verify(ctx, signature, payload)
if !ok {
	panic("CSRF verification fails")
}
```

## Implements io.ReadSeeker, io.ReaderAt, and io.WriterTo using S3 ranged GetObject

I wrote this module when [s3/manager](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager) didn't make it
easy to provide progress monitoring. Additionally, my [xy3](https://github.com/nguyengg/xy3) project needs a way to read
backwards an S3 object in order to find ZIP central directory. As a result, I wrote this module with the explicit goal
of implementing `io.Seeker` and `io.ReaderAt` for S3 objects. If you only need to download an entire S3 object to file
or to memory, [s3/manager](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager) may suffice. See
[s3reader](s3reader) for examples.

## Implements io.Writer and io.ReaderFrom to upload to S3

Similar to [s3reader](s3reader), I wrote this module when I needed a way to provide progress monitoring when uploading
files using [s3/manager](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager). I could have used an
`io.TeeReader` passing an [io.Writer progressbar](https://github.com/schollz/progressbar), but this will only report
progress on reading the file, not uploading the file. As a result, I wrote this module with the explicit goal of
accurately showing upload progress. See [s3writer](s3writer) for examples.

## Protect EC2 instances from being scaled down while busy

Monitor workers' statuses to enable or disable instance scale-in protection accordingly, inspired by
https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-using-sqs-queue.html#scale-sqs-queue-scale-in-protection.
Essentially, if you have any number of workers who can be either ACTIVE or IDLE, you generally want to enable scale-in
protection when any of your worker is actively doing some work, while once all the workers have become idle, you would
want to disable scale-in protection to let the Auto Scaling group reclaim your instance naturally. See
[scale-in-protection](scale-in-protection) for examples.

## Subresource Integrity computation and verification

Subresource Integrity ([SRI](https://developer.mozilla.org/en-US/docs/Web/Security/Subresource_Integrity)) is a hash
prefixed with the hash function name. The [sri](sri) module provides functionality to generate and verify SRI hashes:

```go
// h implements hash.Hash which implements io.Writer so just pipes an entire file to it.
h := sri.NewSha256()
f, _ := os.Open("path/to/file")
_, _ = f.WriteTo(h)
_ = f.Close()

// SumToString will produce a digest in format sha256-aOZWslHmfoNYvvhIOrDVHGYZ8+ehqfDnWDjUH/No9yg for example.
h.SumToString(nil)

// To verify against a set of expected hashes, pass them into NewVerifier.
// v also implements hash.Hash so just pipes the entire file to it.
v, _ := sri.NewVerifier(
	"sha256-aOZWslHmfoNYvvhIOrDVHGYZ8+ehqfDnWDjUH/No9yg", 
	"sha384-b58jhCXsokOe1Fgawf20X8djeef7qUvAp2JPo+erHsNwG0v83aN2ynVRkub0XypO", 
	"sha512-bCYYNY2gfIMLiMWvjDU1CA6OYDyIuJECiiWczbmsgC0PwBcMmdWK/88AeGzhiPxddT6MZiivIHHDJw1QRFxLHA")
f, _ := os.Open("path/to/file")
_, _ = f.WriteTo(v)
_ = f.Close()

// SumAndVerify will return true if and only if the hash matches against the set of hashes passed to NewVerifier.
if matches := v.SumAndVerify(nil); !matches {
	// the file's content may have been corrupted.
}
```
