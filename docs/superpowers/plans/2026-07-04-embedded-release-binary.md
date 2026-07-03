# Embedded Release Binary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce tag-release binaries that include and serve the admin, user, and docs frontends.

**Architecture:** Add a focused `internal/server/static` package that serves prefixed embedded web roots with SPA fallback where needed. Frontend builds remain in their existing locations, then a preparation script copies the built outputs into `internal/server/static/dist` because Go embed patterns can only read files under the package directory. The `embed` build variant embeds that package-local dist tree; the default build variant is a no-op for normal backend development.

**Tech Stack:** Go 1.25, Gin, `io/fs`, `embed`, GitHub Actions, Bash packaging scripts, pnpm-built Vite/VitePress frontends.

## Global Constraints

- Release binaries are built with `go build -tags embed`.
- Admin frontend is served under `/admin`.
- User frontend is served under `/user`.
- Documentation site is served under `/docs`.
- Existing API routes under `/api/*`, `/v1/*`, `/v1beta/*`, `/healthz`, and `/metrics` must keep their behavior.
- Normal non-embed backend builds must not require frontend dist folders.
- Important code comments are required only at build-tag boundary, asset-copy boundary, SPA fallback behavior, and docs fallback behavior.
- Do not introduce Node or pnpm as runtime requirements.

---

## File Structure

- Create `internal/server/static/server.go`: shared static serving implementation, route registration, path sanitization, cache headers, SPA/docs fallback.
- Create `internal/server/static/server_test.go`: tests against in-memory `fstest.MapFS`.
- Create `internal/server/static/assets_embed.go`: `//go:build embed` asset roots and `RegisterEmbedded`.
- Create `internal/server/static/assets_noembed.go`: default `RegisterEmbedded` no-op.
- Create `scripts/prepare-embed-assets.sh`: copy frontend build outputs into `internal/server/static/dist`.
- Modify `cmd/proapi/main.go`: call static registration after API routes.
- Modify `Makefile`: run asset preparation before embedded backend build.
- Modify `.github/workflows/ci.yml`: run asset preparation before CI embedded build.
- Modify `.github/workflows/release.yml`: run asset preparation before release embedded build.
- Modify `scripts/package-release.sh`: include config sample and migrations in release archives.

---

### Task 1: Static Web Root Serving

**Files:**
- Create: `internal/server/static/server.go`
- Create: `internal/server/static/server_test.go`

**Interfaces:**
- Produces: `type Site struct`, `type Mode int`, `const ModeSPA`, `const ModeDocs`, `func RegisterSites(r *gin.Engine, sites []Site) error`
- Later tasks consume `RegisterSites` from embed/noembed variants.

- [ ] **Step 1: Write failing tests for SPA and docs routing**

Create `internal/server/static/server_test.go`:

```go
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
```

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
go test ./internal/server/static
```

Expected: fail because `internal/server/static` package or `RegisterSites` is not defined.

- [ ] **Step 3: Implement static serving module**

Create `internal/server/static/server.go`:

```go
// Package static mounts embedded frontend bundles on the API server.
package static

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Mode int

const (
	ModeSPA Mode = iota
	ModeDocs
)

type Site struct {
	Prefix string
	FS     fs.FS
	Index  string
	Mode   Mode
}

func RegisterSites(r *gin.Engine, sites []Site) error {
	for _, site := range sites {
		if err := validateSite(site); err != nil {
			return err
		}
		registerSite(r, site)
	}
	return nil
}

func validateSite(site Site) error {
	if site.Prefix == "" || !strings.HasPrefix(site.Prefix, "/") {
		return fmt.Errorf("static: invalid prefix %q", site.Prefix)
	}
	if site.FS == nil {
		return fmt.Errorf("static: nil fs for %s", site.Prefix)
	}
	if site.Index == "" {
		return fmt.Errorf("static: empty index for %s", site.Prefix)
	}
	if _, err := fs.Stat(site.FS, site.Index); err != nil {
		return fmt.Errorf("static: %s %s: %w", site.Prefix, site.Index, err)
	}
	return nil
}

