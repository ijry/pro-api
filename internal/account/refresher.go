package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ErrRefreshLockHeld 表示另一个进程/协程正在 refresh 同一个账号。
var ErrRefreshLockHeld = errors.New("refresh lock held")

type refresherImpl struct {
	repo  Repo
	cache *redis.Client
	oauth OAuthFlow
	log   *zap.Logger
	stop  chan struct{}
}

// NewRefresher 构造 Refresher。cache 用于 Redis SET NX 分布式锁(避免多副本同时刷一个账号);
// 测试可传 nil cache(将跳过加锁直接刷)。
func NewRefresher(r Repo, c *redis.Client, o OAuthFlow, log *zap.Logger) Refresher {
	return &refresherImpl{repo: r, cache: c, oauth: o, log: log, stop: make(chan struct{})}
}

func (r *refresherImpl) Run(ctx context.Context) error {
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.stop:
			return nil
		case <-tick.C:
			before := time.Now().Add(5 * time.Minute)
			list, err := r.repo.ListForRefresher(ctx, before, 100)
			if err != nil {
				if r.log != nil {
					r.log.Warn("refresher list", zap.Error(err))
				}
				continue
			}
			for _, a := range list {
				go func(id int64) { _ = r.RefreshOne(ctx, id) }(a.ID)
			}
		}
	}
}

func (r *refresherImpl) RefreshOne(ctx context.Context, accountID int64) error {
	if r.cache != nil {
		lockKey := fmt.Sprintf("account:lock:refresh:%d", accountID)
		ok, err := r.cache.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
		if err != nil {
			return err
		}
		if !ok {
			return ErrRefreshLockHeld
		}
		defer r.cache.Del(ctx, lockKey)
	}

	a, err := r.repo.Get(ctx, accountID)
	if err != nil {
		return err
	}
	if a.CredType != "oauth" || a.Cred.RefreshToken == "" {
		return nil
	}
	newCred, err := r.oauth.ExchangeRefreshToken(ctx, a.Provider, a.Cred.RefreshToken)
	if err != nil {
		a.RefreshTokenValid = 2
		_ = r.repo.Update(ctx, a)
		_ = r.repo.AppendEvent(ctx, accountID, "refresh_failed", map[string]any{"err": err.Error()})
		return err
	}
	a.Cred = *newCred
	a.RefreshTokenValid = 1
	now := time.Now().UTC()
	a.LastRefreshedAt = &now
	if !newCred.ExpiresAt.IsZero() {
		t := newCred.ExpiresAt
		a.AccessTokenExpiresAt = &t
	}
	if err := r.repo.Update(ctx, a); err != nil {
		return err
	}
	return r.repo.AppendEvent(ctx, accountID, "token_refreshed", nil)
}
