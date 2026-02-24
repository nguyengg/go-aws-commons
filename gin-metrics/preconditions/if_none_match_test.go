package preconditions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_IfNoneMatchMatches(t *testing.T) {
	tests := []struct {
		name              string
		ifNoneMatchHeader []string
		weakETagValue     string
		strongETagValue   string
		noneMatches       bool
	}{
		/*
			Weak comparison (https://www.rfc-editor.org/rfc/rfc9110.html#name-comparison-2):
			0	W/"1"	W/"1"	match
			1	W/"1"	W/"2"	no match
			2	W/"1"	"1"	    match
			3	"1"	    "1"		match
			4 	"1"		"2"		no match
		*/
		{
			name:              `0. W/"1" vs. W/"1"`,
			ifNoneMatchHeader: []string{`W/"1234"`},
			weakETagValue:     `"1234"`,
			noneMatches:       false, // noneMatches will be the opposite of the table above.
		},
		{
			name:              `1. W/"1" vs. W/"2"`,
			ifNoneMatchHeader: []string{`W/"1"`},
			weakETagValue:     `"2"`,
			noneMatches:       true,
		},
		{
			name:              `2. W/"1" vs. "1"`,
			ifNoneMatchHeader: []string{`W/"1"`},
			strongETagValue:   `"1"`,
			noneMatches:       false,
		},
		{
			name:              `3. "1" vs. "1"`,
			ifNoneMatchHeader: []string{`"1"`},
			strongETagValue:   `"1"`,
			noneMatches:       false,
		},
		{
			name:              `4. "1" vs. "2"`,
			ifNoneMatchHeader: []string{`"1"`},
			strongETagValue:   `"2"`,
			noneMatches:       true,
		},
		{
			name:              "any none matches",
			ifNoneMatchHeader: []string{`*`},
			weakETagValue:     `"1"`,
			noneMatches:       false, // * will cause all tag values to return false for NoneMatch
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", New(), func(c *gin.Context) {
				var etag ETag
				if tt.weakETagValue != "" {
					etag = NewWeakETag(tt.weakETagValue)
				} else {
					etag = NewStrongETag(tt.strongETagValue)
				}
				exists, matches, err := IfNoneMatch(c, etag)
				assert.Nil(t, err)
				assert.Truef(t, exists, "If-None-Match should exist")
				assert.Equalf(t, tt.noneMatches, matches, "%v vs. %s noneMatches: expected %t, got %t", tt.ifNoneMatchHeader, etag, tt.noneMatches, matches)
				c.Status(200)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header["If-None-Match"] = tt.ifNoneMatchHeader
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		})
	}
}
