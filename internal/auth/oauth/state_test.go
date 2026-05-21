package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
)

func newStateStore(t *testing.T) (StateStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewStateStore(rdb, clock.Real), mr
}

func TestState_IssueAndConsume(t *testing.T) {
	s, _ := newStateStore(t)
	payload, _ := json.Marshal(map[string]any{"redirect": "/x"})
	state, err := s.Issue(context.Background(), "github", payload)
	if err != nil {
		t.Fatal(err)
	}
	if state == "" {
		t.Fatal("want state")
	}
	got, err := s.Consume(context.Background(), "github", state)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("payload mismatch: %s vs %s", got, payload)
	}
}

func TestState_Consume_OnceOnly(t *testing.T) {
	s, _ := newStateStore(t)
	state, _ := s.Issue(context.Background(), "github", nil)
	_, _ = s.Consume(context.Background(), "github", state)
	_, err := s.Consume(context.Background(), "github", state)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeCaptchaInvalid {
		t.Fatalf("want CodeCaptchaInvalid, got %v", err)
	}
}

func TestState_Consume_ExpiredAfter10Min(t *testing.T) {
	s, mr := newStateStore(t)
	state, _ := s.Issue(context.Background(), "github", nil)
	mr.FastForward(11 * time.Minute)
	_, err := s.Consume(context.Background(), "github", state)
	if err == nil {
		t.Fatal("want expired")
	}
}

func TestState_Consume_WrongProvider(t *testing.T) {
	s, _ := newStateStore(t)
	state, _ := s.Issue(context.Background(), "github", nil)
	_, err := s.Consume(context.Background(), "google", state)
	if err == nil {
		t.Fatal("want fail for wrong provider")
	}
}
