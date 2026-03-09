package sessions_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	sessions "github.com/nguyengg/go-aws-commons/gin-dynamodb-sessions"
	"github.com/nguyengg/go-dynamodb-local"
	"github.com/stretchr/testify/require"
)

type TestSession struct {
	SessionId string `dynamodbav:"sessionId,hashkey" tablename:"session"`
	User      string `dynamodbav:"user"`
}

func Test(t *testing.T) {
	client := local.DefaultSkippable(t)
	ddb.DefaultClientProvider = &ddb.StaticClientProvider{Client: client}
	require.NoError(t, ddb.CreateTable(t.Context(), TestSession{}))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	cookieName := "my-session"
	sid := "my-session-id"

	sessions.New[TestSession]()
	require.NoError(t, err)

	r.GET("/", func(c *gin.Context) {
		// initial Manager.Get will return this expected TestSession with version == 0 because session does not
		// exist in database yet.
		v, err := sessions.Get(c)
		require.NoError(t, err)
		require.Equal(t, &TestSession{SessionId: sid, Version: 0}, v)

		// after Manager.Save, version becomes 1.
		v.User = "my-user"
		err = sessions.Save(c)
		require.NoError(t, err)
		require.Equal(t, &TestSession{SessionId: sid, User: "another-user", Version: 1}, v)

		// test Manager.Destroy too, after which TryGet must return nil.
		require.NoError(t, sessions.Destroy(c))
		v, err = sessions.TryGet(c)
		require.NoError(t, err)
		require.Nil(t, v)
	})

	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: sid,
	})
	r.ServeHTTP(w, c.Request)
}
