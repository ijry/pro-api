// Package wallet 实现用户钱包的余额、流水与额度同步。
//
// 设计稿:M1-06 spec §3.1。
//
// 关键点:
//   - DB 真源 + Redis hash 缓存(永久,无 TTL)
//   - Credit 是管理员加额度 / 兑换码 / 手动充值的统一入口(事务 + Redis 同步 + audit)
//   - ApplyLedgerBatch 给 billing 的异步 ledger worker 调,单事务保证 ledger 与累计字段一致
package wallet

import "time"

// 所有者类型。
const (
	OwnerTypeUser = "user"
	OwnerTypeDept = "dept" // M3,M1 不允许使用
)

// Wallet 是 wallets 表的 GORM 模型。
type Wallet struct {
	ID                  int64     `gorm:"primaryKey;column:id"`
	OwnerType           string    `gorm:"column:owner_type;size:16"`
	OwnerID             int64     `gorm:"column:owner_id"`
	QuotaBalance        int64     `gorm:"column:quota_balance"`
	QuotaTotalRecharged int64     `gorm:"column:quota_total_recharged"`
	QuotaTotalConsumed  int64     `gorm:"column:quota_total_consumed"`
	Currency            string    `gorm:"column:currency;size:8"`
	Version             int       `gorm:"column:version"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

// TableName 固定表名。
func (Wallet) TableName() string { return "wallets" }

// LedgerEntry 是 ledger_entries 表的 GORM 模型。
type LedgerEntry struct {
	ID           int64     `gorm:"primaryKey;column:id"`
	WalletID     int64     `gorm:"column:wallet_id"`
	Direction    string    `gorm:"column:direction;size:8"`
	AmountQuota  int64     `gorm:"column:amount_quota"`
	AmountMoney  int64     `gorm:"column:amount_money"`
	Currency     string    `gorm:"column:currency;size:8"`
	RefType      string    `gorm:"column:ref_type;size:16"`
	RefID        *int64    `gorm:"column:ref_id"`
	BalanceAfter int64     `gorm:"column:balance_after"`
	Description  string    `gorm:"column:description;size:256"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

// TableName 固定表名。
func (LedgerEntry) TableName() string { return "ledger_entries" }

// direction 取值。
const (
	DirectionDebit  = "debit"
	DirectionCredit = "credit"
)

// ref_type 取值。
const (
	RefTypeUsage  = "usage"
	RefTypeManual = "manual"
	RefTypeRedeem = "redeem"
	RefTypeRefund = "refund"
	RefTypeAdjust = "adjust"
)

// LedgerEvent 是异步 ledger 入队事件。
//
// billing 模块通过 WalletWriter.ApplyLedgerBatch 提交一批,wallet 在单事务内写
// ledger_entries + 同步 wallets 累计字段。
//
// 决策:LedgerEvent 是叶子数据类型(仅基本字段、无业务方法),所以 billing 可以单向
// import wallet(总纲 §3.2)。wallet 不反向 import billing。
type LedgerEvent struct {
	WalletID    int64
	UserID      int64
	Direction   string // DirectionDebit | DirectionCredit
	AmountQuota int64
	RefType     string
	RefID       *int64
	Description string
	OccurredAt  time.Time
}
