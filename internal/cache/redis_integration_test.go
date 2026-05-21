//go:build integration

package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ijry/pro-api/internal/app/config"
)

func TestNewClient_RealRedis(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Skipf("docker unavailable: %v", err)
	}
	if err := pool.Client.Ping(); err != nil {
		t.Skipf("docker daemon unreachable: %v", err)
	}
	res, err := pool.Run("redis", "7", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Purge(res) })

	pool.MaxWait = 30 * time.Second
	addr := fmt.Sprintf("127.0.0.1:%s", res.GetPort("6379/tcp"))
	if err := pool.Retry(func() error {
		_, err := NewClient(context.Background(), config.RedisConfig{Addr: addr})
		return err
	}); err != nil {
		t.Fatal(err)
	}
}
