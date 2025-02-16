# Decouple how to retrieve a variable of any type (usually string or binary)

In Lambda, more often than not you would use environment variables to configure the runtime. You may also want to
retrieve from Parameter Store or Secrets Manager instead if you have credentials or secrets. This module provides an
abstract to decouple usage of an environment from how it is retrieved.

```go
package main

import (
   "context"
   "crypto/hmac"
   "crypto/sha256"
   "log"

   "github.com/aws/aws-sdk-go-v2/aws"
   "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
   "github.com/aws/aws-sdk-go-v2/service/ssm"
   "github.com/nguyengg/go-aws-commons/lambda/getenv"
)

func main() {
   // while prototyping, you can retrieve from environment variable
   // r is the string value from environment variable named TEST
   v := getenv.Env("TEST")
   r := v.MustGet()
   log.Printf("%s (%T)", r, r) // r is a string

   // now you want to retrieve from Parameter Store using the AWS Parameter Store and Secrets Lambda extension
   // (available since we're running in Lambda) instead.
   v = getenv.ParameterString(&ssm.GetParameterInput{
      Name:           aws.String("my-parameter-name"),
      WithDecryption: aws.Bool(true),
   })
   r, err := v.GetWithContext(context.Background())
   log.Printf("%s (%T)", r, r) // r is a string

   // if you need to retrieve some binary key from Secrets Manager using the AWS Parameter Store and Secrets Lambda
   // extension (available since we're running in Lambda).
   // in this example, the key is retrieved and then used as secret key for HMAC verification.
   key := getenv.SecretBinary(&secretsmanager.GetSecretValueInput{
      SecretId:     aws.String("my-secret-id"),
      VersionId:    nil,
      VersionStage: nil,
   })
   h := hmac.New(sha256.New, key.MustGetWithContext(context.Background()))
   h.Write( /* some data */ )
   h.Sum(nil)
}

```
