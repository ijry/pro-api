package wallet

import (
	"context"
	"errors"
	"strconv"

	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Redis hash 字段名。Lua 脚本约定:balance / reserved / consumed。
const (
	hashFieldBalance  = "balance"
	hashFieldReserved = "reserved"
	hashFieldConsumed = "consumed"
)

// Balance 取余额。优先 Redis hash,miss 时回源 DB。
//
// 错误:
//   - wallet 不存在 → CodeNotFound
//   - Redis / DB 双 fail → CodeDatabase
func (s *store) Balance(ctx context.Context, walletID int64) (int64, error) {
	if walletID <= 0 {
		return 0, apierr.New(apierr.CodeInvalidParam, "wallet_id must be > 0")
	}
	if s.rdb != nil {
		// 先查 DB 拿 owner,才能拼出 Redis hash key
		w, err := s.Snapshot(ctx, walletID)
		if err != nil {
			return 0, err
		}
		key := walletRedisKey(w.OwnerType, w.OwnerID)
		raw, err := s.rdb.HGet(ctx, key, hashFieldBalance).Result()
		if err == nil {
			n, perr := strconv.ParseInt(raw, 10, 64)
			if perr == nil {
				return n, nil
			}
		} else if !errors.Is(err, redis.Nil) {
			s.log.Warn("wallet: redis HGet failed, falling back to DB",
				zap.String("key", key), zap.Error(err))
		}
		// miss / parse fail → 回填
		s.ensureRedisHash(ctx, w)
		return w.QuotaBalance, nil
	}
	w, err := s.Snapshot(ctx, walletID)
	if err != nil {
		return 0, err
	}
	return w.QuotaBalance, nil
}

// ensureRedisHash 把 wallet 的 balance / reserved(0)/ consumed 写到 Redis hash。
// 仅在 hash 不存在时写;若已存在,只补缺失字段。
func (s *store) ensureRedisHash(ctx context.Context, w *Wallet) {
	if s.rdb == nil || w == nil {
		return
	}
	key := walletRedisKey(w.OwnerType, w.OwnerID)
	// HSETNX 三个字段;balance 已存在不覆盖(以 Lua 路径维护的为真源)。
	pipe := s.rdb.Pipeline()
	pipe.HSetNX(ctx, key, hashFieldBalance, w.QuotaBalance)
	pipe.HSetNX(ctx, key, hashFieldReserved, 0)
	pipe.HSetNX(ctx, key, hashFieldConsumed, w.QuotaTotalConsumed)
	if _, err := pipe.Exec(ctx); err != nil {
		s.log.Warn("wallet: redis HSetNX failed", zap.String("key", key), zap.Error(err))
	}
}
