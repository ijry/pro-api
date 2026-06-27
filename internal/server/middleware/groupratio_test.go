package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/ctxkeys"
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
			c.Set("proapi:token", &MockTokenView{GroupID: 5})
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

func TestGroupRatioMiddleware_TokenGroupID_WrittenToContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	lookup := func(_ context.Context, gid int64) (float64, error) {
		return float64(gid) * 0.5, nil
	}

	var gotGroupID int64
	r.GET("/test",
		func(c *gin.Context) {
			c.Set("proapi:token", &MockTokenView{GroupID: 7})
			c.Next()
		},
		GroupRatioMiddleware(lookup),
		func(c *gin.Context) {
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
		t.Errorf("want group_id=7 (from token), got %d", gotGroupID)
	}
}

func TestGroupRatioMiddleware_TokenGroupID_WrittenToRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var gotGroupID int64
	r.GET("/test",
		func(c *gin.Context) {
			c.Set("proapi:token", &MockTokenView{GroupID: 7})
			c.Next()
		},
		GroupRatioMiddleware(nil),
		func(c *gin.Context) {
			if v := c.Request.Context().Value(ctxkeys.GroupID); v != nil {
				gotGroupID, _ = v.(int64)
			}
			c.Status(200)
		},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if gotGroupID != 7 {
		t.Errorf("want request context group_id=7, got %d", gotGroupID)
	}
}
