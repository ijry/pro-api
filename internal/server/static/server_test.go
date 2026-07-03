package static

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestRegisterSites_SPAFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	err := RegisterSites(r, []Site{{
		Prefix: "/admin",
		FS: fstest.MapFS{
			"index.html":     &fstest.MapFile{Data: []byte("<html>admin</html>")},
			"assets/app.js":  &fstest.MapFile{Data: []byte("console.log('admin')")},
			"assets/app.css": &fstest.MapFile{Data: []byte("body{}")},
		},
		Index: "index.html",
		Mode:  ModeSPA,
	}})
	if err != nil {
		t.Fatalf("RegisterSites: %v", err)
	}

	for _, path := range []string{"/admin/", "/admin/settings/accounts"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "admin") {
			t.Fatalf("%s body = %q", path, rec.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/assets/app.js", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "console.log") {
		t.Fatalf("asset response code=%d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Fatalf("asset Cache-Control = %q", got)
	}
}

func TestRegisterSites_DocsPrefersRealFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	err := RegisterSites(r, []Site{{
		Prefix: "/docs",
		FS: fstest.MapFS{
			"index.html":       &fstest.MapFile{Data: []byte("<html>docs-home</html>")},
			"guide/index.html": &fstest.MapFile{Data: []byte("<html>guide</html>")},
			"assets/style.css": &fstest.MapFile{Data: []byte("body{}")},
		},
		Index: "index.html",
		Mode:  ModeDocs,
	}})
	if err != nil {
		t.Fatalf("RegisterSites: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs/guide/", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "guide") {
		t.Fatalf("guide response code=%d body=%q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/docs/missing/page", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "docs-home") {
		t.Fatalf("missing docs fallback code=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestRegisterSites_DoesNotCaptureAPIRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/health", func(c *gin.Context) { c.String(http.StatusOK, "api") })
	if err := RegisterSites(r, []Site{{
		Prefix: "/admin",
		FS:     fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("<html>admin</html>")}},
		Index:  "index.html",
		Mode:   ModeSPA,
	}}); err != nil {
		t.Fatalf("RegisterSites: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "api" {
		t.Fatalf("api response code=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestRegisterSites_MissingIndexFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	err := RegisterSites(r, []Site{{
		Prefix: "/admin",
		FS:     fstest.MapFS{"asset.js": &fstest.MapFile{Data: []byte("x")}},
		Index:  "index.html",
		Mode:   ModeSPA,
	}})
	if err == nil {
		t.Fatal("expected missing index error")
	}
	if !strings.Contains(err.Error(), fs.ErrNotExist.Error()) && !strings.Contains(err.Error(), "index.html") {
		t.Fatalf("unexpected error: %v", err)
	}
}
