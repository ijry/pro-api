package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// LuaScript 是一个已加载到 Redis 的脚本。
type LuaScript struct {
	name string
	src  string
	sha  string
	rdb  *redis.Client
}

// LoadScript 加载脚本并 SCRIPT LOAD,返回带 SHA 的 LuaScript。
// 同一脚本多次 LoadScript 是幂等的(SHA 由内容决定)。
func LoadScript(ctx context.Context, rdb *redis.Client, name, src string) (*LuaScript, error) {
	sha, err := rdb.ScriptLoad(ctx, src).Result()
	if err != nil {
		return nil, fmt.Errorf("cache: load lua %s: %w", name, err)
	}
	return &LuaScript{name: name, src: src, sha: sha, rdb: rdb}, nil
}

// SHA 返回脚本的 SHA1(40 字符 hex)。
func (l *LuaScript) SHA() string { return l.sha }

// Name 返回脚本名(用于日志/指标)。
func (l *LuaScript) Name() string { return l.name }

// Run 优先 EVALSHA,NOSCRIPT 时自动 fallback 到 EVAL(会自带 LOAD)。
func (l *LuaScript) Run(ctx context.Context, keys []string, args ...any) (any, error) {
	r, err := l.rdb.EvalSha(ctx, l.sha, keys, args...).Result()
	if err == nil {
		return r, nil
	}
	if !isNoScript(err) {
		return nil, fmt.Errorf("cache: evalsha %s: %w", l.name, err)
	}
	r, err = l.rdb.Eval(ctx, l.src, keys, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("cache: eval %s: %w", l.name, err)
	}
	return r, nil
}

// RunInts 是 Run 的便利版本,要求脚本返回 []int64。
func (l *LuaScript) RunInts(ctx context.Context, keys []string, args ...any) ([]int64, error) {
	raw, err := l.Run(ctx, keys, args...)
	if err != nil {
		return nil, err
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("cache: %s: want array, got %T", l.name, raw)
	}
	out := make([]int64, len(arr))
	for i, v := range arr {
		n, ok := v.(int64)
		if !ok {
			return nil, fmt.Errorf("cache: %s: element %d is %T, want int64", l.name, i, v)
		}
		out[i] = n
	}
	return out, nil
}

func isNoScript(err error) bool {
	return err != nil && strings.Contains(err.Error(), "NOSCRIPT")
}
