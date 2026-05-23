package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/crypto"
	"github.com/ijry/pro-api/internal/util/idgen"
	"github.com/ijry/pro-api/pkg/apierr"
	"gorm.io/gorm"
)

// repo 是 channels / channel_model_mappings 的 ORM 仓储。
type repo struct {
	db     *gorm.DB
	crypto *crypto.AESGCM
	id     *idgen.Generator
	clock  clock.Clock
}

func newRepo(db *gorm.DB, c *crypto.AESGCM, id *idgen.Generator, clk clock.Clock) *repo {
	return &repo{db: db, crypto: c, id: id, clock: clk}
}

// SaveChannel 创建或更新一个 channel(upsert by id)。
// 若 ch.Cred.APIKey != "" 则重新加密写入。
func (r *repo) SaveChannel(ctx context.Context, ch *Channel) error {
	if !ch.Cred.IsZero() {
		b, err := json.Marshal(ch.Cred)
		if err != nil {
			return apierr.New(apierr.CodeInvalidParam, "encode credentials: "+err.Error())
		}
		enc, err := r.crypto.Encrypt(string(b))
		if err != nil {
			return apierr.New(apierr.CodeInternal, "encrypt credentials: "+err.Error())
		}
		ch.Credentials = enc
	}
	if ch.ID == 0 {
		ch.ID = r.id.Generate()
	}
	now := r.clock.Now().UTC()
	if ch.CreatedAt.IsZero() {
		ch.CreatedAt = now
	}
	ch.UpdatedAt = now
	if ch.Tags == nil {
		ch.Tags = json.RawMessage("[]")
	}
	if ch.Extra == nil {
		ch.Extra = json.RawMessage("{}")
	}
	return r.db.WithContext(ctx).Save(ch).Error
}

// GetByID 按主键加载 channel(软删过滤)。
func (r *repo) GetByID(ctx context.Context, id int64) (*Channel, error) {
	var ch Channel
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&ch).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apierr.New(apierr.CodeNotFound, "channel not found")
	}
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// GetByIDIncludingDeleted 包含软删的记录(cache.Reload 用)。
func (r *repo) GetByIDIncludingDeleted(ctx context.Context, id int64) (*Channel, error) {
	var ch Channel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ch).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// ListQuery 列表筛选。
type listQuery struct {
	Provider string
	Status   *int8
	Tag      string
	Keyword  string
	Page     int
	PageSize int
}

// ListChannels 分页查询。
func (r *repo) ListChannels(ctx context.Context, q listQuery) ([]*Channel, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	tx := r.db.WithContext(ctx).Model(&Channel{}).Where("deleted_at IS NULL")
	if q.Provider != "" {
		tx = tx.Where("provider = ?", q.Provider)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.Keyword != "" {
		tx = tx.Where("name LIKE ?", "%"+q.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*Channel
	if err := tx.Order("priority DESC, weight DESC, id DESC").
		Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// SoftDelete 软删一条 channel。
func (r *repo) SoftDelete(ctx context.Context, id int64) error {
	now := r.clock.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&Channel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

// UpdateStatus 更新 status 字段。
func (r *repo) UpdateStatus(ctx context.Context, id int64, status int8) error {
	now := r.clock.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&Channel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"status": status, "updated_at": now}).Error
}

// ListAllActive 全量加载所有未软删的 channels(cache warmup 用)。
func (r *repo) ListAllActive(ctx context.Context) ([]*Channel, error) {
	var items []*Channel
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("priority DESC, weight DESC, id DESC").
		Find(&items).Error
	return items, err
}

// ListMappings 查询一个 channel 的所有模型映射。
func (r *repo) ListMappings(ctx context.Context, channelID int64) ([]*ModelMapping, error) {
	var items []*ModelMapping
	err := r.db.WithContext(ctx).Where("channel_id = ?", channelID).Find(&items).Error
	return items, err
}

// ListMappingsByChannelIDs 批量查询多个 channel 的映射。
func (r *repo) ListMappingsByChannelIDs(ctx context.Context, ids []int64) ([]*ModelMapping, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var items []*ModelMapping
	err := r.db.WithContext(ctx).Where("channel_id IN ?", ids).Find(&items).Error
	return items, err
}

// ReplaceMappings 批量替换(事务内全删全建)。
func (r *repo) ReplaceMappings(ctx context.Context, channelID int64, mappings []*ModelMapping) error {
	now := r.clock.Now().UTC()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("channel_id = ?", channelID).Delete(&ModelMapping{}).Error; err != nil {
			return err
		}
		if len(mappings) == 0 {
			return nil
		}
		for _, m := range mappings {
			if m.ID == 0 {
				m.ID = r.id.Generate()
			}
			m.ChannelID = channelID
			if m.CreatedAt.IsZero() {
				m.CreatedAt = now
			}
			m.UpdatedAt = now
		}
		return tx.Create(mappings).Error
	})
}

// hydrate 解密 credentials 并填入 ch.Cred。
func (r *repo) hydrate(ch *Channel) error {
	if ch.Credentials == "" {
		return nil
	}
	raw, err := r.crypto.Decrypt(ch.Credentials)
	if err != nil {
		return apierr.New(apierr.CodeChannelMisconfig,
			"decrypt credentials failed for channel "+strconv.FormatInt(ch.ID, 10))
	}
	return json.Unmarshal([]byte(raw), &ch.Cred)
}

// DeleteBreakerKey 删除 Redis breaker hash(软删 channel 时调用)。
// 这里只返回 channel id,由 service 调用 breaker 清理。
func (r *repo) ChannelExists(ctx context.Context, id int64) (bool, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Model(&Channel{}).
		Where("id = ? AND deleted_at IS NULL", id).Count(&cnt).Error
	return cnt > 0, err
}

// NamesByIDs 批量查渠道名(log handler label 用)。
func (r *repo) NamesByIDs(ctx context.Context, ids []int64) map[int64]string {
	if len(ids) == 0 {
		return nil
	}
	var items []struct {
		ID   int64
		Name string
	}
	_ = r.db.WithContext(ctx).Model(&Channel{}).
		Select("id, name").Where("id IN ?", ids).Scan(&items).Error
	m := make(map[int64]string, len(items))
	for _, item := range items {
		m[item.ID] = item.Name
	}
	return m
}


// int64SliceIDs 把 []*Channel 转成 id 切片。
func int64SliceIDs(chs []*Channel) []int64 {
	ids := make([]int64, len(chs))
	for i, c := range chs {
		ids[i] = c.ID
	}
	return ids
}

// buildModelMap 从 mappings 构建 clientModel->upstreamModel 映射。
func buildModelMap(mappings []*ModelMapping) (map[string]string, map[string]ModelOverride) {
	mm := make(map[string]string, len(mappings))
	mo := make(map[string]ModelOverride, len(mappings))
	for _, m := range mappings {
		mm[m.ClientModel] = m.UpstreamModel
		mo[m.ClientModel] = ModelOverride{
			InputRatio:     m.InputRatio,
			OutputRatio:    m.OutputRatio,
			CachedRatio:    m.CachedRatio,
			ReasoningRatio: m.ReasoningRatio,
		}
	}
	return mm, mo
}

// unused helper suppressed via blank
var _ = fmt.Sprintf
