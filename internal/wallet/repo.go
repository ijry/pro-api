package wallet

import (
	"context"
	"errors"

	"github.com/ijry/pro-api/pkg/apierr"
	"gorm.io/gorm"
)

// GetOrCreate 取或创建 wallet。
//
// 并发安全:
//   - 先查;命中即返回
//   - 不存在则插入,失败若是唯一约束冲突,再查一次返回(被别人插了)
func (s *store) GetOrCreate(ctx context.Context, ownerType string, ownerID int64) (*Wallet, error) {
	if ownerType == "" {
		return nil, apierr.New(apierr.CodeInvalidParam, "owner_type required")
	}
	if ownerID <= 0 {
		return nil, apierr.New(apierr.CodeInvalidParam, "owner_id must be > 0")
	}
	var w Wallet
	err := s.db.WithContext(ctx).
		Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
		First(&w).Error
	if err == nil {
		s.ensureRedisHash(ctx, &w)
		return &w, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apierr.New(apierr.CodeDatabase, err.Error())
	}
	now := s.clk.Now().UTC()
	nw := Wallet{
		ID:        s.idgen.Generate(),
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Currency:  "USD",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.db.WithContext(ctx).Create(&nw).Error; err != nil {
		// 并发插入冲突 → 再查一次
		var existing Wallet
		e2 := s.db.WithContext(ctx).
			Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
			First(&existing).Error
		if e2 == nil {
			s.ensureRedisHash(ctx, &existing)
			return &existing, nil
		}
		return nil, apierr.New(apierr.CodeDatabase, err.Error())
	}
	s.ensureRedisHash(ctx, &nw)
	return &nw, nil
}

// Snapshot 从 DB 全量读取(不读 Redis)。
func (s *store) Snapshot(ctx context.Context, walletID int64) (*Wallet, error) {
	if walletID <= 0 {
		return nil, apierr.New(apierr.CodeInvalidParam, "wallet_id must be > 0")
	}
	var w Wallet
	if err := s.db.WithContext(ctx).Where("id = ?", walletID).First(&w).Error; err != nil {
		return nil, wrapDB(err)
	}
	return &w, nil
}

// ByOwner 按 owner 取(不存在返回 CodeNotFound)。
func (s *store) ByOwner(ctx context.Context, ownerType string, ownerID int64) (*Wallet, error) {
	if ownerType == "" || ownerID <= 0 {
		return nil, apierr.New(apierr.CodeInvalidParam, "owner_type/owner_id invalid")
	}
	var w Wallet
	if err := s.db.WithContext(ctx).
		Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
		First(&w).Error; err != nil {
		return nil, wrapDB(err)
	}
	return &w, nil
}

// ListLedger 流水分页查询。
func (s *store) ListLedger(ctx context.Context, filter LedgerFilter) ([]*LedgerEntry, int64, error) {
	if filter.WalletID <= 0 {
		return nil, 0, apierr.New(apierr.CodeInvalidParam, "wallet_id required")
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Size <= 0 {
		filter.Size = 20
	}
	if filter.Size > 100 {
		filter.Size = 100
	}
	q := s.db.WithContext(ctx).Model(&LedgerEntry{}).Where("wallet_id = ?", filter.WalletID)
	if filter.RefType != "" {
		q = q.Where("ref_type = ?", filter.RefType)
	}
	if filter.Since != nil {
		q = q.Where("created_at >= ?", filter.Since.UTC())
	}
	if filter.Until != nil {
		q = q.Where("created_at < ?", filter.Until.UTC())
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apierr.New(apierr.CodeDatabase, err.Error())
	}
	var rows []*LedgerEntry
	if err := q.Order("created_at DESC, id DESC").
		Offset((filter.Page - 1) * filter.Size).
		Limit(filter.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, apierr.New(apierr.CodeDatabase, err.Error())
	}
	return rows, total, nil
}
