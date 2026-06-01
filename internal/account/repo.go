package account

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/crypto"
	"github.com/ijry/pro-api/internal/util/idgen"
	"github.com/ijry/pro-api/pkg/apierr"
	"gorm.io/gorm"
)

type repo struct {
	db     *gorm.DB
	crypto *crypto.AESGCM
	id     *idgen.Generator
	clock  clock.Clock
}

// NewRepository 构造 Repo。
func NewRepository(db *gorm.DB, c *crypto.AESGCM, id *idgen.Generator, clk clock.Clock) Repo {
	return &repo{db: db, crypto: c, id: id, clock: clk}
}

func (r *repo) Create(ctx context.Context, a *Account) error {
	if err := r.encryptCred(a); err != nil {
		return err
	}
	if a.ID == 0 {
		a.ID = r.id.Generate()
	}
	now := r.clock.Now().UTC()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	if len(a.Extra) == 0 {
		a.Extra = json.RawMessage("{}")
	}
	return r.db.WithContext(ctx).Create(a).Error
}

// Update applies non-zero fields of a to the row keyed by a.ID. Zero-value fields
// are skipped (GORM struct update semantics). To re-encrypt credentials, populate
// a.Cred with at least one non-empty token field; otherwise the stored ciphertext
// is preserved. To intentionally clear a non-cred field to zero, use a map-based
// update instead (not exposed here).
func (r *repo) Update(ctx context.Context, a *Account) error {
	if a.ID == 0 {
		return apierr.New(apierr.CodeInvalidParam, "account: update requires ID")
	}
	// Only re-encrypt when caller populated any cred field.
	if a.Cred.AccessToken != "" || a.Cred.RefreshToken != "" || a.Cred.APIKey != "" || a.Cred.IDToken != "" {
		if err := r.encryptCred(a); err != nil {
			return err
		}
	} else {
		// Caller did not set cred; ensure we don't accidentally clobber stored ciphertext.
		a.Credentials = ""
	}
	a.UpdatedAt = r.clock.Now().UTC()
	// db.Updates with a struct skips zero-value fields — including empty Credentials.
	return r.db.WithContext(ctx).Model(&Account{}).Where("id = ?", a.ID).Updates(a).Error
}

// Reactivate 把账号从 cooldown 恢复到 active。使用 map-based 更新,
// 因为 GORM 的 Updates(struct) 会跳过零值字段(StatusActive=0 / cooldown_until=nil / consec_failures=0)。
func (r *repo) Reactivate(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&Account{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":          StatusActive,
			"cooldown_until":  nil,
			"consec_failures": 0,
			"updated_at":      r.clock.Now().UTC(),
		}).Error
}

// ResetFailures 把 consec_failures 清零并刷新 last_success_at / last_used_at。
// 同样用 map-based 写入避免零值跳过。
func (r *repo) ResetFailures(ctx context.Context, id int64) error {
	now := r.clock.Now().UTC()
	return r.db.WithContext(ctx).Model(&Account{}).Where("id = ?", id).
		Updates(map[string]any{
			"consec_failures": 0,
			"last_success_at": now,
			"last_used_at":    now,
			"updated_at":      now,
		}).Error
}

func (r *repo) Get(ctx context.Context, id int64) (*Account, error) {
	var a Account
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apierr.New(apierr.CodeAccountNotFound, "account not found")
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repo) ListByChannel(ctx context.Context, channelID int64) ([]*Account, error) {
	var out []*Account
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND deleted_at IS NULL", channelID).
		Order("priority DESC, weight DESC, id DESC").Find(&out).Error
	if err != nil {
		return nil, err
	}
	for _, a := range out {
		// TODO(P8): log hydrate failures via repo logger
		_ = r.hydrate(a)
	}
	return out, nil
}

func (r *repo) ListByShareTag(ctx context.Context, tag string) ([]*Account, error) {
	var out []*Account
	err := r.db.WithContext(ctx).
		Where("share_tag = ? AND deleted_at IS NULL", tag).
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	for _, a := range out {
		// TODO(P8): log hydrate failures via repo logger
		_ = r.hydrate(a)
	}
	return out, nil
}

func (r *repo) ListForRefresher(ctx context.Context, before time.Time, limit int) ([]*Account, error) {
	var out []*Account
	err := r.db.WithContext(ctx).
		Where("cred_type = 'oauth' AND status = ? AND access_token_expires_at IS NOT NULL AND access_token_expires_at < ? AND deleted_at IS NULL",
			StatusActive, before).
		Limit(limit).Find(&out).Error
	if err != nil {
		return nil, err
	}
	for _, a := range out {
		// TODO(P8): log hydrate failures via repo logger
		_ = r.hydrate(a)
	}
	return out, nil
}

func (r *repo) ListForReaper(ctx context.Context, now time.Time, limit int) ([]*Account, error) {
	var out []*Account
	err := r.db.WithContext(ctx).
		Where("status = ? AND cooldown_until IS NOT NULL AND cooldown_until <= ? AND deleted_at IS NULL",
			StatusCooldown, now).
		Limit(limit).Find(&out).Error
	if err != nil {
		return nil, err
	}
	for _, a := range out {
		// TODO(P8): log hydrate failures via repo logger
		_ = r.hydrate(a)
	}
	return out, nil
}

func (r *repo) Delete(ctx context.Context, id int64) error {
	now := r.clock.Now().UTC()
	return r.db.WithContext(ctx).Model(&Account{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *repo) AppendEvent(ctx context.Context, accountID int64, eventType string, payload any) error {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return apierr.New(apierr.CodeInvalidParam, "encode event payload: "+err.Error())
		}
		raw = b
	}
	ev := &AccountEvent{
		ID:        r.id.Generate(),
		AccountID: accountID,
		EventType: eventType,
		Payload:   raw,
		CreatedAt: r.clock.Now().UTC(),
	}
	return r.db.WithContext(ctx).Create(ev).Error
}

func (r *repo) encryptCred(a *Account) error {
	b, err := json.Marshal(a.Cred)
	if err != nil {
		return apierr.New(apierr.CodeInvalidParam, "encode cred: "+err.Error())
	}
	enc, err := r.crypto.Encrypt(string(b))
	if err != nil {
		return apierr.New(apierr.CodeInternal, "encrypt cred: "+err.Error())
	}
	a.Credentials = enc
	return nil
}

func (r *repo) hydrate(a *Account) error {
	if a.Credentials == "" {
		return nil
	}
	raw, err := r.crypto.Decrypt(a.Credentials)
	if err != nil {
		return apierr.New(apierr.CodeInternal,
			"decrypt cred failed for account "+strconv.FormatInt(a.ID, 10))
	}
	return json.Unmarshal([]byte(raw), &a.Cred)
}
