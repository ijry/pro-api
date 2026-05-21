package apierr

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew_CarriesAllFields(t *testing.T) {
	e := New(CodeInvalidToken, "令牌无效")
	if e.Code != CodeInvalidToken {
		t.Fatalf("want code %d, got %d", CodeInvalidToken, e.Code)
	}
	if e.Message != "令牌无效" {
		t.Fatalf("want message 令牌无效, got %q", e.Message)
	}
	if e.HTTPStatus != http.StatusUnauthorized {
		t.Fatalf("want HTTP 401, got %d", e.HTTPStatus)
	}
}

func TestError_ErrorString(t *testing.T) {
	e := New(CodeInternal, "出错")
	want := "[10000] 出错"
	if got := e.Error(); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestError_WithDetails(t *testing.T) {
	e := New(CodeEmailRegistered, "邮箱已注册").WithDetails(map[string]any{"field": "email"})
	if e.Details["field"] != "email" {
		t.Fatalf("details not set: %+v", e.Details)
	}
}

func TestError_Is(t *testing.T) {
	a := New(CodeInvalidToken, "x")
	b := New(CodeInvalidToken, "y")
	if !errors.Is(a, b) {
		t.Fatal("want errors.Is to match by code")
	}
	c := New(CodeInternal, "x")
	if errors.Is(a, c) {
		t.Fatal("different codes should not match")
	}
}

func TestHTTPStatus_FixedMapping(t *testing.T) {
	cases := map[Code]int{
		CodeInternal:            http.StatusInternalServerError,
		CodeNotLoggedIn:         http.StatusUnauthorized,
		CodeForbidden:           http.StatusForbidden,
		CodeMissingParam:        http.StatusBadRequest,
		CodeBalanceInsufficient: http.StatusPaymentRequired,
		CodeRateLimitUser:       http.StatusTooManyRequests,
		CodeUpstreamError:       http.StatusBadGateway,
	}
	for code, want := range cases {
		if got := New(code, "").HTTPStatus; got != want {
			t.Errorf("code %d want HTTP %d, got %d", code, want, got)
		}
	}
}
