# Very opinionated gin session middleware with DynamoDB backend

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb.svg)](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb)

There are already several excellent DynamoDB store plugins for
[`github.com/gin-contrib/sessions`](https://github.com/gin-contrib/sessions) (well, mostly from
[`github.com/gorilla/sessions`](https://github.com/gorilla/sessions)). This module (named `sessions`) does something a
bit different: you must bring your own struct that uses `dynamodbav` struct tags to model the DynamoDB table that
contains session data. When handling a request, you can either work directly with a pointer to this struct, or use a
type-safe `sessions.Session`-compatible implementation that can return an error or panic if you attempt to set a field
with the wrong type.

I created this module because I love how easy it is to use the middleware to manage sessions, but I already have my own
DynamoDB table for session data. If you're starting new, the various DynamoDB store plugins will abstract away the need
to define the DynamoDB schema so you don't have to care about it at all. But if you already have your own table, this
module is for you.

## Usage

Get with:

```shell
go get github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb
```

```go
package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	sessions "github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb"
	"github.com/nguyengg/go-aws-commons/gin-sessions-dynamodb/groups"
)

type Session struct {
	Id   string `dynamodbav:"sessionId,hashkey" tableName:"session"`
	User *User  `dynamodbav:"user"`
}

type User struct {
	Sub    string   `dynamodbav:"sub"`
	Groups []string `dynamodbav:"groups,stringset"`
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(sessions.Sessions[Session]("sid", func(s *sessions.Session) {
		// if you don't explicitly provide a client, `config.LoadDefaultConfig` is used similar to this example.
		s.Client = dynamodb.NewFromConfig(cfg)
	}))

	r.GET("/", func(c *gin.Context) {
		// this is type-safe way to interaction with my session struct.
		var mySession *Session = sessions.Get[Session](c)
		mySession.User = &User{Sub: "henry", Groups: []string{"poweruser"}}
		if err = sessions.Save(c); err != nil {
			_ = c.AbortWithError(http.StatusBadGateway, err)
			return
		}

		// alternatively, I can use the sessions.Session interface "compatible" with gin and gorilla.
		s := sessions.Default(c)
		s.Set("user", "henry")
		if err = s.Save(); err != nil {
			_ = c.AbortWithError(http.StatusBadGateway, err)
			return
		}
	})

	// the module also provides a basic middleware to verify user from the session is authorised based on group
	// membership.
	r.GET("/protected/resource", groups.MustHave(func(c *gin.Context) (bool, groups.Groups) {
		user := sessions.Get[Session](c).User
		if user == nil {
			return false, nil
		}

		return true, user.Groups
	}, groups.OneOf("canReadResource", "canWriteResource")))
}

```
