package preconditions

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestIfModifiedSince(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name                string
		headers             map[string]string
		t                   time.Time
		wantIgnored         bool
		wantIsModifiedSince bool
	}{
		{
			name:                "ignored due to zero time",
			t:                   time.Time{},
			wantIgnored:         true,
			wantIsModifiedSince: false,
		},
		{
			name:                "ignored due to If-None-Match presence",
			headers:             map[string]string{"If-None-Match": "*"},
			t:                   now,
			wantIgnored:         true,
			wantIsModifiedSince: false,
		},
		{
			name:                "ignored due to If-Modified-Since absence",
			headers:             map[string]string{},
			t:                   now,
			wantIgnored:         true,
			wantIsModifiedSince: false,
		},
		{
			name:                "ignored due to invalid If-Modified-Since",
			headers:             map[string]string{"If-Modified-Since": "hello, world!"},
			t:                   now,
			wantIgnored:         true,
			wantIsModifiedSince: false,
		},
		{
			name:                "not ignored; isModifiedSince is true",
			headers:             map[string]string{"If-Modified-Since": now.Add(-1 * time.Hour).UTC().Format(http.TimeFormat)},
			t:                   now,
			wantIgnored:         false,
			wantIsModifiedSince: true,
		},
		{
			name:                "not ignored; isModifiedSince is false",
			headers:             map[string]string{"If-Modified-Since": now.Add(1 * time.Hour).UTC().Format(http.TimeFormat)},
			t:                   now,
			wantIgnored:         false,
			wantIsModifiedSince: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", func(c *gin.Context) {
				ignored, isModifiedSince := IfModifiedSince(c, tt.t)
				assert.Equalf(t, tt.wantIgnored, ignored, "wantIgnored (%t) vs ignored (%t)", tt.wantIgnored, ignored)
				assert.Equalf(t, tt.wantIsModifiedSince, isModifiedSince, "wantIsUnmodifiedSince (%t) vs modifiedSince (%t)", tt.wantIsModifiedSince, isModifiedSince)
				c.Status(200)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		})
	}
}

func TestIfUnmodifiedSince(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name                  string
		headers               map[string]string
		t                     time.Time
		wantIgnored           bool
		wantIsUnmodifiedSince bool
	}{
		{
			name:                  "ignored due to zero time",
			t:                     time.Time{},
			wantIgnored:           true,
			wantIsUnmodifiedSince: false,
		},
		{
			name:                  "ignored due to If-Match presence",
			headers:               map[string]string{"If-Match": "*"},
			t:                     now,
			wantIgnored:           true,
			wantIsUnmodifiedSince: false,
		},
		{
			name:                  "ignored due to If-Unmodified-Since absence",
			headers:               map[string]string{},
			t:                     now,
			wantIgnored:           true,
			wantIsUnmodifiedSince: false,
		},
		{
			name:                  "ignored due to invalid If-Unmodified-Since",
			headers:               map[string]string{"If-Unmodified-Since": "hello, world!"},
			t:                     now,
			wantIgnored:           true,
			wantIsUnmodifiedSince: false,
		},
		{
			name:                  "not ignored; isUnmodifiedSince is true",
			headers:               map[string]string{"If-Unmodified-Since": now.Add(1 * time.Hour).UTC().Format(http.TimeFormat)},
			t:                     now,
			wantIgnored:           false,
			wantIsUnmodifiedSince: true,
		},
		{
			name:                  "not ignored; isUnmodifiedSince is false",
			headers:               map[string]string{"If-Unmodified-Since": now.Add(-1 * time.Hour).UTC().Format(http.TimeFormat)},
			t:                     now,
			wantIgnored:           false,
			wantIsUnmodifiedSince: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", func(c *gin.Context) {
				ignored, isUnmodifiedSince := IfUnmodifiedSince(c, tt.t)
				assert.Equalf(t, tt.wantIgnored, ignored, "wantIgnored (%t) vs ignored (%t)", tt.wantIgnored, ignored)
				assert.Equalf(t, tt.wantIsUnmodifiedSince, isUnmodifiedSince, "wantIsUnmodifiedSince (%t) vs isUnmodifiedSince (%t)", tt.wantIsUnmodifiedSince, isUnmodifiedSince)
				c.Status(200)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		})
	}
}
