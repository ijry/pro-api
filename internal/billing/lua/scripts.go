// Package lua 加载并持有 billing 用到的三个 Lua 脚本。
package lua

import (
	_ "embed"
	"fmt"

	"github.com/ijry/pro-api/internal/cache"
	"github.com/redis/go-redis/v9"
	"context"
)

//go:embed reserve.lua
var reserveSrc string

//go:embed commit.lua
var commitSrc string

//go:embed refund.lua
var refundSrc string

// Scripts 持有三个已加载的脚本。
type Scripts struct {
	Reserve *cache.LuaScript
	Commit  *cache.LuaScript
	Refund  *cache.LuaScript
}

// Load 加载脚本到 Redis。
func Load(rdb *redis.Client) (*Scripts, error) {
	ctx := context.Background()
	r, err := cache.LoadScript(ctx, rdb, "billing.reserve", reserveSrc)
	if err != nil {
		return nil, fmt.Errorf("load reserve: %w", err)
	}
	c, err := cache.LoadScript(ctx, rdb, "billing.commit", commitSrc)
	if err != nil {
		return nil, fmt.Errorf("load commit: %w", err)
	}
	f, err := cache.LoadScript(ctx, rdb, "billing.refund", refundSrc)
	if err != nil {
		return nil, fmt.Errorf("load refund: %w", err)
	}
	return &Scripts{Reserve: r, Commit: c, Refund: f}, nil
}

// ReserveSHA / CommitSHA / RefundSHA 用于日志。
func (s *Scripts) ReserveSHA() string { return s.Reserve.SHA() }
func (s *Scripts) CommitSHA() string  { return s.Commit.SHA() }
func (s *Scripts) RefundSHA() string  { return s.Refund.SHA() }
