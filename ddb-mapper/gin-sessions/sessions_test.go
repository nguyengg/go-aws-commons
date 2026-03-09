package sessions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb"
	"github.com/stretchr/testify/require"
)

// Test_UsualWorkflow is equivalent of Test_manager_UsualWorkflow but using package-level methods.
func Test_UsualWorkflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	sid := "my-session-id"

	client := &MockManagerAPIClient{}
	newDDBManager = func(cfg aws.Config) *ddb.Manager {
		return ddb.NewManager(client)
	}

	r.GET("/", func(c *gin.Context) {
		// initial Get will return this expected TestSession with version == 1.
		client.mockGetItem(t, sid, &TestSession{SessionId: sid, Version: 1})
		v, err := Get[TestSession](c)
		require.NoError(t, err)
		require.Equal(t, &TestSession{SessionId: sid, Version: 1}, v)
		client.AssertNumberOfCalls(t, "GetItem", 1)

		client.mockPutItem(
			t,
			// ddb.PutItem will modify the map[string]AttributeValue to add optimistic locking and bumping
			// version to 2.
			&TestSession{SessionId: sid, User: "my-user", Version: 2},
			// we mock the client returning "another-user" to show that v will be updated with whatever
			// ddb.PutItem returns via ReturnValueAllNew.
			&TestSession{SessionId: sid, User: "another-user", Version: 2})
		v.User = "my-user"
		err = Save[TestSession](c)
		require.NoError(t, err)
		require.Equal(t, &TestSession{SessionId: sid, User: "another-user", Version: 2}, v)
		client.AssertNumberOfCalls(t, "PutItem", 1)

		// test Destroy too, after which TryGet must return nil.
		client.mockDestroyItem(t, sid)
		require.NoError(t, Destroy[TestSession](c))
		client.AssertNumberOfCalls(t, "DeleteItem", 1)

		client.mockGetItem(t, sid, nil)
		v, err = TryGet[TestSession](c)
		require.NoError(t, err)
		require.Nil(t, v)
		client.AssertNumberOfCalls(t, "GetItem", 2)
	})

	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.AddCookie(&http.Cookie{
		Name:  DefaultSessionIdCookieName,
		Value: sid,
	})
	r.ServeHTTP(w, c.Request)
}
