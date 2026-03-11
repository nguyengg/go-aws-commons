package sessions

import "github.com/gin-gonic/gin"

// session is a specific session for a specific request.
//
// It's a type-agnostic wrapper around a type-T value (v) and the value of the session Id.
type session struct {
	v   any
	sid string
}

const (
	// sessionKeyPrefix is the gin.Context key prefix that stores *session for the request.
	sessionKeyPrefix = "github.com/nguyengg/go-aws-commons/gin-dynamodb-sessions/sessionKey_"
	// managerKeyPrefix is the gin.Context key prefix that stores *manager[T] for the request.
	managerKeyPrefix = "github.com/nguyengg/go-aws-commons/gin-dynamodb-sessions/managerKey_"
)

// get retrieves a session from context.
func get(c *gin.Context, name string) (*session, bool) {
	if v, ok := c.Get(sessionKeyPrefix + name); ok {
		return v.(*session), true
	}

	return nil, false
}

// set sets the session to context.
func (s *session) set(c *gin.Context, name string) {
	c.Set(sessionKeyPrefix+name, s)
}

// unset deletes the session from context.
func (s *session) unset(c *gin.Context, name string) {
	c.Delete(sessionKeyPrefix + name)
}
