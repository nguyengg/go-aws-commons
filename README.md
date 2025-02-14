# go-aws-commons - JakartaCommons meets Go, for lack of better naming

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons)

Henry's Golang multi-module workspace containing various libraries to make using AWS just a little bit more fun.

## DynamoDB goodies

This package adds optimistic locking and auto-generated timestamps by modifying the expressions being created as part of
a DynamoDB service call. Here's a snippet.

First, add new tags to your struct that can be parsed by `ddb` module:
1. `hashkey`, `sortkey`, and `tableName`: must be tagged on valid key types (S, N, and B).
2. `version` (optional): must be a valid N type to enable optimistic locking.
3. `createdTime` and `modifiedTime` (both optional): must be a valid time.Time. Can choose from
   [timestamp](ddb/timestamp) module if you'd like more control the types, such as `timestamp.Day` (2006-01-02),
   `timestamp.EpochMillisecond`, `timestamp.EpochSecond` (useful as TTL values) or `timestamp.Timestamp`
   (2006-01-02T15:04:05.000Z - also how `time.Time` is marshalled to DynamoDB by default),
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

// to get a stable hash (same input produces same output), pass 0 for nonce size.
payload := []byte("hello, world")
signature, _ := signer.Sign(ctx, payload, 0)
ok, _ := signer.Verify(ctx, signature, payload)
if !ok {
	panic("signature verification fails")
}

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

Monitor workers' statuses to enable or disable instance scale-in protection accordingly. Inspired by
https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-using-sqs-queue.html#scale-sqs-queue-scale-in-protection:

```java
while (true)
{
    SetInstanceProtection(False);
    Work = GetNextWorkUnit();
    SetInstanceProtection(True);
    ProcessWorkUnit(Work);
    SetInstanceProtection(False);
}
```

Essentially, if you have any number of workers who can be either ACTIVE or IDLE, you generally want to enable scale-in
protection when any of your worker is actively doing some work, while once all the workers have become idle, you would
want to disable scale-in protection to let the Auto Scaling group reclaim your instance naturally.

**Note**: there is a possibility that your instance is terminated in-between the `GetWorkUnit()` and the
`ProcessWorkUnit(Work)` calls. Generally if your visibility timeout is low enough, this is not an issue as a different
worker would be able to pick up the message again.

See [scale-in-protection](scale-in-protection) for examples.
