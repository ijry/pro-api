package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// StateData 是一次 PKCE 授权在 Start 与 Callback 之间需要保留的状态。
type StateData struct {
	Provider  string `json:"provider"`
	ChannelID int64  `json:"channel_id"`
	Verifier  string `json:"verifier"`
}

// StateStore 持久化 PKCE state,一次性:Take 成功后立即失效。
type StateStore interface {
	Save(ctx context.Context, state string, d StateData, ttl time.Duration) error
	Take(ctx context.Context, state string) (StateData, error)
}

// ErrStateNotFound 表示 state 不存在或已被消费/过期(可能是 CSRF 或重放)。
var ErrStateNotFound = errors.New("oauth: state not found or expired")

const stateKeyPrefix = "account:oauth:state:"

// RedisStateStore 用 Redis 实现 StateStore(对应路线图 oauth:state:* 规约)。
type RedisStateStore struct{ rdb *redis.Client }

// NewRedisStateStore 构造 Redis 状态存储。
func NewRedisStateStore(rdb *redis.Client) *RedisStateStore { return &RedisStateStore{rdb: rdb} }

func (s *RedisStateStore) Save(ctx context.Context, state string, d StateData, ttl time.Duration) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, stateKeyPrefix+state, b, ttl).Err()
}

func (s *RedisStateStore) Take(ctx context.Context, state string) (StateData, error) {
	key := stateKeyPrefix + state
	b, err := s.rdb.GetDel(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return StateData{}, ErrStateNotFound
	}
	if err != nil {
		return StateData{}, err
	}
	var d StateData
	if err := json.Unmarshal(b, &d); err != nil {
		return StateData{}, err
	}
	return d, nil
}

// MemStateStore 是进程内实现,仅用于测试。
type MemStateStore struct {
	mu sync.Mutex
	m  map[string]StateData
}

// NewMemStateStore 构造内存状态存储。
func NewMemStateStore() *MemStateStore { return &MemStateStore{m: map[string]StateData{}} }

func (s *MemStateStore) Save(_ context.Context, state string, d StateData, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[state] = d
	return nil
}

func (s *MemStateStore) Take(_ context.Context, state string) (StateData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.m[state]
	if !ok {
		return StateData{}, ErrStateNotFound
	}
	delete(s.m, state)
	return d, nil
}
