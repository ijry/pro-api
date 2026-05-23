package manual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IDGenerator 是对 idgen 的最小依赖。
type IDGenerator interface {
	Generate() int64
}

// Clock 是对 clock.Clock 的最小依赖。
type Clock interface {
	Now() time.Time
}

// SettingStore 是对 setting.Store 的最小依赖(仅 typed getters)。
type SettingStore interface {
	GetBool(ctx context.Context, key string, def bool) bool
	GetInt(ctx context.Context, key string, def int) int
	GetFloat(ctx context.Context, key string, def float64) float64
	GetString(ctx context.Context, key string, def string) string
}

// WalletCredit 是 M1-06 wallet.Store.Credit 的最小依赖。
//
// 解耦说明:本包不直接 import M1-06 wallet 包;wire 时由 main.go
// 注入符合此签名的 adapter。amount 严格为正(非零正整数)。
type WalletCredit interface {
	Credit(ctx context.Context, userID int64, amount int64, refType string, refID int64, desc string) error
}

// Service 是 manual recharge 业务接口。
type Service interface {
	// Apply 由用户调用。amountMoney 单位"厘";note ≤ 512 字符。
	Apply(ctx context.Context, userID int64, amountMoney int64, note string) (*Recharge, error)

	// Approve 由管理员调用。状态 pending → approved + wallet.Credit。
	Approve(ctx context.Context, id, reviewerID int64, note string) (*Recharge, error)

	// Reject 由管理员调用。状态 pending → rejected,不调钱包。
	Reject(ctx context.Context, id, reviewerID int64, note string) (*Recharge, error)

	// Cancel 由用户本人调用。状态 pending → canceled。
	Cancel(ctx context.Context, id, userID int64) (*Recharge, error)

	// List 列出。filter.UserID=0 表示 admin(不限);=具体值 = 用户自己。
	List(ctx context.Context, f ListFilter) ([]*Recharge, int64, error)

	// Get 单条;userID=0 表示 admin(不做 owner 校验)。
	Get(ctx context.Context, id int64, userID int64) (*Recharge, error)
}

// Config 是 New 的配置。
type Config struct {
	DB      *gorm.DB
	Setting SettingStore
	Wallet  WalletCredit
	Audit   audit.Logger
	IDGen   IDGenerator
	Clock   Clock
	Log     *zap.Logger
}

type service struct {
	repo    Repo
	db      *gorm.DB
	setting SettingStore
	wallet  WalletCredit
	audit   audit.Logger
	idgen   IDGenerator
	clock   Clock
	log     *zap.Logger
}

// New 构造默认 Service。
func New(cfg Config) Service {
	log := cfg.Log
	if log == nil {
		log = zap.NewNop()
	}
	aud := cfg.Audit
	if aud == nil {
		aud = audit.NewNoop()
	}
	return &service{
		repo:    NewRepository(cfg.DB),
		db:      cfg.DB,
		setting: cfg.Setting,
		wallet:  cfg.Wallet,
		audit:   aud,
		idgen:   cfg.IDGen,
		clock:   cfg.Clock,
		log:     log,
	}
}

// ---------- Apply ----------

func (s *service) Apply(ctx context.Context, userID int64, amountMoney int64, note string) (*Recharge, error) {
	if s.setting != nil && !s.setting.GetBool(ctx, "manual_recharge.enabled", true) {
		return nil, apierr.New(apierr.CodeInvalidParam, "manual recharge feature disabled").
			WithDetails(map[string]any{"reason": "manual_recharge_disabled"})
	}
	if amountMoney <= 0 {
		return nil, apierr.New(apierr.CodeInvalidParam, "amount must be > 0").
			WithDetails(map[string]any{"reason": "amount_yuan_out_of_range"})
	}
	minY := 1
	maxY := 1_000_000
	if s.setting != nil {
		minY = s.setting.GetInt(ctx, "manual_recharge.min_amount_cny", 100)
		maxY = s.setting.GetInt(ctx, "manual_recharge.max_amount_cny", 100000)
	}
	// amountMoney 是厘,minY/maxY 是元
	amountYuan := amountMoney / 10000
	if amountYuan < int64(minY) || amountYuan > int64(maxY) {
		return nil, apierr.New(apierr.CodeInvalidParam, fmt.Sprintf("amount must be in [%d, %d] CNY", minY, maxY)).
			WithDetails(map[string]any{
				"reason": "amount_yuan_out_of_range",
				"min":    minY,
				"max":    maxY,
			})
	}
	if len(note) > 512 {
		return nil, apierr.New(apierr.CodeInvalidParam, "note too long").
			WithDetails(map[string]any{"reason": "note_too_long"})
	}

	now := s.clock.Now()
	rec := &Recharge{
		ID:            s.idgen.Generate(),
		UserID:        userID,
		AmountMoney:   amountMoney,
		Currency:      CurrencyCNY,
		Status:        StatusPending,
		ApplicantNote: note,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "create recharge failed", err)
	}
	s.auditLog(ctx, "recharge.apply", rec.ID, userID, nil, rec)
	return rec, nil
}