func registerSite(r *gin.Engine, site Site) {
	prefix := strings.TrimRight(site.Prefix, "/")
	handler := siteHandler(site)

	r.GET(prefix, func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, prefix+"/")
	})
	r.GET(prefix+"/", handler)
	r.GET(prefix+"/*filepath", handler)
}

func siteHandler(site Site) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := strings.TrimPrefix(c.Param("filepath"), "/")
		if name == "" {
			name = site.Index
		}
		name = cleanAssetPath(name)

		fileName, ok := resolveFile(site, name)
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}
		serveFile(c, site.FS, fileName)
	}
}

func cleanAssetPath(name string) string {
	cleaned := path.Clean("/" + name)
	return strings.TrimPrefix(cleaned, "/")
}

func resolveFile(site Site, name string) (string, bool) {
	if fileExists(site.FS, name) {
		return name, true
	}
	if site.Mode == ModeDocs {
		if index := path.Join(name, "index.html"); fileExists(site.FS, index) {
			return index, true
		}
		// VitePress emits static HTML for docs; missing docs routes fall back to the docs shell.
		return site.Index, true
	}
	// Admin and user are client-side SPAs; unknown in-app routes must return index.html.
	return site.Index, true
}

func fileExists(root fs.FS, name string) bool {
	info, err := fs.Stat(root, name)
	return err == nil && !info.IsDir()
}

func serveFile(c *gin.Context, root fs.FS, name string) {
	file, err := root.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}
	setCacheHeaders(c, name)
	if typ := mime.TypeByExtension(path.Ext(name)); typ != "" {
		c.Header("Content-Type", typ)
	}
	http.ServeContent(c.Writer, c.Request, path.Base(name), info.ModTime(), readSeekCloser{file})
}

func setCacheHeaders(c *gin.Context, name string) {
	if path.Base(name) == "index.html" {
		c.Header("Cache-Control", "no-cache")
		return
	}
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Expires", time.Now().Add(365*24*time.Hour).UTC().Format(http.TimeFormat))
}

type readSeekCloser struct {
	fs.File
}

func (r readSeekCloser) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := r.File.(io.Seeker)
	if !ok {
		return 0, fmt.Errorf("static: file is not seekable")
	}
	return seeker.Seek(offset, whence)
}
```

- [ ] **Step 4: Run static tests**

Run:

```bash
go test ./internal/server/static
```

Expected: PASS.

- [ ] **Step 5: Commit static serving module**

```bash
git add internal/server/static/server.go internal/server/static/server_test.go
git commit -m "feat: add embedded static site router"
```

---

### Task 2: Embed Asset Preparation and Build Variants

**Files:**
- Create: `internal/server/static/assets_embed.go`
- Create: `internal/server/static/assets_noembed.go`
- Create: `scripts/prepare-embed-assets.sh`
- Modify: `Makefile`
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/release.yml`

**Interfaces:**
- Consumes: `func RegisterSites(r *gin.Engine, sites []Site) error`
- Produces: `func RegisterEmbedded(r *gin.Engine) error`
- Produces: package-local generated directories `internal/server/static/dist/admin`, `internal/server/static/dist/user`, `internal/server/static/dist/docs`

- [ ] **Step 1: Add asset preparation script**

Create `scripts/prepare-embed-assets.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target="${root_dir}/internal/server/static/dist"

require_dir() {
  local path="$1"
  if [ ! -d "${path}" ]; then
    echo "missing frontend build output: ${path}" >&2
    exit 1
  fi
}

require_dir "${root_dir}/web/admin/dist"
require_dir "${root_dir}/web/user/dist"
require_dir "${root_dir}/docs-site/.vitepress/dist"

rm -rf "${target}"
mkdir -p "${target}"

# Go embed can only read files below the package directory, so release builds
# copy frontend outputs into this package-local dist tree before compilation.
cp -R "${root_dir}/web/admin/dist" "${target}/admin"
cp -R "${root_dir}/web/user/dist" "${target}/user"
cp -R "${root_dir}/docs-site/.vitepress/dist" "${target}/docs"
```

