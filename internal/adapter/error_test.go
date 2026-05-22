package adapter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/pkg/apierr"
)

func mustCode(t *testing.T, err error, want apierr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %d, got nil", want)
	}
	var ae *apierr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apierr.Error, got %T", err)
	}
	if ae.Code != want {
		t.Fatalf("expected code %d, got %d (%s)", want, ae.Code, ae.Message)
	}
}

func TestClassifyHTTPStatus_429(t *testing.T) {
	mustCode(t, adapter.ClassifyHTTPStatus(429, []byte("rate limit")), apierr.CodeUpstreamRateLimit)
}

func TestClassifyHTTPStatus_500(t *testing.T) {
	mustCode(t, adapter.ClassifyHTTPStatus(500, []byte("oops")), apierr.CodeUpstreamError)
}

func TestClassifyHTTPStatus_503(t *testing.T) {
	mustCode(t, adapter.ClassifyHTTPStatus(503, nil), apierr.CodeUpstreamUnavail)
}

func TestClassifyHTTPStatus_401(t *testing.T) {
	mustCode(t, adapter.ClassifyHTTPStatus(401, []byte("bad key")), apierr.CodeChannelMisconfig)
}

func TestClassifyHTTPStatus_402(t *testing.T) {
	mustCode(t, adapter.ClassifyHTTPStatus(402, nil), apierr.CodeUpstreamQuota)
}

func TestClassifyHTTPStatus_400(t *testing.T) {
	mustCode(t, adapter.ClassifyHTTPStatus(400, []byte("bad")), apierr.CodeInvalidParam)
}

func TestClassifyHTTPStatus_ContentFilterByBody(t *testing.T) {
	body := `{"error":{"code":"content_filter","message":"blocked"}}`
	mustCode(t, adapter.ClassifyHTTPStatus(400, []byte(body)), apierr.CodeUpstreamContentFilter)
}

func TestClassifyNetErr_ContextCanceled(t *testing.T) {
	mustCode(t, adapter.ClassifyNetErr(context.Canceled), apierr.CodeUpstreamTimeout)
}

func TestClassifyNetErr_DeadlineExceeded(t *testing.T) {
	mustCode(t, adapter.ClassifyNetErr(context.DeadlineExceeded), apierr.CodeUpstreamTimeout)
}

func TestClassifyNetErr_Nil(t *testing.T) {
	if e := adapter.ClassifyNetErr(nil); e != nil {
		t.Fatalf("expected nil, got %v", e)
	}
}
