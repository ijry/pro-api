package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/pkg/apierr"
)

func setupGin(format string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorResponse(format))
	r.GET("/x", handler)
	return r
}

func TestErrorResponse_OpenAIFormat(t *testing.T) {
	r := setupGin("openai", func(c *gin.Context) {
		SetErr(c, apierr.New(apierr.CodeInvalidToken, "bad token"))
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	var body struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Type != "invalid_request_error" {
		t.Errorf("want invalid_request_error, got %s", body.Error.Type)
	}
	if body.Error.Code != "invalid_api_key" {
		t.Errorf("want invalid_api_key, got %s", body.Error.Code)
	}
}

func TestErrorResponse_JSONFormat(t *testing.T) {
	r := setupGin("json", func(c *gin.Context) {
		SetErr(c, apierr.New(apierr.CodeMissingParam, "需要 email"))
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != int(apierr.CodeMissingParam) {
		t.Errorf("want %d, got %d", apierr.CodeMissingParam, body.Code)
	}
	if body.Message != "需要 email" {
		t.Errorf("got %s", body.Message)
	}
}

func TestErrorResponse_NoError_PassThrough(t *testing.T) {
	r := setupGin("json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestSetErr_AbortsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	called := false
	r.Use(ErrorResponse("json"))
	r.Use(func(c *gin.Context) {
		SetErr(c, apierr.New(apierr.CodeForbidden, "forbidden"))
	})
	r.GET("/x", func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if called {
		t.Fatal("handler should not be reached after SetErr/Abort")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("want 403, got %d", rec.Code)
	}
}
