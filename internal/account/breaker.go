package account

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type breakerImpl struct {
	repo  Repo
	cache *redis.Client
	log   *zap.Logger
	stop  chan struct{}
}

// NewBreaker 构造 Breaker。cache/log 可为 nil(单元测试无需 Redis/日志)。
func NewBreaker(r Repo, c *redis.Client, log *zap.Logger) Breaker {
	return &breakerImpl{repo: r, cache: c, log: log, stop: make(chan struct{})}
}

func (b *breakerImpl) MarkCooldown(ctx context.Context, id int64, until time.Time, reason string) error {
	a, err := b.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	a.Status = StatusCooldown
	a.CooldownUntil = &until
	a.LastFailureReason = trunc(reason, 256)
	now := time.Now().UTC()
	a.LastFailureAt = &now
	if err := b.repo.Update(ctx, a); err != nil {
		return err
	}
	return b.repo.AppendEvent(ctx, id, "cooldown_entered", map[string]any{"until": until, "reason": reason})
}

func (b *breakerImpl) MarkExpired(ctx context.Context, id int64, reason string) error {
	return b.setStatus(ctx, id, StatusExpired, reason, "refresh_failed")
}

func (b *breakerImpl) MarkInvalid(ctx context.Context, id int64, reason string) error {
	return b.setStatus(ctx, id, StatusInvalid, reason, "marked_invalid")
}

func (b *breakerImpl) setStatus(ctx context.Context, id int64, st Status, reason, evType string) error {
	a, err := b.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	a.Status = st
	a.LastFailureReason = trunc(reason, 256)
	now := time.Now().UTC()
	a.LastFailureAt = &now
	if err := b.repo.Update(ctx, a); err != nil {
		return err
	}
	return b.repo.AppendEvent(ctx, id, evType, map[string]any{"reason": reason})
}

func (b *breakerImpl) IncConsecFailure(ctx context.Context, id int64) (int, error) {
	a, err := b.repo.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	a.ConsecFailures++
	if err := b.repo.Update(ctx, a); err != nil {
		return 0, err
	}
	return a.ConsecFailures, nil
}

func (b *breakerImpl) Run(ctx context.Context) error {
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-b.stop:
			return nil
		case <-tick.C:
			n, err := b.RunReaperOnce(ctx)
			if err != nil && b.log != nil {
				b.log.Warn("breaker reaper", zap.Error(err))
			}
			if n > 0 && b.log != nil {
				b.log.Info("breaker reaped", zap.Int("count", n))
			}
		}
	}
}

// RunReaperOnce 把 cooldown_until 已过期的账号恢复成 active。
// 使用 Repo.Reactivate(map-based)而非 Update,避免 GORM 跳过 status=0 / cooldown_until=nil 的零值。
func (b *breakerImpl) RunReaperOnce(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	list, err := b.repo.ListForReaper(ctx, now, 100)
	if err != nil {
		return 0, err
	}
	for _, a := range list {
		_ = b.repo.Reactivate(ctx, a.ID)
		_ = b.repo.AppendEvent(ctx, a.ID, "cooldown_exited", nil)
	}
	return len(list), nil
}

func (b *breakerImpl) Close() error {
	close(b.stop)
	return nil
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
