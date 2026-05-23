package redeem

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
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

// SettingStore 是对 setting.Store 的最小依赖。
type SettingStore interface {
	GetInt(ctx context.Context, key string, def int) int
}

// WalletCredit 是 M1-06 wallet.Store.Credit 的最小依赖。
//
// 解耦说明:本包不直接 import M1-06 wallet 包;wire 时由 main.go
// 注入符合此签名的 adapter。amount 严格为正(非零正整数)。
type WalletCredit interface {
	Credit(ctx context.Context, userID int64, amount int64, refType string, refID int64, desc string) error
}

// Service 是 redeem 业务接口。
type Service interface {
	// BatchCreate 批量生成兑换码。
	//
	//   count: 1..1000;否则 CodeInvalidParam
	//   amountQuota: > 0
	//   expiresAt: nil → 用 setting redeem.default_expires_days 算;0 days = 永不过期
	//   batchNo: 可空,空则自动 "B" + yyyymmddHHMMSS + 4 位随机
	//
	// plaintexts 与 codeIDs 等长且顺序一致;plaintexts 仅此一次返回。
	BatchCreate(ctx context.Context, actorID int64, count int, amountQuota int64, expiresAt *time.Time, batchNo string) (plaintexts []string, batchNoUsed string, codeIDs []int64, err error)

	// Redeem 兑换。失败语义统一用 CodeRedeemInvalid + details.reason 区分:
	//   format / not_found / used / disabled / expired
	Redeem(ctx context.Context, userID int64, plaintext string) (*RedeemResult, error)

	// Disable 批量禁用(unused → disabled,其他状态跳过)。
	Disable(ctx context.Context, actorID int64, ids []int64, reason string) (disabledCount int, err error)

	List(ctx context.Context, f ListFilter) ([]*Code, int64, error)
	Get(ctx context.Context, id int64) (*Code, error)

	// Export 流式 CSV 写到 w,filter 同 List 但不分页;不包含明文。
	Export(ctx context.Context, w io.Writer, f ListFilter) error
}

