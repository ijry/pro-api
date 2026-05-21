package cache

import (
	"context"
	"testing"
	"time"

	"github.com/proapi/proapi/internal/app/config"
)

func TestNewClient_BadAddr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := NewClient(ctx, config.RedisConfig{Addr: "127.0.0.1:1"}); err == nil {
		t.Fatal("want error on unreachable redis")
	}
}
