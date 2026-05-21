package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/pkg/apierr"
	"gorm.io/gorm"
)

// idGenerator 是 repo 对 idgen 的最小依赖,便于测试。
type idGenerator interface {
	Generate() int64
}

// CreateInput 是 Create 调用的参数。
type CreateInput struct {
	UserID        int64
	Name          string
	QuotaLimit    *int64
	AllowedModels []string
	AllowedIPs    []string
	RPMLimit      int
	TPMLimit      int
	ExpiresAt     *time.Time
}

// repo 是 GORM 仓储,封装 api_tokens 表 CRUD。
// 不持有 cache / flusher / audit,这些由 service / store 编排。
type repo struct {
	db    *gorm.DB
	idgen idGenerator
	clk   clock.Clock
	// prefixShowLen 是 prefix 截断的可调长度,Service 层启动时从 setting 读;0 视为默认。
	prefixShowLen int
}

// Authenticate 用明文 key 校验。
//
//   - 校验前缀 / 长度;不符合 → CodeInvalidToken
//   - SELECT WHERE key_hash = sha256(plaintext) AND deleted_at IS NULL
//   - 未命中 / status=disabled → CodeInvalidToken
//   - expires_at 已过 → CodeTokenExpired
//   - 命中返回 *View(不带明文)
func (r *repo) Authenticate(ctx context.Context, plaintext string) (*View, error) {
	if !strings.HasPrefix(plaintext, keyPrefixLit) || len(plaintext) < 8 {
		return nil, apierr.New(apierr.CodeInvalidToken, "token format invalid")
	}
	h := hashPlaintext(plaintext)
	var row Token
	err := r.db.WithContext(ctx).Where("key_hash = ? AND deleted_at IS NULL", h).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierr.New(apierr.CodeInvalidToken, "token not found")
		}
		return nil, apierr.Wrap(apierr.CodeDatabase, err.Error(), err)
	}
	return r.validateState(&row)
}

// AuthenticateByHash 用 sha256 hash 校验,供 cache 命中路径复用一致的状态判断。
// (该方法不返回明文,亦不查 DB —— 由调用方决定走 cache 还是 DB)
func (r *repo) validateState(row *Token) (*View, error) {
	if row.Status == StatusDisabled {
		return nil, apierr.New(apierr.CodeInvalidToken, "token disabled")
	}
	if row.ExpiresAt != nil && !row.ExpiresAt.After(r.clk.Now()) {
		return nil, apierr.New(apierr.CodeTokenExpired, "token expired")
	}
	return row.ToView(), nil
}

// Get 按 id 取详情。不存在返回 CodeNotFound。
func (r *repo) Get(ctx context.Context, id int64) (*View, error) {
	var row Token
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error
	if err != nil {
		return nil, wrapDBErr(err)
	}
	return row.ToView(), nil
}

// getRowByID 是 Get 的内部版本,返回原始 row,供 Revoke / Regenerate 复用。
func (r *repo) getRowByID(ctx context.Context, id int64) (*Token, error) {
	var row Token
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error
	if err != nil {
		return nil, wrapDBErr(err)
	}
	return &row, nil
}

// List 列出令牌。filter.UserID == 0 视为不限。
func (r *repo) List(ctx context.Context, f ListFilter) ([]*View, int64, error) {
	q := r.db.WithContext(ctx).Model(&Token{}).Where("deleted_at IS NULL")
	if f.UserID > 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if k := strings.TrimSpace(f.Keyword); k != "" {
		like := "%" + k + "%"
		q = q.Where("name LIKE ? OR key_prefix LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apierr.Wrap(apierr.CodeDatabase, err.Error(), err)
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}
	size := f.Size
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	var rows []Token
	if err := q.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		return nil, 0, apierr.Wrap(apierr.CodeDatabase, err.Error(), err)
	}
	out := make([]*View, len(rows))
	for i := range rows {
		out[i] = rows[i].ToView()
	}
	return out, total, nil
}

