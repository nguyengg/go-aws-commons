# Convert DynamoDB last evaluated key to opaque token; create and validate CSRF tokens

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commonds/opaque-token.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/opaque-token)

This library was born out of my need to encrypt the `map[string]AttributeValue` last evaluated key from my DynamoDB
Query or Scan operations before passing it as the pagination token to the caller, though the library has grown to
support any `[]byte` token. ChaCha20-Poly1305 (preferred) and AES with GCM encryption are available, and you can either
provide a key statically, or from AWS Secrets Manager to get rotation support for free.

## Usage

Get with:

```shell
go get github.com/nguyengg/go-aws-commons/opaque-token
```

### Fixed key with ChaCha20-Poly1305 or AES encryption

Binary secret of valid ChaCha20-Poly1305 key size (256-bit) or AES key sizes (128-bit, 192-bit, or 256-bit) must be
given at construction time. Use this version if you're just testing out or aren't worried about having some impact when
rotating the secret (i.e. you can take some downtime, or it's a personal project where traffic is low or impact is not
business critical).

```go
package main

import (
	"context"
	"crypto/rand"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/opaque-token/ddb"
)

func main() {
	ctx := context.Background()
	cfg, _ := config.LoadDefaultConfig(ctx)
	client := dynamodb.NewFromConfig(cfg)
	queryOutputItem, _ := client.Query(ctx, &dynamodb.QueryInput{})

	key := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, key)
	c, _ := ddb.New(ddb.WithChaCha20Poly1305(key))

	// continuationToken is an opaque token that can be returned to user without leaking details about the table.
	continuationToken, _ := c.EncodeKey(ctx, queryOutputItem.LastEvaluatedKey)

	// to decrypt the opaque token and use it as exclusive start key in Query or Scan.
	exclusiveStartKey, _ := c.DecodeToken(ctx, continuationToken)
	_, _ = client.Query(ctx, &dynamodb.QueryInput{ExclusiveStartKey: exclusiveStartKey})
}

```

### Key from AWS Secrets Manager

AES key is retrieved from AWS Secrets Manager instead. Because each secret in AWS Secrets Manager has a version Id, this
pair of encoder/decoder will prefix the version Id to the opaque token (since the secret name and AWS account and region
are not leaked, this should be OK). Be mindful of the cost of calling AWS Secrets Manager for every invocation. If
running in AWS Lambda functions, you can make use of
[Dynamic key with AWS Secrets Manager in Lambda functions](#dynamic-key-with-aws-secrets-manager-in-lambda-functions).

```go
package main

import (
	"context"
	"crypto/rand"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/opaque-token"
)

func main() {
	ctx := context.Background()
	cfg, _ := config.LoadDefaultConfig(ctx)
	client := dynamodb.NewFromConfig(cfg)
	queryOutputItem, _ := client.Query(ctx, &dynamodb.QueryInput{})

	key := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, key)
	c, _ := token.NewDynamoDBKeyConverter(token.WithChaCha20Poly1305(key))

	// continuationToken is an opaque token that can be returned to user without leaking details about the table.
	continuationToken, _ := c.EncodeKey(ctx, queryOutputItem.LastEvaluatedKey)

	// to decrypt the opaque token and use it as exclusive start key in Query or Scan.
	exclusiveStartKey, _ := c.DecodeToken(ctx, continuationToken)
	_, _ = client.Query(ctx, &dynamodb.QueryInput{ExclusiveStartKey: exclusiveStartKey})
}

```
### Key from AWS Parameters and Secrets Lambda Extension

If running in AWS Lambda, this pair of encoder/decoder can make use of the AWS Parameters and Secrets Lambda Extension
(https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets_lambda.html) instead of directly using
Secrets Manager SDK.

```go
package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/go-aws-commons/opaque-token"
)

func main() {
	ctx := context.Background()
	cfg, _ := config.LoadDefaultConfig(ctx)
	client := dynamodb.NewFromConfig(cfg)
	queryOutputItem, _ := client.Query(ctx, &dynamodb.QueryInput{})

	c, _ := token.NewDynamoDBKeyConverter(token.WithKeyFromLambdaExtensionSecrets("my-secret-id"))

	// continuationToken is an opaque token that can be returned to user without leaking details about the table.
	// the token includes the plaintext version id so that DecodeToken knows which key to use.
	continuationToken, _ := c.EncodeKey(ctx, queryOutputItem.LastEvaluatedKey)
	exclusiveStartKey, _ := c.DecodeToken(ctx, continuationToken)
	_, _ = client.Query(ctx, &dynamodb.QueryInput{ExclusiveStartKey: exclusiveStartKey})
}

```

### HMAC and CSRF token generation and verification

The module also provides way to sign and verify payload. To make the signature a suitable CSRF token, be sure to pass a
non-zero nonce size for anti-collision purposes, while also including the session id or any other session-dependent
value according to https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#pseudo-code-for-implementing-hmac-csrf-tokens.

```go
package main

import (
	"context"

	"github.com/nguyengg/go-aws-commons/opaque-token/hmac"
)

func main() {
	ctx := context.Background()

	// you can use `hasher.WithKey` or `hasher.WithKeyFromSecretsManager` as well.
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
}

```

## Key Rotation

### To create a new 32-byte binary secret

```shell
file=$(mktemp)
openssl rand 32 > "${file}"
aws secretsmanager create-secret --name my-secret-name --secret-binary "fileb://${file}"
rm "${file}"

```

### To update an existing binary secret

```shell
file=$(mktemp)
openssl rand 32 > "${file}"
aws secretsmanager put-secret-value --name my-secret-name --secret-binary "fileb://${file}"
rm "${file}"

```
