# Gin adapter for Function URL

A Gin adapter for API Gateway V1 and V2 are already available from https://github.com/awslabs/aws-lambda-go-api-proxy.
This module provides an adapter specifically for Function URL events with both BUFFERED (which, technically, is no
different from API Gateway V2/HTTP events) and RESPONSE_STREAM mode which uses
https://github.com/aws/aws-lambda-go/tree/main/lambdaurl under the hood.

```go
package main

import (
	"github.com/gin-gonic/gin"
	sessions "github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb"
	ginadapter "github.com/nguyengg/go-aws-commons/lambda/functionurl/gin"
	"github.com/nguyengg/go-aws-commons/lambda/functionurl/gin/rules"
)

type Session struct {
	SessionId string `dynamodbav:"sessionId" tableName:"session"`
	User      *User  `dynamodbav:"user,omitempty"`
}

type User struct {
	Sub    string   `dynamodbav:"sub"`
	Groups []string `dynamodbav:"groups,stringset"`
}

func main() {
	r := gin.Default()

	// this example uses github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb to provide session management.
	r.GET("/",
		sessions.Sessions[Session]("sid"),
		ginadapter.RequireGroupMembership(func(c *gin.Context) (authenticated bool, groups rules.Groups) {
			var s *Session = sessions.Get[Session](c)
			if s.User == nil {
				return false, nil
			}
			return true, s.User.Groups
		}, rules.AllOf("a", "b"), rules.OneOf("b", "c")))

	// start the Lambda handler either in BUFFERED or STREAM_RESPONSE mode.
	ginadapter.StartBuffered(r)
	ginadapter.StartStream(r)
}

```