- [ ] **Step 2: Add embed build variant**

Create `internal/server/static/assets_embed.go`:

```go
//go:build embed

package static

import (
	"embed"
	"io/fs"

	"github.com/gin-gonic/gin"
)

// The dist tree is generated by scripts/prepare-embed-assets.sh before
// `go build -tags embed`; keeping it package-local satisfies Go embed rules.
//
//go:embed dist/admin
var adminDist embed.FS

//go:embed dist/user
var userDist embed.FS

//go:embed dist/docs
var docsDist embed.FS

func RegisterEmbedded(r *gin.Engine) error {
	adminFS, err := fs.Sub(adminDist, "dist/admin")
	if err != nil {
		return err
	}
	userFS, err := fs.Sub(userDist, "dist/user")
	if err != nil {
		return err
	}
	docsFS, err := fs.Sub(docsDist, "dist/docs")
	if err != nil {
		return err
	}
	return RegisterSites(r, []Site{
		{Prefix: "/admin", FS: adminFS, Index: "index.html", Mode: ModeSPA},
		{Prefix: "/user", FS: userFS, Index: "index.html", Mode: ModeSPA},
		{Prefix: "/docs", FS: docsFS, Index: "index.html", Mode: ModeDocs},
	})
}
```

- [ ] **Step 3: Add default no-embed build variant**

Create `internal/server/static/assets_noembed.go`:

```go
//go:build !embed

package static

import "github.com/gin-gonic/gin"

// RegisterEmbedded is a no-op in local backend builds so developers can run the
// API server without first producing frontend dist directories.
func RegisterEmbedded(_ *gin.Engine) error {
	return nil
}
```

- [ ] **Step 4: Wire asset preparation into local build**

Modify `Makefile`.

Change `build-backend` from:

```make
build-backend:
	CGO_ENABLED=0 go build -tags embed -ldflags="-s -w" -o bin/proapi ./cmd/proapi
```

to:

```make
build-backend:
	bash scripts/prepare-embed-assets.sh
	CGO_ENABLED=0 go build -tags embed -ldflags="-s -w" -o bin/proapi ./cmd/proapi
```

- [ ] **Step 5: Wire asset preparation into CI build**

Modify `.github/workflows/ci.yml`.

After:

```yaml
      - run: pnpm -C docs-site build
```

add:

```yaml
      - run: bash scripts/prepare-embed-assets.sh
```

This must run before the existing `go build -tags embed` step.

- [ ] **Step 6: Wire asset preparation into release build**

Modify `.github/workflows/release.yml`.

After:

```yaml
      - name: Build embedded frontend assets
        run: |
          pnpm -C web/admin build
          pnpm -C web/user build
          pnpm -C docs-site build
```

add:

```yaml
      - name: Prepare embedded frontend assets
        run: bash scripts/prepare-embed-assets.sh
```

- [ ] **Step 7: Run default-build tests**

Run:

```bash
go test ./internal/server/static
```

Expected: PASS without frontend dist directories.

- [ ] **Step 8: Commit embed preparation**

```bash
git add internal/server/static/assets_embed.go internal/server/static/assets_noembed.go scripts/prepare-embed-assets.sh Makefile .github/workflows/ci.yml .github/workflows/release.yml
git commit -m "build: prepare frontend assets for embed"
```

---

### Task 3: Wire Embedded Static Routes

**Files:**
- Modify: `cmd/proapi/main.go`

**Interfaces:**
- Consumes: `func static.RegisterEmbedded(r *gin.Engine) error`

- [ ] **Step 1: Import static package**

Modify `cmd/proapi/main.go`.

Add import:

```go
	frontendstatic "github.com/ijry/pro-api/internal/server/static"
```

