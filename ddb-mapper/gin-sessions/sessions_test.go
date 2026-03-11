package sessions_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	sessions "github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sessions_Get(t *testing.T) {
	m, _ := setup(t)

	r := gin.New()
	r.GET("/", m.Middleware(), func(c *gin.Context) {
		// Get will create a new session Id, and Save will store it.
		v, err := sessions.Get(c)
		assert.NoError(t, err)

		s := v.(*Session)
		s.User = "me"
		assert.NoError(t, sessions.Save(c))

		assert.Equal(t, 1, s.Version)

		c.String(200, s.ID)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)

	// assert that a new session exists with the expected data.
	s := &Session{ID: w.Body.String()}
	_, err := ddb.Get(t.Context(), s)
	require.NoError(t, err)
	assert.Equal(t, w.Body.String(), s.ID)
	assert.Equal(t, "me", s.User)
	assert.Equal(t, 1, s.Version)
	assert.False(t, s.Created.IsZero())
	assert.False(t, s.Modified.IsZero())
}

func Test_sessions_Regenerate(t *testing.T) {
	m, _ := setup(t)

	// Regenerate will change the Id and Version, but nothing else.
	original := &Session{ID: "my-session-id", User: "me", Version: 6}
	_, err := ddb.Put(t.Context(), original, func(opts *config.PutOptions) {
		opts.DisableOptimisticLocking = true
	})
	require.NoError(t, err)

	r := gin.New()
	r.GET("/", m.Middleware(), func(c *gin.Context) {
		v, err := sessions.Regenerate(c)
		require.NoError(t, err)
		require.NoError(t, sessions.Save(c))
		c.String(200, v.(*Session).ID)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "sid=my-session-id")
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)

	s := &Session{ID: w.Body.String()}
	assert.NotEqual(t, s.ID, original.ID)
	_, err = ddb.Get(t.Context(), s)
	require.NoError(t, err)
	assert.Equal(t, w.Body.String(), s.ID)
	assert.Equal(t, "me", s.User) // unchanged
	assert.Equal(t, 1, s.Version) // reset due to new session
	assert.False(t, s.Created.IsZero())
	assert.False(t, s.Modified.IsZero())
}

func Test_sessions_Destroy(t *testing.T) {
	m, _ := setup(t)

	_, err := ddb.Put(t.Context(), &Session{ID: "my-session-id", User: "me"})
	require.NoError(t, err)

	r := gin.New()
	r.GET("/", m.Middleware(), func(c *gin.Context) {
		require.NoError(t, sessions.Destroy(c))
		require.NoError(t, sessions.Destroy(c))
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "sid=my-session-id")
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)

	getItemOutput, err := ddb.Get(t.Context(), &Session{ID: "my-session-id"})
	require.NoError(t, err)
	require.Empty(t, getItemOutput.Item)
}

func Test_sessions_TryGet(t *testing.T) {
	m, _ := setup(t)

	r := gin.New()
	r.GET("/", m.Middleware(), func(c *gin.Context) {
		v, err := sessions.TryGet(c)
		assert.NoError(t, err)
		assert.Nil(t, v)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
}
