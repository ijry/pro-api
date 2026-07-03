//go:build embed

package static

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterEmbedded_RealAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterEmbedded(r); err != nil {
		t.Fatalf("RegisterEmbedded: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/docs/" {
		t.Fatalf("root redirect code=%d location=%q", rec.Code, rec.Header().Get("Location"))
	}

	for _, path := range []string{"/admin/", "/user/", "/docs/"} {
		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "<html") {
			t.Fatalf("%s did not return HTML", path)
		}
	}
}