// ---------- Approve ----------

func (s *service) Approve(ctx context.Context, id, reviewerID int64, note string) (*Recharge, error) {
	if len(note) > 512 {
		return nil, apierr.New(apierr.CodeInvalidParam, "review_note too long")
	}

	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get recharge failed", err)
	}
	if cur == nil {
		return nil, apierr.New(apierr.CodeOrderNotFound, "充值单不存在")
	}
	if cur.Status != StatusPending {
		return nil, apierr.New(apierr.CodeInvalidParam, "充值单已被处理").
			WithDetails(map[string]any{"reason": "not_pending", "status": StatusName(cur.Status)})
	}

	// 计算 quota:approval-time rate
	rate := 7.0
	base := 500_000
	if s.setting != nil {
		rate = s.setting.GetFloat(ctx, "manual_recharge.exchange_rate_cny_per_usd", 7.0)
		base = s.setting.GetInt(ctx, "pricing.base_quota_per_dollar", 500_000)
	}
	quota := ComputeQuota(cur.AmountMoney, rate, int64(base))
	if quota <= 0 {
		return nil, apierr.New(apierr.CodeInternal, "exchange rate or base quota misconfigured").
			WithDetails(map[string]any{"rate": rate, "base": base})
	}

	now := s.clock.Now()
	rid := reviewerID
	fields := map[string]any{
		"status":       StatusApproved,
		"amount_quota": quota,
		"reviewer_id":  rid,
		"review_note":  note,
		"reviewed_at":  now,
		"updated_at":   now,
	}
	affected, err := s.repo.UpdateStatusFromPending(ctx, id, fields)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "approve recharge failed", err)
	}
	if affected == 0 {
		// 并发:别处已经改了状态
		return nil, apierr.New(apierr.CodeInvalidParam, "充值单已被处理").
			WithDetails(map[string]any{"reason": "not_pending"})
	}

	// 调 wallet.Credit
	desc := fmt.Sprintf("manual recharge approved by admin %d", reviewerID)
	if err := s.wallet.Credit(ctx, cur.UserID, quota, "manual", id, desc); err != nil {
		// 补偿:回滚状态
		s.rollbackApprove(ctx, id)
		return nil, apierr.Wrap(apierr.CodeInternal, "wallet credit failed", err)
	}

	after, err := s.repo.GetByID(ctx, id)
	if err != nil || after == nil {
		// DB 拉不到 — 极少见;返回内存版,日志记 warn
		s.log.Warn("approve: re-read failed",
			zap.Int64("id", id), zap.Error(err))
		after = &Recharge{
			ID:            cur.ID,
			UserID:        cur.UserID,
			AmountMoney:   cur.AmountMoney,
			Currency:      cur.Currency,
			AmountQuota:   quota,
			Status:        StatusApproved,
			ApplicantNote: cur.ApplicantNote,
			ReviewerID:    &rid,
			ReviewNote:    note,
			ReviewedAt:    &now,
			CreatedAt:     cur.CreatedAt,
			UpdatedAt:     now,
		}
	}

	s.auditLog(ctx, "recharge.approve", id, reviewerID, cur, after)
	return after, nil
}

