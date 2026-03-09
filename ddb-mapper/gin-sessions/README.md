# Gin helper methods to manage sessions backed by DynamoDB

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commonds/ddb-mapper/gin-sessions.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/db-mapper/gin-sessions)

Get with:
```shell
go get github.com/nguyengg/go-aws-commonds/ddb-mapper/gin-sessions
```

Usage:
```go
package main

import (
	"github.com/gin-gonic/gin"
	sessions "github.com/nguyengg/go-aws-commons/gin-dynamodb-sessions"
)

// Sessions is a global variable that can be used across multiple handlers.
var Sessions *sessions.Sessions[Session]

// init will initialise Sessions and fail fast if Session is invalid.
func init() {
	var err error
	Sessions, err = sessions.New[Session]()
	if err != nil {
		panic(err)
	}
}

func main() {
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		// retrieve the session or create a new one.
		s, _ := Sessions.Get(c)

		// add new metadata to the session.
		s.ShoppingCart = []string{"shirt"}

		// save back to DynamoDB.
		_ = Sessions.Save(c)
	})
}

type Session struct {
	Id           string   `dynamodbav:"sessionId,hashkey" tableName:"session"`
	User         *User    `dynamodbav:"user"`
	ShoppingCart []string `dynamodbav:"shoppingCart,stringset"`
}

type User struct {
	Sub    string   `dynamodbav:"sub"`
	Groups []string `dynamodbav:"groups,stringset"`
}

```
