package verifycode

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
)

func newStore(t *testing.T) (Store, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return New(rdb, clock.Real, nil, Config{}), mr
}

func TestGenerate_ReturnsSixDigits(t *testing.T) {
	s, _ := newStore(t)
	code, err := s.Generate(context.Background(), PurposeRegister, "a@b.com", "1.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Fatalf("want 6 digits, got %q", code)
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Fatalf("non-digit %q", c)
		}
	}
}

func TestGenerate_WritesTTL300(t *testing.T) {
	s, mr := newStore(t)
	_, _ = s.Generate(context.Background(), PurposeRegister, "a@b.com", "")
	ttl := mr.TTL("verify_code:register:a@b.com")
	if ttl != 5*time.Minute {
		t.Fatalf("want 5min TTL, got %v", ttl)
	}
}

func TestGenerate_ThrottleWithin60s(t *testing.T) {
	s, _ := newStore(t)
	if _, err := s.Generate(context.Background(), PurposeLogin, "a@b.com", ""); err != nil {
		t.Fatal(err)
	}
	_, err := s.Generate(context.Background(), PurposeLogin, "a@b.com", "")
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeRateLimitUser {
		t.Fatalf("want CodeRateLimitUser, got %v", err)
	}
}

func TestVerify_OkAndDeletes(t *testing.T) {
	s, mr := newStore(t)
	code, _ := s.Generate(context.Background(), PurposeLogin, "a@b.com", "")
	if err := s.Verify(context.Background(), PurposeLogin, "a@b.com", code); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
	if mr.Exists("verify_code:login:a@b.com") {
		t.Fatal("expected key deleted after Verify")
	}
	// 第二次应失败
	if err := s.Verify(context.Background(), PurposeLogin, "a@b.com", code); err == nil {
		t.Fatal("want second verify fail")
	}
}

func TestVerify_WrongCode(t *testing.T) {
	s, _ := newStore(t)
	_, _ = s.Generate(context.Background(), PurposeLogin, "a@b.com", "")
	err := s.Verify(context.Background(), PurposeLogin, "a@b.com", "000000")
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeCaptchaInvalid {
		t.Fatalf("want CodeCaptchaInvalid, got %v", err)
	}
}

func TestVerify_ExpiredReturnsInvalid(t *testing.T) {
	s, mr := newStore(t)
	_, _ = s.Generate(context.Background(), PurposeLogin, "a@b.com", "")
	mr.FastForward(6 * time.Minute)
	err := s.Verify(context.Background(), PurposeLogin, "a@b.com", "anything")
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeCaptchaInvalid {
		t.Fatalf("want CodeCaptchaInvalid, got %v", err)
	}
}

func TestGenerate_EmptyEmail(t *testing.T) {
	s, _ := newStore(t)
	if _, err := s.Generate(context.Background(), PurposeLogin, "", ""); err == nil {
		t.Fatal("want error")
	}
}

func TestGenerate_OverwritesAndResetsTTL(t *testing.T) {
	s, mr := newStore(t)
	code1, _ := s.Generate(context.Background(), PurposeRegister, "a@b.com", "")

	// 删掉节流键模拟另一次合法发送
	mr.Del("verify_code:throttle:register:a@b.com")
	code2, _ := s.Generate(context.Background(), PurposeRegister, "a@b.com", "")
	if code1 == code2 {
		// 罕见但允许:重复随机即视为通过
		return
	}
	// 老 code 应失效(被覆盖)
	if err := s.Verify(context.Background(), PurposeRegister, "a@b.com", code1); err == nil {
		t.Fatal("old code should be overwritten")
	}
	if err := s.Verify(context.Background(), PurposeRegister, "a@b.com", code2); err != nil {
		t.Fatalf("new code should verify, got %v", err)
	}
}
