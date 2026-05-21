// Package cache 封装 Redis 客户端。
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ijry/pro-api/internal/app/config"
	"github.com/redis/go-redis/v9"
)

// NewClient 用 cfg 创建并 Ping 一次。
func NewClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := rc.Ping(cctx).Err(); err != nil {
		_ = rc.Close()
		return nil, fmt.Errorf("cache: redis ping %s: %w", cfg.Addr, err)
	}
	return rc, nil
}
