package billing

import (
	"context"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"github.com/ijry/pro-api/internal/wallet"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WalletWriter 是 billing 对 wallet 的最小依赖。
type WalletWriter interface {
	GetOrCreate(ctx context.Context, ownerType string, ownerID int64) (*wallet.Wallet, error)
	ApplyLedgerBatch(ctx context.Context, events []wallet.LedgerEvent) error
}

// UsageIncrementer 是 billing 对 token quota_used 的最小依赖。
type UsageIncrementer interface {
	IncrementUsage(tokenID, delta int64)
}

// Config 是 New 的参数。
type Config struct {
	DB      *gorm.DB
	Cache   *redis.Client
	Log     *zap.Logger
	Clock   clock.Clock
	IDGen   *idgen.Generator
	Setting setting.Store
	Audit   audit.Logger
	Wallet  WalletWriter
	Usage   UsageIncrementer

	LedgerQueueCap    int
	LedgerBatchSize   int
	LedgerRetryMax    int
	LedgerFlushEvery  int // ms
	ReconcileEvery    int // seconds
	ReconcileBatch    int
	ReserveTTLSeconds int

	DisableLedgerWorker bool
	DisableReconciler   bool
}
