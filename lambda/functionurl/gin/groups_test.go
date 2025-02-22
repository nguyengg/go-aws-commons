package ginadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nguyengg/go-aws-commons/lambda/functionurl/gin/rules"
	"github.com/stretchr/testify/assert"
)

func TestRequireGroupMembership_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/", RequireGroupMembership(func(c *gin.Context) (authenticated bool, groups rules.Groups) {
		return false, nil
	}, rules.OneOf("a", "b")))
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireGroupMembership_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/", RequireGroupMembership(func(c *gin.Context) (authenticated bool, groups rules.Groups) {
		return true, []string{"c"}
	}, rules.OneOf("a", "b")))
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireGroupMembership_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/", RequireGroupMembership(func(c *gin.Context) (authenticated bool, groups rules.Groups) {
		return true, []string{"a"}
	}, rules.OneOf("a", "b")))
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}
