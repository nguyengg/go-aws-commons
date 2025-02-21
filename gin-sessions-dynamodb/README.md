# Very opinionated Gin middleware for session management backed by DynamoDB

There exist several implementations for a DynamoDB store plugin for https://github.com/gin-contrib/sessions and
https://github.com/gorilla/sessions. My module is more focused on:
1. Services that already have a DynamoDB table storing session data with defined schema. gin/gorilla sessions abstract
   this for you so generally you may not even know what your table schema looks like; if this works for you, go for it.
2. Type-safe `sessions.Session` implementation, or work directly on the session struct with the DynamoDB tags.

The module makes use of [github.com/nguyengg/go-aws-commons/ddb](https://pkg.go.dev/github.com/nguyengg/go-aws-commons/ddb)
to parse the `dynamodbav` struct tags. A struct representing your session can always be retrieved from the current
`gin.Context`. You can call the usual `Get`, `Set`, `Delete`, `Flash` on the `sessions.Session` instance (if you pass in
the wrong type, the program will panic), or you can retrieve and work with a pointer to the session struct directly.
