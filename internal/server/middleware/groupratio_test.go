package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// MockTokenView is a minimal mock of token.View for testing.
type MockTokenView struct {
	GroupID int64
}

// GetGroupID returns the group ID.
func (m *MockTokenView) GetGroupID() int64 {
	return m.GroupID
}

func TestGroupRatioMiddleware_SetsRatio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	lookup := func(_ context.Context, gid int64) (float64, error) {
		if gid == 5 {
			return 2.0, nil
		}
		return 1.0, nil
	}

	var gotRatio float64
	r.GET("/test",
		func(c *gin.Context) {
			c.Set("proapi:group_id", int64(5))
			c.Next()
		},
		GroupRatioMiddleware(lookup),
		func(c *gin.Context) {
			v, _ := c.Get(CtxKeyBillingGroupRatio)
			if f, ok := v.(float64); ok {
				gotRatio = f
			}
			c.Status(200)
		},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if gotRatio != 2.0 {
		t.Errorf("billing ratio: want 2.0, got %v", gotRatio)
	}
}

func TestGroupRatioMiddleware_TokenGroupOverridesUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	lookup := func(_ context.Context, gid int64) (float64, error) {
		return float64(gid) * 0.5, nil
	}

	var gotGroupID int64
	r.GET("/test",
		func(c *gin.Context) {
			c.Set("proapi:group_id", int64(3))
			c.Set("proapi:token", &MockTokenView{GroupID: 7})
			c.Next()
		},
		GroupRatioMiddleware(lookup),
		func(c *gin.Context) {
			// Extract group ID from context the same way groupIDFromContext does.
			v, _ := c.Get("proapi:group_id")
			if id, ok := v.(int64); ok {
				gotGroupID = id
			}
			c.Status(200)
		},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if gotGroupID != 7 {
		t.Errorf("want group_id=7 (token override), got %d", gotGroupID)
	}
}
