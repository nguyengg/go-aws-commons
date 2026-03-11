package sessions_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	sessions "github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions"
	"github.com/nguyengg/go-aws-commons/ddb-mapper/gin-sessions/csrf"
	"github.com/stretchr/testify/require"
)

func TestManager_ValidateCSRF(t *testing.T) {
	m, _ := setup(t)

	r := gin.New()

	// GET "/" will generate the session Id and CSRF csrfToken.
	r.GET("/", func(c *gin.Context) {
		_, err := m.Get(c)
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}

		if err = m.Save(c); err != nil {
			_ = c.AbortWithError(500, err)
			return
		}

		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)

	// save the CSRF csrfToken so that it can be sent back as both cookie and header for double-submit.
	// the "Set-Cookie" will have multiple values for sid and __Host-csrf.
	var (
		vs             = w.Header().Values("Set-Cookie")
		sid, csrfToken string
	)
	require.Len(t, vs, 2)
	for _, c := range vs {
		switch {
		case strings.HasPrefix(c, "sid="):
			sid = strings.TrimPrefix(c, "sid=")
			sid = sid[:strings.Index(sid, ";")]
		case strings.HasPrefix(c, "__Host-csrf="):
			csrfToken = strings.TrimPrefix(c, "__Host-csrf=")
			csrfToken = csrfToken[:strings.Index(csrfToken, ";")]
		}
	}
	require.NotEmpty(t, csrfToken)

	// POST "/" will validate the CSRF csrfToken.
	r.POST("/", m.ValidateCSRF(csrf.DoubleSubmit(csrf.FromCookie(), csrf.FromHeader())), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// since we have CSRF csrfToken in both cookie and header, the request gets 200.
	log.Printf("%s, %s", sid, csrfToken)
	req, _ = http.NewRequest("POST", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  csrf.DefaultCookieName,
		Value: csrfToken,
	})
	req.AddCookie(&http.Cookie{
		Name:  sessions.DefaultSessionIdCookieName,
		Value: sid,
	})
	req.Header.Add(csrf.DefaultHeaderName, csrfToken)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
}
