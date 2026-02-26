package preconditions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_IfMatchMatches(t *testing.T) {
	tests := []struct {
		name          string
		ifMatchHeader []string
		etag          string
		matches       bool
	}{
		{
			name:          "single v",
			ifMatchHeader: []string{`"1234"`},
			etag:          `"1234"`,
			matches:       true,
		},
		{
			name:          "against multiple values",
			ifMatchHeader: []string{`"2345"`, `"1234"`, `"3456"`},
			etag:          `"1234"`,
			matches:       true,
		},
		{
			name:          "any matches",
			ifMatchHeader: []string{`*`},
			etag:          `"1234"`,
			matches:       true,
		},
		{
			name:          "does not match against single v",
			ifMatchHeader: []string{`"3456"`},
			etag:          `"1234"`,
			matches:       false,
		},
		{
			name:          "does not match against multiple values",
			ifMatchHeader: []string{`"3456"`, `"6789"`},
			etag:          `"1234"`,
			matches:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", New(), func(c *gin.Context) {
				etag, err := ParseETag(tt.etag)
				assert.NoError(t, err)

				exists, matches, err := IfMatch(c, etag.(StrongETag))
				assert.NoError(t, err)
				assert.Truef(t, exists, "If-Match should exist")
				assert.Equalf(t, tt.matches, matches, "%v vs. %s matches: expected %t, got %t", tt.ifMatchHeader, tt.etag, tt.matches, matches)
				c.Status(200)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header["If-Match"] = tt.ifMatchHeader
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		})
	}
}
