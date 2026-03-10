package sessions_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/ddb-mapper"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Get(t *testing.T) {
	m, _ := setup(t)

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		// Get will create a new session Id, and Save will store it.
		s, err := m.Get(c)
		assert.NoError(t, err)

		s.User = "me"
		assert.NoError(t, m.Save(c))

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

func TestManager_Regenerate(t *testing.T) {
	m, _ := setup(t)

	// Regenerate will change the Id and Version, but nothing else.
	original := &Session{ID: "my-session-id", User: "me", Version: 6}
	_, err := ddb.Put(t.Context(), original, func(opts *config.PutOptions) {
		opts.DisableOptimisticLocking = true
	})
	require.NoError(t, err)

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		s, err := m.Regenerate(c)
		require.NoError(t, err)
		require.NoError(t, m.Save(c))
		c.String(200, s.ID)
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

func TestManager_Destroy(t *testing.T) {
	m, _ := setup(t)

	_, err := ddb.Put(t.Context(), &Session{ID: "my-session-id", User: "me"})
	require.NoError(t, err)

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		require.NoError(t, m.Destroy(c))
		require.NoError(t, m.Destroy(c))
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

func TestManager_TryGet(t *testing.T) {
	m, _ := setup(t)

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		s, err := m.TryGet(c)
		assert.NoError(t, err)
		assert.Nil(t, s)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
}
