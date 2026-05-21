package verifycode

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Purpose 标识验证码用途。
type Purpose string

const (
	PurposeRegister      Purpose = "register"
	PurposeLogin         Purpose = "login"
	PurposePasswordReset Purpose = "password_reset"
	PurposeBindEmail     Purpose = "bind_email"
)

// 常量。
const (
	CodeLength      = 6
	DefaultTTL      = 5 * time.Minute
	DefaultThrottle = 60 * time.Second
)

// Store 是验证码存储抽象。
type Store interface {
	Generate(ctx context.Context, purpose Purpose, email, ip string) (code string, err error)
	Verify(ctx context.Context, purpose Purpose, email, code string) error
}

// store Redis 实现。
type store struct {
	rdb      *redis.Client
	clock    clock.Clock
	log      *zap.Logger
	ttl      time.Duration
	throttle time.Duration
}

// Config 是 New 的参数。
type Config struct {
	TTL      time.Duration // 0 走 DefaultTTL
	Throttle time.Duration // 0 走 DefaultThrottle
}

// New 构造 Store。
func New(rdb *redis.Client, c clock.Clock, log *zap.Logger, cfg Config) Store {
	if c == nil {
		c = clock.Real
	}
	if log == nil {
		log = zap.NewNop()
	}
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	throttle := cfg.Throttle
	if throttle <= 0 {
		throttle = DefaultThrottle
	}
	return &store{rdb: rdb, clock: c, log: log, ttl: ttl, throttle: throttle}
}

func codeKey(purpose Purpose, email string) string {
	return fmt.Sprintf("verify_code:%s:%s", purpose, email)
}

func throttleKey(purpose Purpose, email string) string {
	return fmt.Sprintf("verify_code:throttle:%s:%s", purpose, email)
}

// Generate 生成 6 位数字码并写入 Redis。
func (s *store) Generate(ctx context.Context, purpose Purpose, email, ip string) (string, error) {
	if email == "" {
		return "", apierr.New(apierr.CodeInvalidParam, "email 为空")
	}
	// 节流检查
	exists, err := s.rdb.Exists(ctx, throttleKey(purpose, email)).Result()
	if err != nil {
		return "", apierr.Wrap(apierr.CodeCache, "verify code throttle check", err)
	}
	if exists > 0 {
		return "", apierr.New(apierr.CodeRateLimitUser, "发送过于频繁,请稍后再试")
	}
	code, err := randDigits(CodeLength)
	if err != nil {
		return "", apierr.Wrap(apierr.CodeInternal, "verify code rand", err)
	}
	if err := s.rdb.Set(ctx, codeKey(purpose, email), code, s.ttl).Err(); err != nil {
		return "", apierr.Wrap(apierr.CodeCache, "verify code set", err)
	}
	if err := s.rdb.Set(ctx, throttleKey(purpose, email), "1", s.throttle).Err(); err != nil {
		s.log.Warn("verifycode: throttle set failed", zap.Error(err))
	}
	return code, nil
}

// Verify 校验验证码并消费。
func (s *store) Verify(ctx context.Context, purpose Purpose, email, code string) error {
	if email == "" || code == "" {
		return apierr.New(apierr.CodeCaptchaInvalid, "验证码错误或已过期")
	}
	got, err := s.rdb.Get(ctx, codeKey(purpose, email)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return apierr.New(apierr.CodeCaptchaInvalid, "验证码错误或已过期")
		}
		return apierr.Wrap(apierr.CodeCache, "verify code get", err)
	}
	if subtle.ConstantTimeCompare([]byte(got), []byte(code)) != 1 {
		return apierr.New(apierr.CodeCaptchaInvalid, "验证码错误或已过期")
	}
	// 命中后删除
	if err := s.rdb.Del(ctx, codeKey(purpose, email)).Err(); err != nil {
		s.log.Warn("verifycode: del failed", zap.Error(err))
	}
	return nil
}

// randDigits 返回 n 位 ASCII 数字字符串。
func randDigits(n int) (string, error) {
	buf := make([]byte, n)
	max := big.NewInt(10)
	for i := 0; i < n; i++ {
		x, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = byte('0' + x.Int64())
	}
	return string(buf), nil
}
