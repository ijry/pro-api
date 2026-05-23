package billing

import (
	"context"
	"time"

	"github.com/ijry/pro-api/internal/wallet"
	"go.uber.org/zap"
)

func (b *biller) runLedgerWorker() {
	defer close(b.ledgerWorkerDone)
	flush := time.NewTicker(time.Duration(b.cfg.LedgerFlushEvery) * time.Millisecond)
	defer flush.Stop()

	buf := make([]wallet.LedgerEvent, 0, b.cfg.LedgerBatchSize)

	doFlush := func() {
		if len(buf) == 0 {
			return
		}
		for attempt := 0; attempt <= b.cfg.LedgerRetryMax; attempt++ {
			if err := b.cfg.Wallet.ApplyLedgerBatch(context.Background(), buf); err == nil {
				break
			} else if attempt == b.cfg.LedgerRetryMax {
				b.cfg.Log.Error("billing: ledger flush failed after retries", zap.Error(err))
			}
		}
		buf = buf[:0]
	}

	for {
		select {
		case e, ok := <-b.ledgerCh:
			if !ok {
				doFlush()
				return
			}
			buf = append(buf, e)
			if len(buf) >= b.cfg.LedgerBatchSize {
				doFlush()
			}
		case <-flush.C:
			doFlush()
		}
	}
}

func (b *biller) runReconciler() {
	defer close(b.reconcileDone)
	ticker := time.NewTicker(time.Duration(b.cfg.ReconcileEvery) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.reconcileStop:
			return
		case <-ticker.C:
			b.reconcileExpired()
		}
	}
}

func (b *biller) reconcileExpired() {
	ctx := context.Background()
	var recs []Reservation
	b.cfg.DB.WithContext(ctx).
		Where("status = 0 AND expires_at < ?", b.cfg.Clock.Now()).
		Limit(b.cfg.ReconcileBatch).
		Find(&recs)

	for _, rec := range recs {
		res, err := b.lua.Refund.RunInts(ctx, []string{walletKey(rec.UserID), reserveKey(rec.ID)})
		if err != nil || (len(res) > 0 && res[0] == -1) {
			// 已被 commit/refund 或 lua 出错,只更新 DB 状态
		}
		b.cfg.DB.WithContext(ctx).Model(&Reservation{}).
			Where("id = ? AND status = 0", rec.ID).
			Updates(map[string]any{"status": 3})
	}
}
