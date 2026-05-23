package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ijry/pro-api/internal/billing/lua"
	"github.com/ijry/pro-api/internal/wallet"
	"go.uber.org/zap"
)

// Biller 是三阶段计费接口。
type Biller interface {
	// Reserve 预扣 quota;余额不足返回 ErrInsufficient。
	Reserve(ctx context.Context, userID, tokenID, estCost int64) (reservationID string, err error)
	// Commit 提交实际成本,自动退还差额。
	Commit(ctx context.Context, reservationID string, actualCost int64) error
	// Refund 完整退还(上游 401/网络错误等)。
	Refund(ctx context.Context, reservationID string) error
	// Close 优雅关停(drain ledger worker)。
	Close() error
}

// ErrInsufficient 余额不足。
var ErrInsufficient = fmt.Errorf("billing: quota insufficient")

type biller struct {
	cfg Config
	lua *lua.Scripts

	ledgerCh        chan wallet.LedgerEvent
	ledgerWorkerDone chan struct{}
	reconcileStop   chan struct{}
	reconcileDone   chan struct{}
}

// New 构造 Biller;内部启动 ledger worker + reconcile cron。
func New(cfg Config) (Biller, error) {
	if cfg.LedgerQueueCap <= 0 { cfg.LedgerQueueCap = 1024 }
	if cfg.LedgerBatchSize <= 0 { cfg.LedgerBatchSize = 100 }
	if cfg.LedgerFlushEvery <= 0 { cfg.LedgerFlushEvery = 1000 }
	if cfg.LedgerRetryMax <= 0 { cfg.LedgerRetryMax = 3 }
	if cfg.ReconcileEvery <= 0 { cfg.ReconcileEvery = 30 }
	if cfg.ReconcileBatch <= 0 { cfg.ReconcileBatch = 100 }
	if cfg.ReserveTTLSeconds <= 0 { cfg.ReserveTTLSeconds = 600 }

	scripts, err := lua.Load(cfg.Cache)
	if err != nil {
		return nil, fmt.Errorf("billing: load lua: %w", err)
	}

	b := &biller{
		cfg:              cfg,
		lua:              scripts,
		ledgerCh:         make(chan wallet.LedgerEvent, cfg.LedgerQueueCap),
		ledgerWorkerDone: make(chan struct{}),
		reconcileStop:    make(chan struct{}),
		reconcileDone:    make(chan struct{}),
	}

	if !cfg.DisableLedgerWorker {
		go b.runLedgerWorker()
	} else {
		close(b.ledgerWorkerDone)
	}
	if !cfg.DisableReconciler {
		go b.runReconciler()
	} else {
		close(b.reconcileDone)
	}

	cfg.Log.Info("billing: started",
		zap.String("reserve_sha", scripts.ReserveSHA()),
		zap.String("commit_sha", scripts.CommitSHA()),
		zap.String("refund_sha", scripts.RefundSHA()),
	)
	return b, nil
}

func walletKey(userID int64) string { return fmt.Sprintf("wallet:user:%d", userID) }
func reserveKey(rid string) string  { return fmt.Sprintf("reservation:%s", rid) }

func (b *biller) Reserve(ctx context.Context, userID, tokenID, estCost int64) (string, error) {
	w, err := b.cfg.Wallet.GetOrCreate(ctx, "user", userID)
	if err != nil {
		return "", fmt.Errorf("billing: get wallet: %w", err)
	}

	rid := uuid.New().String()
	ttl := b.cfg.ReserveTTLSeconds

	res, err := b.lua.Reserve.RunInts(ctx, []string{walletKey(userID), reserveKey(rid)}, estCost, ttl)
	if err != nil {
		return "", fmt.Errorf("billing: reserve lua: %w", err)
	}
	if res[0] == 0 {
		return "", ErrInsufficient
	}

	// 写 DB
	expiresAt := b.cfg.Clock.Now().Add(time.Duration(ttl) * time.Second)
	rec := &Reservation{
		ID:            rid,
		WalletID:      w.ID,
		UserID:        userID,
		TokenID:       tokenID,
		ReservedQuota: estCost,
		Status:        0,
		CreatedAt:     b.cfg.Clock.Now(),
		ExpiresAt:     expiresAt,
	}
	if dbErr := b.cfg.DB.WithContext(ctx).Create(rec).Error; dbErr != nil {
		// DB 失败: 回滚 Redis
		b.cfg.Log.Warn("billing: reserve db create failed, refunding redis",
			zap.Error(dbErr), zap.String("rid", rid))
		_, _ = b.lua.Refund.RunInts(ctx, []string{walletKey(userID), reserveKey(rid)})
		return "", fmt.Errorf("billing: reserve db: %w", dbErr)
	}
	return rid, nil
}

func (b *biller) Commit(ctx context.Context, rid string, actualCost int64) error {
	var rec Reservation
	if err := b.cfg.DB.WithContext(ctx).Where("id = ?", rid).First(&rec).Error; err != nil {
		return fmt.Errorf("billing: commit find: %w", err)
	}

	res, err := b.lua.Commit.RunInts(ctx, []string{walletKey(rec.UserID), reserveKey(rid)}, actualCost)
	if err != nil {
		return fmt.Errorf("billing: commit lua: %w", err)
	}
	if res[0] != 1 {
		return fmt.Errorf("billing: commit lua returned %d", res[0])
	}

	now := b.cfg.Clock.Now()
	if dbErr := b.cfg.DB.WithContext(ctx).Model(&Reservation{}).
		Where("id = ? AND status = 0", rid).
		Updates(map[string]any{
			"status":          1,
			"committed_quota": actualCost,
			"committed_at":    now,
		}).Error; dbErr != nil {
		b.cfg.Log.Error("billing: commit db update failed", zap.Error(dbErr), zap.String("rid", rid))
	}

	// 异步写 ledger
	b.pushLedger(wallet.LedgerEvent{
		WalletID:    rec.WalletID,
		UserID:      rec.UserID,
		Direction:   "debit",
		AmountQuota: actualCost,
		RefType:     "usage",
		OccurredAt:  now,
	})

	if b.cfg.Usage != nil {
		b.cfg.Usage.IncrementUsage(rec.TokenID, actualCost)
	}
	return nil
}

func (b *biller) Refund(ctx context.Context, rid string) error {
	var rec Reservation
	if err := b.cfg.DB.WithContext(ctx).Where("id = ?", rid).First(&rec).Error; err != nil {
		return fmt.Errorf("billing: refund find: %w", err)
	}

	_, err := b.lua.Refund.RunInts(ctx, []string{walletKey(rec.UserID), reserveKey(rid)})
	if err != nil {
		return fmt.Errorf("billing: refund lua: %w", err)
	}

	if dbErr := b.cfg.DB.WithContext(ctx).Model(&Reservation{}).
		Where("id = ? AND status = 0", rid).
		Updates(map[string]any{"status": 2}).Error; dbErr != nil {
		b.cfg.Log.Warn("billing: refund db update failed", zap.Error(dbErr), zap.String("rid", rid))
	}
	return nil
}

func (b *biller) pushLedger(e wallet.LedgerEvent) {
	select {
	case b.ledgerCh <- e:
	default:
		b.cfg.Log.Warn("billing: ledger channel full, dropping event")
	}
}

func (b *biller) Close() error {
	close(b.reconcileStop)
	<-b.reconcileDone
	close(b.ledgerCh)
	<-b.ledgerWorkerDone
	return nil
}