// Create 生成明文 + sha256 + prefix 并写 DB,返回明文与 View。
//
//   - AllowedIPs 必须是有效 CIDR 或单 IP;校验失败 → CodeInvalidParam
//   - AllowedModels 不验合法字符(handler 层做语法校验)
//   - 调用 idgen.Generate 取 id
func (r *repo) Create(ctx context.Context, in CreateInput) (string, *View, error) {
	if err := validateCIDRs(in.AllowedIPs); err != nil {
		return "", nil, err
	}
	plaintext, hash, err := newPlaintext()
	if err != nil {
		return "", nil, apierr.Wrap(apierr.CodeInternal, "generate key: "+err.Error(), err)
	}

	modelsJSON, err := encodeStrings(in.AllowedModels)
	if err != nil {
		return "", nil, apierr.Wrap(apierr.CodeInternal, "encode models: "+err.Error(), err)
	}
	ipsJSON, err := encodeStrings(in.AllowedIPs)
	if err != nil {
		return "", nil, apierr.Wrap(apierr.CodeInternal, "encode ips: "+err.Error(), err)
	}

	now := r.clk.Now().UTC()
	row := Token{
		ID:            r.idgen.Generate(),
		UserID:        in.UserID,
		Name:          in.Name,
		KeyHash:       hash,
		KeyPrefix:     prefixForDisplay(plaintext, r.prefixShowLen),
		QuotaLimit:    in.QuotaLimit,
		QuotaUsed:     0,
		AllowedModels: modelsJSON,
		AllowedIPs:    ipsJSON,
		RPMLimit:      in.RPMLimit,
		TPMLimit:      in.TPMLimit,
		ExpiresAt:     in.ExpiresAt,
		Status:        StatusEnabled,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return "", nil, apierr.Wrap(apierr.CodeDatabase, err.Error(), err)
	}
	return plaintext, row.ToView(), nil
}

