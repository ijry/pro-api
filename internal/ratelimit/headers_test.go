package ratelimit

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newGinTestCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestWriteHeaders_Denied_UserRPM(t *testing.T) {
	c, w := newGinTestCtx()
	reset := time.Now().Add(45 * time.Second)
	WriteHeaders(c, Decision{
		Allowed:   false,
		Dimension: DimUserRPM,
		Limit:     60,
		Remaining: 0,
		Reset:     reset,
	})
	if got := w.Header().Get("X-RateLimit-Limit-User-RPM"); got != "60" {
		t.Errorf("Limit-User-RPM = %q", got)
	}
	if got := w.Header().Get("X-RateLimit-Remaining-User-RPM"); got != "0" {
		t.Errorf("Remaining-User-RPM = %q", got)
	}
	if got := w.Header().Get("X-RateLimit-Reset"); got != strconv.FormatInt(reset.Unix(), 10) {
		t.Errorf("Reset = %q", got)
	}
	ra := w.Header().Get("Retry-After")
	if ra == "" {
		t.Fatal("Retry-After missing")
	}
	n, _ := strconv.Atoi(ra)
	if n < 1 {
		t.Errorf("Retry-After < 1: %s", ra)
	}
}

func TestWriteHeaders_Denied_RetryAfter_MinOne(t *testing.T) {
	c, w := newGinTestCtx()
	// reset 已过
	WriteHeaders(c, Decision{
		Allowed:   false,
		Dimension: DimUserRPM,
		Limit:     60,
		Reset:     time.Now().Add(-5 * time.Second),
	})
	ra := w.Header().Get("Retry-After")
	if ra != "1" {
		t.Errorf("want Retry-After=1; got %q", ra)
	}
}

func TestWriteHeaders_Allowed_TightestDim(t *testing.T) {
	c, w := newGinTestCtx()
	reset := time.Now().Add(30 * time.Second)
	WriteHeaders(c, Decision{
		Allowed:   true,
		Dimension: DimTokenRPM,
		Limit:     10,
		Remaining: 7,
		Reset:     reset,
	})
	if w.Header().Get("X-RateLimit-Limit-Token-RPM") != "10" {
		t.Error("missing Limit-Token-RPM")
	}
	if w.Header().Get("X-RateLimit-Remaining-Token-RPM") != "7" {
		t.Error("missing Remaining-Token-RPM=7")
	}
	if w.Header().Get("Retry-After") != "" {
		t.Error("Retry-After should not be set on allowed")
	}
}

func TestWriteHeaders_AllDimensions_HeaderSuffix(t *testing.T) {
	all := []Dimension{DimUserRPM, DimUserTPM, DimTokenRPM, DimTokenTPM, DimIPRPM, DimModelRPM, DimModelTPM}
	for _, dim := range all {
		c, w := newGinTestCtx()
		WriteHeaders(c, Decision{
			Allowed:   false,
			Dimension: dim,
			Limit:     1,
			Reset:     time.Now().Add(time.Minute),
		})
		header := "X-RateLimit-Limit-" + dim.HeaderSuffix()
		if w.Header().Get(header) == "" {
			t.Errorf("dim=%s: missing header %s", dim, header)
		}
	}
}
