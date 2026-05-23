package wallet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IDGenerator 是 wallet 模块对 idgen 的最小依赖。
type IDGenerator interface {
	Generate() int64
}

// Store 是 wallet 模块对外的统一接口。
//
// 设计稿:M1-06 spec §3.1。
type Store interface {
	// GetOrCreate 取一个 wallet;不存在则插入 quota_balance=0 行,自动初始化 Redis hash。
	GetOrCreate(ctx context.Context, ownerType string, ownerID int64) (*Wallet, error)

	// Balance 取实时余额。优先 Redis hash,miss 时回源 DB 并回填。
	Balance(ctx context.Context, walletID int64) (int64, error)

	// Credit 给钱包加额度(管理员加额度 / 兑换码 / 手动充值)。
	// amount > 0,refType 必须是 manual / redeem / refund / adjust。
	Credit(ctx context.Context, walletID int64, amount int64, refType string, refID *int64, desc string, actor int64) error

	// Snapshot 从 DB 全量读取(用于 ledger 页 / 钱包详情)。不读 Redis。
	Snapshot(ctx context.Context, walletID int64) (*Wallet, error)

	// ByOwner 按 owner 取(不存在返回 CodeNotFound)。
	ByOwner(ctx context.Context, ownerType string, ownerID int64) (*Wallet, error)

	// ListLedger 列我的流水(分页 + 过滤)。
	ListLedger(ctx context.Context, filter LedgerFilter) ([]*LedgerEntry, int64, error)

	// ApplyLedgerBatch 单事务写入一批 ledger + 同步钱包累计字段。
	// events 必须同 wallet_id(调用方分组);执行流程见 spec §4.6。
	ApplyLedgerBatch(ctx context.Context, events []LedgerEvent) error

	// Close 关停。
	Close() error
}

// LedgerFilter 是 ListLedger 的过滤条件。
type LedgerFilter struct {
	WalletID int64
	RefType  string     // 空表示不过滤
	Since    *time.Time // created_at >= since(可选)
	Until    *time.Time // created_at < until(可选)
	Page     int
	Size     int
}

// Config 是 New 的参数。
type Config struct {
	DB    *gorm.DB
	Cache *redis.Client
	Log   *zap.Logger
	Clock clock.Clock
	IDGen IDGenerator
	Audit audit.Logger
}

// New 构造一个 Store 实例。
func New(cfg Config) (Store, error) {
	if cfg.DB == nil {
		return nil, errors.New("wallet: Config.DB is nil")
	}
	if cfg.IDGen == nil {
		return nil, errors.New("wallet: Config.IDGen is nil")
	}
	if cfg.Log == nil {
		cfg.Log = zap.NewNop()
	}
	if cfg.Clock == nil {
		cfg.Clock = clock.Real
	}
	if cfg.Audit == nil {
		cfg.Audit = audit.NewNoop()
	}
	return &store{
		db:    cfg.DB,
		rdb:   cfg.Cache,
		log:   cfg.Log,
		clk:   cfg.Clock,
		idgen: cfg.IDGen,
		audit: cfg.Audit,
	}, nil
}

// store 是 Store 的默认实现。
type store struct {
	db    *gorm.DB
	rdb   *redis.Client
	log   *zap.Logger
	clk   clock.Clock
	idgen IDGenerator
	audit audit.Logger
}

// walletRedisKey 返回 Redis hash 的 key 名。
func walletRedisKey(ownerType string, ownerID int64) string {
	return fmt.Sprintf("wallet:%s:%d", ownerType, ownerID)
}

// Close 占位。当前 wallet 包无后台 goroutine。
func (s *store) Close() error { return nil }

// wrapDB 把 DB 错误统一包成 apierr。
func wrapDB(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apierr.New(apierr.CodeNotFound, err.Error())
	}
	return apierr.New(apierr.CodeDatabase, err.Error())
}