// Update 部分字段更新。nil 字段不变。
func (r *repo) Update(ctx context.Context, id int64, p UpdatePatch) (*View, error) {
	if p.Status != nil && (*p.Status != StatusEnabled && *p.Status != StatusDisabled) {
		return nil, apierr.New(apierr.CodeInvalidParam, "status only 0/1 allowed")
	}
	if p.AllowedIPs != nil {
		if err := validateCIDRs(*p.AllowedIPs); err != nil {
			return nil, err
		}
	}
	updates := map[string]any{"updated_at": r.clk.Now().UTC()}
	if p.Name != nil {
		updates["name"] = *p.Name
	}
	if p.ClearQuotaLimit {
		updates["quota_limit"] = nil
	} else if p.QuotaLimit != nil {
		updates["quota_limit"] = *p.QuotaLimit
	}
	if p.AllowedModels != nil {
		raw, err := encodeStrings(*p.AllowedModels)
		if err != nil {
			return nil, apierr.Wrap(apierr.CodeInternal, "encode models: "+err.Error(), err)
		}
		updates["allowed_models"] = raw
	}
	if p.AllowedIPs != nil {
		raw, err := encodeStrings(*p.AllowedIPs)
		if err != nil {
			return nil, apierr.Wrap(apierr.CodeInternal, "encode ips: "+err.Error(), err)
		}
		updates["allowed_ips"] = raw
	}
	if p.RPMLimit != nil {
		updates["rpm_limit"] = *p.RPMLimit
	}
	if p.TPMLimit != nil {
		updates["tpm_limit"] = *p.TPMLimit
	}
	if p.ClearExpiresAt {
		updates["expires_at"] = nil
	} else if p.ExpiresAt != nil {
		updates["expires_at"] = *p.ExpiresAt
	}
	if p.Status != nil {
		updates["status"] = *p.Status
	}

	res := r.db.WithContext(ctx).Model(&Token{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates)
	if res.Error != nil {
		return nil, apierr.Wrap(apierr.CodeDatabase, res.Error.Error(), res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, apierr.New(apierr.CodeNotFound, "token not found")
	}
	return r.Get(ctx, id)
}

// Revoke 软禁用(status=1)。返回旧 key_hash 以便上层失效缓存。
func (r *repo) Revoke(ctx context.Context, id int64) (string, error) {
	row, err := r.getRowByID(ctx, id)
	if err != nil {
		return "", err
	}
	res := r.db.WithContext(ctx).Model(&Token{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"status":     StatusDisabled,
			"updated_at": r.clk.Now().UTC(),
		})
	if res.Error != nil {
		return "", apierr.Wrap(apierr.CodeDatabase, res.Error.Error(), res.Error)
	}
	if res.RowsAffected == 0 {
		return "", apierr.New(apierr.CodeNotFound, "token not found")
	}
	return row.KeyHash, nil
}

// Regenerate 替换 key_hash + key_prefix,保留其他字段。
// 返回新明文、新 View、旧 hash(供上层 Pub/Sub 失效)。
func (r *repo) Regenerate(ctx context.Context, id int64) (string, *View, string, error) {
	row, err := r.getRowByID(ctx, id)
	if err != nil {
		return "", nil, "", err
	}
	plaintext, hash, err := newPlaintext()
	if err != nil {
		return "", nil, "", apierr.Wrap(apierr.CodeInternal, "generate key: "+err.Error(), err)
	}
	prefix := prefixForDisplay(plaintext, r.prefixShowLen)
	res := r.db.WithContext(ctx).Model(&Token{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"key_hash":   hash,
			"key_prefix": prefix,
			"updated_at": r.clk.Now().UTC(),
		})
	if res.Error != nil {
		return "", nil, "", apierr.Wrap(apierr.CodeDatabase, res.Error.Error(), res.Error)
	}
	if res.RowsAffected == 0 {
		return "", nil, "", apierr.New(apierr.CodeNotFound, "token not found")
	}
	updated, err := r.Get(ctx, id)
	if err != nil {
		return "", nil, "", err
	}
	return plaintext, updated, row.KeyHash, nil
}

// flushLastUsed 用 GREATEST + CASE WHEN 批量更新 last_used_at。
// SQLite 没有 GREATEST,这里走 MAX 兼容(SQLite + MySQL + Postgres 都支持 MAX(scalar) 在 UPDATE 子查询里有不同)。
//
// 为简化跨库,采用最小子集:对每个 (id, t) 单独 UPDATE … SET last_used_at = ? WHERE id = ? AND (last_used_at IS NULL OR last_used_at < ?)。
// 在 N <= ~200 时性能可接受(单事务批量)。
func (r *repo) flushLastUsed(ctx context.Context, m map[int64]time.Time) error {
	if len(m) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for id, t := range m {
			ut := t.UTC()
			if err := tx.Exec(
				"UPDATE api_tokens SET last_used_at = ? WHERE id = ? AND (last_used_at IS NULL OR last_used_at < ?)",
				ut, id, ut,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// flushQuota 累加 quota_used。每个 id 单独 UPDATE 防 SQL 注入与跨库差异。
func (r *repo) flushQuota(ctx context.Context, m map[int64]int64) error {
	if len(m) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for id, delta := range m {
			if delta == 0 {
				continue
			}
			if err := tx.Exec(
				"UPDATE api_tokens SET quota_used = quota_used + ? WHERE id = ?",
				delta, id,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// validateCIDRs 校验 IP 列表里每项都能解析为 CIDR 或单 IP。
func validateCIDRs(items []string) error {
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, err := netip.ParsePrefix(s); err == nil {
			continue
		}
		if _, err := netip.ParseAddr(s); err == nil {
			continue
		}
		return apierr.New(apierr.CodeInvalidParam, fmt.Sprintf("invalid CIDR/IP %q", s))
	}
	return nil
}

// encodeStrings 把 []string 编码为 JSON RawMessage,nil 视为 "[]"。
func encodeStrings(s []string) (json.RawMessage, error) {
	if s == nil {
		s = []string{}
	}
	return json.Marshal(s)
}