- [ ] **Step 2: Register embedded routes after API routes**

After this block:

```go
	if err := wireRoutes(ctx, engine, application, log); err != nil {
		return fmt.Errorf("routes: %w", err)
	}
```

add:

```go
	if err := frontendstatic.RegisterEmbedded(engine); err != nil {
		return fmt.Errorf("static assets: %w", err)
	}
```

- [ ] **Step 3: Run default full Go tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Commit route wiring**

```bash
git add cmd/proapi/main.go
git commit -m "feat: wire embedded frontend routes"
```

---

### Task 4: Release Archive Support Files

**Files:**
- Modify: `scripts/package-release.sh`

**Interfaces:**
- Consumes: existing script arguments `goos goarch version binary_path output_dir`
- Produces: archives containing `proapi`, `README.md`, `LICENSE`, `configs/proapi.example.yaml`, and `migrations/`

- [ ] **Step 1: Update packaging script**

Modify `scripts/package-release.sh` after the existing README/LICENSE copies:

```bash
cp "${binary_path}" "${stage_dir}/${binary_name}"
cp LICENSE "${stage_dir}/LICENSE"
cp README.md "${stage_dir}/README.md"

mkdir -p "${stage_dir}/configs"
cp configs/proapi.example.yaml "${stage_dir}/configs/proapi.example.yaml"
cp -R migrations "${stage_dir}/migrations"
```

- [ ] **Step 2: Run shell syntax check**

Run:

```bash
bash -n scripts/package-release.sh
```

Expected: no output and exit code 0.

- [ ] **Step 3: Smoke package a local binary**

Run:

```bash
mkdir -p dist/bin
go build -o dist/bin/proapi ./cmd/proapi
bash scripts/package-release.sh linux amd64 test dist/bin/proapi dist/release
tar -tzf dist/release/proapi_linux_amd64.tar.gz | sort | rg "configs/proapi.example.yaml|migrations/|proapi$|README.md|LICENSE"
```

Expected output includes all five classes of files.

- [ ] **Step 4: Commit packaging changes**

```bash
git add scripts/package-release.sh
git commit -m "chore: include deployment files in release archive"
```

---

### Task 5: Embedded Build Verification

**Files:**
- No source changes unless verification exposes a defect.

**Interfaces:**
- Consumes: static package, asset preparation script, and route wiring from previous tasks.
- Produces: verified embedded binary.

- [ ] **Step 1: Build frontend assets**

Run:

```bash
pnpm -C web install --frozen-lockfile
pnpm -C docs-site install --frozen-lockfile
pnpm -C web/admin build
pnpm -C web/user build
pnpm -C docs-site build
bash scripts/prepare-embed-assets.sh
```

Expected: all commands pass and create `internal/server/static/dist/admin`, `internal/server/static/dist/user`, and `internal/server/static/dist/docs`.

- [ ] **Step 2: Build embedded binary**

Run:

```bash
go build -tags embed -o bin/proapi ./cmd/proapi
```

Expected: PASS.

- [ ] **Step 3: Run complete verification**

Run:

```bash
go test ./...
bash -n scripts/package-release.sh
```

Expected: PASS.

- [ ] **Step 4: Commit any verification fixes**

If verification changes source files:

```bash
git add <changed-files>
git commit -m "fix: complete embedded release verification"
```

If no source files change, skip the commit.

---

## Self-Review

- Spec coverage: route prefixes, non-embed behavior, release packaging, cache behavior, tests, and verification are covered by Tasks 1-5.
- Placeholder scan: the plan contains exact paths, exact commands, and concrete code for each code-writing step.
- Type consistency: `RegisterEmbedded(*gin.Engine) error`, `RegisterSites(*gin.Engine, []Site) error`, `Site`, `ModeSPA`, and `ModeDocs` are defined before later tasks consume them.
- Go embed consistency: frontend assets are copied into `internal/server/static/dist` before embedded builds, so embed patterns stay under their package directory.
