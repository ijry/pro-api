package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
)

// StateTTL 默认 10 分钟。
const StateTTL = 10 * time.Minute

// StateStore 是 OAuth state 抽象。
type StateStore interface {
	Issue(ctx context.Context, provider string, payload []byte) (state string, err error)
	Consume(ctx context.Context, provider, state string) (payload []byte, err error)
}

type stateStore struct {
	rdb   *redis.Client
	clock clock.Clock
	ttl   time.Duration
}

// NewStateStore 构造 state 存储。
func NewStateStore(rdb *redis.Client, c clock.Clock) StateStore {
	if c == nil {
		c = clock.Real
	}
	return &stateStore{rdb: rdb, clock: c, ttl: StateTTL}
}

func stateKey(provider, state string) string {
	return fmt.Sprintf("oauth_state:%s:%s", provider, state)
}

func (s *stateStore) Issue(ctx context.Context, provider string, payload []byte) (string, error) {
	if provider == "" {
		return "", apierr.New(apierr.CodeInvalidParam, "provider 为空")
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", apierr.Wrap(apierr.CodeInternal, "oauth state rand", err)
	}
	state := base64.RawURLEncoding.EncodeToString(buf)
	if payload == nil {
		payload = []byte("{}")
	}
	if err := s.rdb.Set(ctx, stateKey(provider, state), payload, s.ttl).Err(); err != nil {
		return "", apierr.Wrap(apierr.CodeCache, "oauth state set", err)
	}
	return state, nil
}

func (s *stateStore) Consume(ctx context.Context, provider, state string) ([]byte, error) {
	if provider == "" || state == "" {
		return nil, apierr.New(apierr.CodeCaptchaInvalid, "state 无效或过期")
	}
	key := stateKey(provider, state)
	v, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apierr.New(apierr.CodeCaptchaInvalid, "state 无效或过期")
		}
		return nil, apierr.Wrap(apierr.CodeCache, "oauth state get", err)
	}
	// 删除防重放
	_ = s.rdb.Del(ctx, key).Err()
	return v, nil
}