// RedeemResult 是 Redeem 成功的返回。
type RedeemResult struct {
	CodeID      int64
	AmountQuota int64
	UsedAt      time.Time
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

// ---------- BatchCreate ----------

const (
	maxBatchCount       = 1000
	codegenMaxRetry     = 5
	batchInsertMaxRetry = 2
)

func (s *service) BatchCreate(
	ctx context.Context, actorID int64, count int, amountQuota int64,
	expiresAt *time.Time, batchNo string,
) ([]string, string, []int64, error) {
	if count < 1 || count > maxBatchCount {
		return nil, "", nil, apierr.New(apierr.CodeInvalidParam, "count must be 1..1000").
			WithDetails(map[string]any{"reason": "count_out_of_range", "count": count})
	}
	if amountQuota <= 0 {
		return nil, "", nil, apierr.New(apierr.CodeInvalidParam, "amount_quota must be > 0").
			WithDetails(map[string]any{"reason": "amount_quota_invalid"})
	}

	if expiresAt == nil && s.setting != nil {
		days := s.setting.GetInt(ctx, "redeem.default_expires_days", 365)
		if days > 0 {
			t := s.clock.Now().UTC().AddDate(0, 0, days)
			expiresAt = &t
		}
		// days == 0 表示永不过期,保持 nil
	}

	if batchNo == "" {
		batchNo = autoBatchNo(s.clock.Now())
	}

	plaintexts, codes, err := s.generateCodes(actorID, count, amountQuota, expiresAt, batchNo)
	if err != nil {
		return nil, "", nil, err
	}

	// 批量插入;唯一冲突重试有限次
	insertErr := s.repo.BatchInsert(ctx, codes)
	for retry := 0; retry < batchInsertMaxRetry && isUniqueConstraintErr(insertErr); retry++ {
		// 整批重生成
		plaintexts, codes, err = s.generateCodes(actorID, count, amountQuota, expiresAt, batchNo)
		if err != nil {
			return nil, "", nil, err
		}
		insertErr = s.repo.BatchInsert(ctx, codes)
	}
	if insertErr != nil {
		return nil, "", nil, apierr.Wrap(apierr.CodeDatabase, "batch insert failed", insertErr)
	}

	codeIDs := make([]int64, len(codes))
	for i, c := range codes {
		codeIDs[i] = c.ID
	}

	// 审计:不含明文
	auditAfter, _ := json.Marshal(map[string]any{
		"batch_no":     batchNo,
		"count":        count,
		"amount_quota": amountQuota,
		"expires_at":   expiresAt,
	})
	actorRef := actorID
	_ = s.audit.Log(ctx, audit.Entry{
		Action:     "redeem.batch_create",
		TargetType: "redeem_batch",
		ActorID:    &actorRef,
		After:      auditAfter,
	})
	return plaintexts, batchNo, codeIDs, nil
}

// generateCodes 生成 count 条明文 + Code 记录(不入库)。
func (s *service) generateCodes(actorID int64, count int, amountQuota int64, expiresAt *time.Time, batchNo string) ([]string, []*Code, error) {
	plaintexts := make([]string, 0, count)
	codes := make([]*Code, 0, count)
	seen := make(map[string]struct{}, count)
	now := s.clock.Now().UTC()

	for i := 0; i < count; i++ {
		var plain string
		attempt := 0
		for {
			attempt++
			if attempt > codegenMaxRetry {
				return nil, nil, apierr.New(apierr.CodeInternal, "redeem codegen retries exhausted")
			}
			p, err := generatePlaintext()
			if err != nil {
				return nil, nil, apierr.Wrap(apierr.CodeInternal, "generate plaintext failed", err)
			}
			h := hashCode(p)
			if _, dup := seen[h]; dup {
				continue
			}
			seen[h] = struct{}{}
			plain = p
			break
		}
		codes = append(codes, &Code{
			ID:          s.idgen.Generate(),
			CodeHash:    hashCode(plain),
			CodePrefix:  prefix(plain),
			AmountQuota: amountQuota,
			BatchNo:     batchNo,
			Status:      StatusUnused,
			ExpiresAt:   expiresAt,
			CreatedBy:   actorID,
			CreatedAt:   now,
		})
		plaintexts = append(plaintexts, plain)
	}
	return plaintexts, codes, nil
}

// ---------- Redeem ----------

func (s *service) Redeem(ctx context.Context, userID int64, plaintext string) (*RedeemResult, error) {
	plain, ok := normalize(plaintext)
	if !ok {
		return nil, apierr.New(apierr.CodeRedeemInvalid, "兑换码格式无效").
			WithDetails(map[string]any{"reason": "format"})
	}
	h := hashCode(plain)

	c, err := s.repo.GetByHash(ctx, h)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "find redeem code failed", err)
	}
	if c == nil {
		return nil, apierr.New(apierr.CodeRedeemInvalid, "兑换码不存在").
			WithDetails(map[string]any{"reason": "not_found"})
	}

	switch c.Status {
	case StatusUsed:
		return nil, apierr.New(apierr.CodeRedeemInvalid, "兑换码已被使用").
			WithDetails(map[string]any{"reason": "used"})
	case StatusDisabled:
		return nil, apierr.New(apierr.CodeRedeemInvalid, "兑换码已被禁用").
			WithDetails(map[string]any{"reason": "disabled"})
	}

	now := s.clock.Now().UTC()
	if c.ExpiresAt != nil && !c.ExpiresAt.IsZero() && now.After(*c.ExpiresAt) {
		return nil, apierr.New(apierr.CodeRedeemInvalid, "兑换码已过期").
			WithDetails(map[string]any{"reason": "expired"})
	}

	// UPDATE unused → used,原子;并发兜底
	affected, err := s.repo.UpdateToUsed(ctx, c.ID, userID, now)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "update redeem code failed", err)
	}
	if affected == 0 {
		// 并发已使用
		return nil, apierr.New(apierr.CodeRedeemInvalid, "兑换码已被使用").
			WithDetails(map[string]any{"reason": "used"})
	}

	// wallet.Credit;失败回滚
	desc := fmt.Sprintf("redeem code %s", c.CodePrefix)
	if err := s.wallet.Credit(ctx, userID, c.AmountQuota, "redeem", c.ID, desc); err != nil {
		s.rollbackRedeem(ctx, c.ID)
		return nil, apierr.Wrap(apierr.CodeInternal, "wallet credit failed", err)
	}

	codeIDRef := c.ID
	userIDRef := userID
	auditAfter, _ := json.Marshal(map[string]any{
		"amount_quota": c.AmountQuota,
		"used_at":      now,
	})
	_ = s.audit.Log(ctx, audit.Entry{
		Action:     "redeem.use",
		TargetType: "redeem_code",
		TargetID:   &codeIDRef,
		ActorID:    &userIDRef,
		After:      auditAfter,
	})

	return &RedeemResult{
		CodeID:      c.ID,
		AmountQuota: c.AmountQuota,
		UsedAt:      now,
	}, nil
}

