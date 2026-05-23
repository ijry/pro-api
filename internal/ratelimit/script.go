package ratelimit

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/ijry/pro-api/internal/cache"
	"github.com/redis/go-redis/v9"
)

//go:embed lua/sliding_window.lua
var slidingWindowSrc string

// scripts 保存所有 ratelimit 用到的 LuaScript。
type scripts struct {
	sliding *cache.LuaScript
}

// loadScripts 在 wire 阶段调用一次。
func loadScripts(ctx context.Context, rdb *redis.Client) (*scripts, error) {
	if rdb == nil {
		return nil, fmt.Errorf("ratelimit: redis client is nil")
	}
	sw, err := cache.LoadScript(ctx, rdb, "ratelimit_sliding_window", slidingWindowSrc)
	if err != nil {
		return nil, fmt.Errorf("ratelimit: load sliding_window: %w", err)
	}
	return &scripts{sliding: sw}, nil
}
