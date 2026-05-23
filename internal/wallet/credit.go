package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Credit 给钱包加额度。
//
// 执行流程:
//  1. 校验 amount > 0、refType 合法
//  2. 单事务:
//     - SELECT wallet FOR UPDATE
//     - UPDATE wallet SET balance += amount, total_recharged += amount(usage/refund 之外), version += 1
//     - INSERT ledger_entries(direction=credit, ref_type=refType, balance_after=new)
//  3. Redis HINCRBY balance += amount(失败仅日志)
//  4. audit.Log(action=wallet.credit)
func (s *store) Credit(
	ctx context.Context, walletID int64, amount int64,
	refType string, refID *int64, desc string, actor int64,
) error {
	if amount <= 0 {
		return apierr.New(apierr.CodeInvalidParam, "amount must be > 0")
	}
	if !isCreditRefType(refType) {
		return apierr.New(apierr.CodeInvalidParam, "invalid ref_type: "+refType)
	}
	now := s.clk.Now().UTC()

	var (
		ownerType string
		ownerID   int64
		newBal    int64
	)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var w Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", walletID).First(&w).Error; err != nil {
			return wrapDB(err)
		}
		newBal = w.QuotaBalance + amount
		updates := map[string]any{
			"quota_balance": newBal,
			"version":       gorm.Expr("version + 1"),
			"updated_at":    now,
		}
		if isRechargeRefType(refType) {
			updates["quota_total_recharged"] = gorm.Expr("quota_total_recharged + ?", amount)
		}
		if err := tx.Model(&Wallet{}).Where("id = ?", walletID).Updates(updates).Error; err != nil {
			return apierr.New(apierr.CodeDatabase, err.Error())
		}
		row := LedgerEntry{
			ID:           s.idgen.Generate(),
			WalletID:     walletID,
			Direction:    DirectionCredit,
			AmountQuota:  amount,
			AmountMoney:  0,
			Currency:     w.Currency,
			RefType:      refType,
			RefID:        refID,
			BalanceAfter: newBal,
			Description:  desc,
			CreatedAt:    now,
		}
		if err := tx.Create(&row).Error; err != nil {
			return apierr.New(apierr.CodeDatabase, err.Error())
		}
		ownerType = w.OwnerType
		ownerID = w.OwnerID
		return nil
	})
	if err != nil {
		return err
	}

	// Redis 同步(尽力而为)
	if s.rdb != nil {
		key := walletRedisKey(ownerType, ownerID)
		if _, herr := s.rdb.HIncrBy(ctx, key, hashFieldBalance, amount).Result(); herr != nil {
			s.log.Warn("wallet: HIncrBy balance failed",
				zap.String("key", key), zap.Int64("amount", amount), zap.Error(herr))
		}
	}

	// audit
	if s.audit != nil {
		var actorPtr *int64
		if actor != 0 {
			a := actor
			actorPtr = &a
		}
		wid := walletID
		afterJSON, _ := json.Marshal(map[string]any{
			"wallet_id":     walletID,
			"amount":        amount,
			"ref_type":      refType,
			"description":   desc,
			"balance_after": newBal,
		})
		_ = s.audit.Log(ctx, audit.Entry{
			ActorID:    actorPtr,
			Action:     "wallet.credit",
			TargetType: "wallet",
			TargetID:   &wid,
			After:      afterJSON,
		})
	}
	return nil
}

// isCreditRefType 判断 ref_type 是否允许在 Credit 调用。
func isCreditRefType(t string) bool {
	switch t {
	case RefTypeManual, RefTypeRedeem, RefTypeRefund, RefTypeAdjust:
		return true
	default:
		return false
	}
}

// isRechargeRefType 判断 ref_type 是否计入 quota_total_recharged。
// refund 是退款回滚,不算"充值";adjust 等也不算。
func isRechargeRefType(t string) bool {
	return t == RefTypeManual || t == RefTypeRedeem
}

// ApplyLedgerBatch 单事务批量写 ledger + 同步 wallet 累计字段。
//
// 调用方(billing 的异步 ledger worker)需保证 events 同 wallet_id。
//
// 事务内顺序:
//  1. SELECT wallet FOR UPDATE
//  2. 算出每条 ledger 的 balance_after(顺序相加 / 相减)
//  3. CreateInBatches(ledger_rows)
//  4. UPDATE wallets SET balance = final, total_consumed += sum_debit, version += 1
//
// 注意:credit 路径(包括 refund)不增 total_recharged(由 Credit 显式维护)。
func (s *store) ApplyLedgerBatch(ctx context.Context, events []LedgerEvent) error {
	if len(events) == 0 {
		return nil
	}
	walletID := events[0].WalletID
	for _, e := range events {
		if e.WalletID != walletID {
			return apierr.New(apierr.CodeInvalidParam,
				fmt.Sprintf("ApplyLedgerBatch: mixed wallet_id (%d vs %d)", walletID, e.WalletID))
		}
		if e.Direction != DirectionDebit && e.Direction != DirectionCredit {
			return apierr.New(apierr.CodeInvalidParam,
				"ApplyLedgerBatch: invalid direction: "+e.Direction)
		}
	}
	now := s.clk.Now().UTC()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var w Wallet
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", walletID).First(&w).Error
		if err != nil {
			return wrapDB(err)
		}
		cur := w.QuotaBalance
		var sumDebit, sumCredit int64
		rows := make([]LedgerEntry, 0, len(events))
		for _, ev := range events {
			if ev.Direction == DirectionDebit {
				cur -= ev.AmountQuota
				sumDebit += ev.AmountQuota
			} else {
				cur += ev.AmountQuota
				sumCredit += ev.AmountQuota
			}
			ts := ev.OccurredAt
			if ts.IsZero() {
				ts = now
			}
			rows = append(rows, LedgerEntry{
				ID:           s.idgen.Generate(),
				WalletID:     ev.WalletID,
				Direction:    ev.Direction,
				AmountQuota:  ev.AmountQuota,
				Currency:     defaultCurrency(w.Currency),
				RefType:      ev.RefType,
				RefID:        ev.RefID,
				BalanceAfter: cur,
				Description:  truncate(ev.Description, 256),
				CreatedAt:    ts,
			})
		}
		if err := tx.CreateInBatches(&rows, 100).Error; err != nil {
			return apierr.New(apierr.CodeDatabase, err.Error())
		}
		updates := map[string]any{
			"quota_balance":        cur,
			"quota_total_consumed": gorm.Expr("quota_total_consumed + ?", sumDebit),
			"version":              gorm.Expr("version + 1"),
			"updated_at":           now,
		}
		_ = sumCredit // 不计入 total_recharged(spec §4.6 决策)
		if err := tx.Model(&Wallet{}).Where("id = ?", walletID).Updates(updates).Error; err != nil {
			return apierr.New(apierr.CodeDatabase, err.Error())
		}
		return nil
	})
}

func defaultCurrency(c string) string {
	if c == "" {
		return "USD"
	}
	return c
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.ToValidUTF8(s[:n], "")
}
