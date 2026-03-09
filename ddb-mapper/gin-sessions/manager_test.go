package sessions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Using Manager interface, a usual workflow would get session, update its data, the save to commit the changes.
func Test_manager_UsualWorkflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	cookieName := "my-session"
	sid := "my-session-id"

	client := &MockManagerAPIClient{}

	sessions, err := New[TestSession](cookieName, func(cfg *Config) {
		cfg.Client = client
		cfg.NewSessionId = func() string {
			return sid
		}
	})
	require.NoError(t, err)

	r.GET("/", func(c *gin.Context) {
		// initial Manager.Get will return this expected TestSession with version == 1.
		client.mockGetItem(t, sid, &TestSession{SessionId: sid, Version: 1})
		v, err := sessions.Get(c)
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
		err = sessions.Save(c)
		require.NoError(t, err)
		require.Equal(t, &TestSession{SessionId: sid, User: "another-user", Version: 2}, v)
		client.AssertNumberOfCalls(t, "PutItem", 1)

		// test Manager.Destroy too, after which TryGet must return nil.
		client.mockDestroyItem(t, sid)
		require.NoError(t, sessions.Destroy(c))
		client.AssertNumberOfCalls(t, "DeleteItem", 1)

		client.mockGetItem(t, sid, nil)
		v, err = sessions.TryGet(c)
		require.NoError(t, err)
		require.Nil(t, v)
		client.AssertNumberOfCalls(t, "GetItem", 2)
	})

	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: sid,
	})
	r.ServeHTTP(w, c.Request)
}