// rollbackApprove 在 Credit 失败后把状态从 approved 回滚到 pending。
func (s *service) rollbackApprove(ctx context.Context, id int64) {
	err := s.db.WithContext(ctx).Model(&Recharge{}).
		Where("id = ? AND status = ?", id, StatusApproved).
		Updates(map[string]any{
			"status":       StatusPending,
			"amount_quota": 0,
			"reviewer_id":  nil,
			"review_note":  "",
			"reviewed_at":  nil,
			"updated_at":   s.clock.Now(),
		}).Error
	if err != nil {
		s.log.Error("manual: approve rollback failed",
			zap.Int64("id", id), zap.Error(err))
		_ = s.audit.Log(ctx, audit.Entry{
			Action:     "recharge.rollback_failed",
			TargetType: "manual_recharge",
			TargetID:   &id,
		})
		return
	}
	_ = s.audit.Log(ctx, audit.Entry{
		Action:     "recharge.rollback",
		TargetType: "manual_recharge",
		TargetID:   &id,
	})
}

// ---------- Reject ----------

func (s *service) Reject(ctx context.Context, id, reviewerID int64, note string) (*Recharge, error) {
	if len(note) > 512 {
		return nil, apierr.New(apierr.CodeInvalidParam, "review_note too long")
	}
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get recharge failed", err)
	}
	if cur == nil {
		return nil, apierr.New(apierr.CodeOrderNotFound, "充值单不存在")
	}
	if cur.Status != StatusPending {
		return nil, apierr.New(apierr.CodeInvalidParam, "充值单已被处理").
			WithDetails(map[string]any{"reason": "not_pending"})
	}

	now := s.clock.Now()
	rid := reviewerID
	fields := map[string]any{
		"status":      StatusRejected,
		"reviewer_id": rid,
		"review_note": note,
		"reviewed_at": now,
		"updated_at":  now,
	}
	affected, err := s.repo.UpdateStatusFromPending(ctx, id, fields)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "reject recharge failed", err)
	}
	if affected == 0 {
		return nil, apierr.New(apierr.CodeInvalidParam, "充值单已被处理")
	}
	after, _ := s.repo.GetByID(ctx, id)
	s.auditLog(ctx, "recharge.reject", id, reviewerID, cur, after)
	return after, nil
}

// ---------- Cancel ----------

func (s *service) Cancel(ctx context.Context, id, userID int64) (*Recharge, error) {
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get recharge failed", err)
	}
	if cur == nil || cur.UserID != userID {
		// 不暴露存在性,他人单一律 not found
		return nil, apierr.New(apierr.CodeOrderNotFound, "充值单不存在")
	}
	if cur.Status != StatusPending {
		return nil, apierr.New(apierr.CodeInvalidParam, "充值单已被处理").
			WithDetails(map[string]any{"reason": "not_pending"})
	}

	now := s.clock.Now()
	fields := map[string]any{
		"status":     StatusCanceled,
		"updated_at": now,
	}
	affected, err := s.repo.UpdateStatusFromPending(ctx, id, fields)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "cancel recharge failed", err)
	}
	if affected == 0 {
		return nil, apierr.New(apierr.CodeInvalidParam, "充值单已被处理").
			WithDetails(map[string]any{"reason": "not_pending"})
	}
	after, _ := s.repo.GetByID(ctx, id)
	s.auditLog(ctx, "recharge.cancel", id, userID, cur, after)
	return after, nil
}

// ---------- List / Get ----------

func (s *service) List(ctx context.Context, f ListFilter) ([]*Recharge, int64, error) {
	items, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, apierr.Wrap(apierr.CodeDatabase, "list recharge failed", err)
	}
	return items, total, nil
}

func (s *service) Get(ctx context.Context, id int64, userID int64) (*Recharge, error) {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get recharge failed", err)
	}
	if rec == nil {
		return nil, apierr.New(apierr.CodeOrderNotFound, "充值单不存在")
	}
	// userID != 0 表示用户视角,需校验 owner;userID = 0 表示 admin,绕过
	if userID != 0 && rec.UserID != userID {
		return nil, apierr.New(apierr.CodeOrderNotFound, "充值单不存在")
	}
	return rec, nil
}

// ---------- audit helper ----------

func (s *service) auditLog(ctx context.Context, action string, id int64, actor int64, before, after any) {
	var beforeJSON, afterJSON json.RawMessage
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			beforeJSON = b
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			afterJSON = b
		}
	}
	actorRef := actor
	idRef := id
	_ = s.audit.Log(ctx, audit.Entry{
		Action:     action,
		TargetType: "manual_recharge",
		TargetID:   &idRef,
		ActorID:    &actorRef,
		Before:     beforeJSON,
		After:      afterJSON,
	})
}

// 兜底:errors 可能未在某些路径用到
var _ = errors.Is
