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
		if html := name + ".html"; fileExists(site.FS, html) {
			return html, true
		}
		if index := path.Join(name, "index.html"); fileExists(site.FS, index) {
			return index, true
		}
		// VitePress cleanUrls emit page.html; unknown docs paths fall back to the docs shell.
		return site.Index, true
	}
	// Admin and user are SPAs; unknown in-app routes must return index.html.
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