// rollbackRedeem 把 status 从 used 回滚到 unused,Credit 失败时调用。
func (s *service) rollbackRedeem(ctx context.Context, codeID int64) {
	if err := s.repo.RollbackUsedToUnused(ctx, codeID); err != nil {
		s.log.Error("redeem: rollback failed",
			zap.Int64("code_id", codeID), zap.Error(err))
		idRef := codeID
		_ = s.audit.Log(ctx, audit.Entry{
			Action:     "redeem.rollback_failed",
			TargetType: "redeem_code",
			TargetID:   &idRef,
		})
		return
	}
	idRef := codeID
	_ = s.audit.Log(ctx, audit.Entry{
		Action:     "redeem.rollback",
		TargetType: "redeem_code",
		TargetID:   &idRef,
	})
}

// ---------- Disable ----------

func (s *service) Disable(ctx context.Context, actorID int64, ids []int64, reason string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	n, err := s.repo.UpdateToDisabledBulk(ctx, ids)
	if err != nil {
		return 0, apierr.Wrap(apierr.CodeDatabase, "disable failed", err)
	}
	auditAfter, _ := json.Marshal(map[string]any{
		"ids":      ids,
		"reason":   reason,
		"disabled": n,
	})
	actorRef := actorID
	_ = s.audit.Log(ctx, audit.Entry{
		Action:     "redeem.disable",
		TargetType: "redeem_code",
		ActorID:    &actorRef,
		After:      auditAfter,
	})
	return int(n), nil
}

// ---------- List / Get ----------

func (s *service) List(ctx context.Context, f ListFilter) ([]*Code, int64, error) {
	items, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, apierr.Wrap(apierr.CodeDatabase, "list redeem failed", err)
	}
	return items, total, nil
}

func (s *service) Get(ctx context.Context, id int64) (*Code, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, "get redeem failed", err)
	}
	if c == nil {
		return nil, apierr.New(apierr.CodeOrderNotFound, "兑换码不存在")
	}
	return c, nil
}

// ---------- Export ----------

var exportHeaders = []string{
	"id", "code_prefix", "code_display", "amount_quota",
	"batch_no", "status", "used_by", "used_at",
	"expires_at", "created_at",
}

func (s *service) Export(ctx context.Context, w io.Writer, f ListFilter) error {
	items, err := s.repo.ListAll(ctx, f)
	if err != nil {
		return apierr.Wrap(apierr.CodeDatabase, "export list failed", err)
	}
	cw := csv.NewWriter(w)
	if err := cw.Write(exportHeaders); err != nil {
		return apierr.Wrap(apierr.CodeInternal, "csv write header failed", err)
	}
	for _, c := range items {
		row := []string{
			strconv.FormatInt(c.ID, 10),
			c.CodePrefix,
			c.CodePrefix + "-****-****-****",
			strconv.FormatInt(c.AmountQuota, 10),
			c.BatchNo,
			StatusName(c.Status),
			ptrInt64String(c.UsedBy),
			ptrTimeString(c.UsedAt),
			ptrTimeString(c.ExpiresAt),
			c.CreatedAt.UTC().Format(time.RFC3339),
		}
		if err := cw.Write(row); err != nil {
			return apierr.Wrap(apierr.CodeInternal, "csv write row failed", err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return apierr.Wrap(apierr.CodeInternal, "csv flush failed", err)
	}
	rowCount := len(items)
	auditAfter, _ := json.Marshal(map[string]any{
		"filter":    map[string]any{"batch_no": f.BatchNo, "status": f.Status},
		"row_count": rowCount,
	})
	_ = s.audit.Log(ctx, audit.Entry{
		Action:     "redeem.export",
		TargetType: "redeem_batch",
		After:      auditAfter,
	})
	return nil
}

func ptrInt64String(p *int64) string {
	if p == nil {
		return ""
	}
	return strconv.FormatInt(*p, 10)
}

func ptrTimeString(p *time.Time) string {
	if p == nil {
		return ""
	}
	return p.UTC().Format(time.RFC3339)
}

// isUniqueConstraintErr 判断 GORM 错误是否为唯一索引冲突。
// 支持 MySQL / PostgreSQL / sqlite。
func isUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// 同时支持小写与原大小写匹配
	switch {
	case contains(msg, "Duplicate entry"),
		contains(msg, "duplicate key value"),
		contains(msg, "UNIQUE constraint failed"),
		contains(msg, "unique constraint"):
		return true
	}
	return false
}

// contains 是 strings.Contains 的本地复制(避免再 import strings)。
func contains(s, sub string) bool {
	return len(s) >= len(sub) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// 兜底:errors 防未引用
var _ = errors.New
