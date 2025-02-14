# go-aws-commons - JakartaCommons meets Go, for lack of better naming

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons)

Henry's Golang multi-module workspace containing various libraries to make using AWS just a little bit more fun.

## DynamoDB goodies

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/ddb.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb)

This package adds optimistic locking and auto-generated timestamps by modifying the expressions being created as part of
a DynamoDB service call. See [ddb](ddb) for examples.

## Logging SDK latency metrics and other custom metrics

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/metrics.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/metrics)

AWS SDK Go v2 middleware to measure and emit latency and fault metrics on the AWS requests. Additionally, you can also
emit custom metrics in JSON format which can then be parsed in CloudWatch Logs or turned into CloudWatch metrics. See
[metrics](metrics) for examples.

## Convert DynamoDB last evaluated key to opaque token; create and validate CSRF tokens

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/opaque-token.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/opaque-token)

This library was born out of my need to encrypt the `map[string]AttributeValue` last evaluated key from my DynamoDB
Query or Scan operations before passing it as the pagination token to the caller, though the library has grown to
support any `[]byte` token. ChaCha20-Poly1305 (preferred) and AES with GCM encryption are available, and you can either
provide a key statically, or from AWS Secrets Manager to get rotation support for free. See [opaque-token](opaque-token)
for examples.

## Implements io.ReadSeeker, io.ReaderAt, and io.WriterTo using S3 ranged GetObject

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/s3reader.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/s3reader)

This module provides implementations of `io.ReadSeeker`, `io.ReaderAt`, and `io.WriterTo` for S3 downloading needs. See
[s3reader](s3reader) for examples.

# Implements io.Writer and io.ReaderFrom to upload to S3

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/s3writer.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/s3writer)

This module provides implementations of `io.Writer` and `io.ReaderFrom` for S3 uploading needs. See [s3writer](s3writer)
for examples.

## Protect EC2 instances from being scaled down while busy

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/scale-in-protection.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/scale-in-protection)

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
