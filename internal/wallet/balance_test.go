package wallet

import (
	"testing"
)

func TestBalance_RedisMiss_FallbackDB_RefillsRedis(t *testing.T) {
	s, rdb := newTestStore(t, true)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 1001)
	// DB 直接改 balance(模拟 Lua 路径没走通的情况)
	s.db.Exec("UPDATE wallets SET quota_balance = ? WHERE id = ?", 42000, w.ID)
	// 删除 Redis key 模拟 miss
	rdb.Del(ctx(), walletRedisKey(OwnerTypeUser, 1001))

	got, err := s.Balance(ctx(), w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42000 {
		t.Fatalf("balance = %d, want 42000", got)
	}
	// 现在 Redis 应该被回填了
	v, err := rdb.HGet(ctx(), walletRedisKey(OwnerTypeUser, 1001), hashFieldBalance).Result()
	if err != nil {
		t.Fatalf("redis HGet after fill: %v", err)
	}
	if v != "42000" {
		t.Fatalf("redis balance after refill = %s, want 42000", v)
	}
}

func TestBalance_RedisHit(t *testing.T) {
	s, rdb := newTestStore(t, true)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 1002)
	// 直接覆盖 Redis 模拟 Lua 已扣
	rdb.HSet(ctx(), walletRedisKey(OwnerTypeUser, 1002), hashFieldBalance, "777")
	got, err := s.Balance(ctx(), w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != 777 {
		t.Fatalf("balance = %d, want 777", got)
	}
}

func TestBalance_HashInitialized_OnGetOrCreate(t *testing.T) {
	s, rdb := newTestStore(t, true)
	_, _ = s.GetOrCreate(ctx(), OwnerTypeUser, 1003)
	key := walletRedisKey(OwnerTypeUser, 1003)
	if got, err := rdb.HGet(ctx(), key, hashFieldBalance).Result(); err != nil || got != "0" {
		t.Fatalf("balance not initialized; got=%s err=%v", got, err)
	}
	if got, err := rdb.HGet(ctx(), key, hashFieldReserved).Result(); err != nil || got != "0" {
		t.Fatalf("reserved not initialized; got=%s err=%v", got, err)
	}
}

func TestBalance_NoRedis_FallsBackToDB(t *testing.T) {
	s, _ := newTestStore(t, false)
	w, _ := s.GetOrCreate(ctx(), OwnerTypeUser, 1004)
	s.db.Exec("UPDATE wallets SET quota_balance = ? WHERE id = ?", 99, w.ID)
	got, err := s.Balance(ctx(), w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != 99 {
		t.Fatalf("balance = %d, want 99", got)
	}
}
