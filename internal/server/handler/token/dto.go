package token

import (
	"time"

	tokensvc "github.com/ijry/pro-api/internal/token"
)

// ViewDTO 是 list / detail / update 响应中 token 的 JSON 形态。
type ViewDTO struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	QuotaLimit    *int64     `json:"quota_limit"`
	QuotaUsed     int64      `json:"quota_used"`
	AllowedModels []string   `json:"allowed_models"`
	AllowedIPs    []string   `json:"allowed_ips"`
	RPMLimit      int        `json:"rpm_limit"`
	TPMLimit      int        `json:"tpm_limit"`
	ExpiresAt     *time.Time `json:"expires_at"`
	LastUsedAt    *time.Time `json:"last_used_at"`
	Status        int8       `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}

func toDTO(v *tokensvc.View) ViewDTO {
	out := ViewDTO{
		ID:            v.ID,
		UserID:        v.UserID,
		Name:          v.Name,
		KeyPrefix:     v.KeyPrefix,
		QuotaLimit:    v.QuotaLimit,
		QuotaUsed:     v.QuotaUsed,
		AllowedModels: v.AllowedModels,
		AllowedIPs:    v.AllowedIPs,
		RPMLimit:      v.RPMLimit,
		TPMLimit:      v.TPMLimit,
		ExpiresAt:     v.ExpiresAt,
		LastUsedAt:    v.LastUsedAt,
		Status:        v.Status,
		CreatedAt:     v.CreatedAt,
	}
	if out.AllowedModels == nil {
		out.AllowedModels = []string{}
	}
	if out.AllowedIPs == nil {
		out.AllowedIPs = []string{}
	}
	return out
}

// ListResponse 是分页响应结构。
type ListResponse struct {
	Items []ViewDTO `json:"items"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

// CreateRequest 是 POST /api/{user,admin}/tokens 的请求体。
type CreateRequest struct {
	Name          string     `json:"name"`
	QuotaLimit    *int64     `json:"quota_limit"`
	AllowedModels []string   `json:"allowed_models"`
	AllowedIPs    []string   `json:"allowed_ips"`
	RPMLimit      int        `json:"rpm_limit"`
	TPMLimit      int        `json:"tpm_limit"`
	ExpiresAt     *time.Time `json:"expires_at"`
}

// CreateResponse 一次性返回明文 key。
type CreateResponse struct {
	View          ViewDTO `json:"view"`
	PlaintextKey  string  `json:"plaintext_key"`
	Warning       string  `json:"warning"`
}

const plaintextWarning = "This is the only time you can view the plaintext key. Store it securely."

// PatchRequest 是 PATCH /api/{user,admin}/tokens/{id} 的请求体。
// 字段全部可选;ClearQuotaLimit / ClearExpiresAt 与对应字段配合实现"清空"。
type PatchRequest struct {
	Name            *string    `json:"name"`
	QuotaLimit      *int64     `json:"quota_limit"`
	ClearQuotaLimit bool       `json:"clear_quota_limit"`
	AllowedModels   *[]string  `json:"allowed_models"`
	AllowedIPs      *[]string  `json:"allowed_ips"`
	RPMLimit        *int       `json:"rpm_limit"`
	TPMLimit        *int       `json:"tpm_limit"`
	ExpiresAt       *time.Time `json:"expires_at"`
	ClearExpiresAt  bool       `json:"clear_expires_at"`
	Status          *int8      `json:"status"`
}

func (p PatchRequest) toUpdatePatch() tokensvc.UpdatePatch {
	return tokensvc.UpdatePatch{
		Name:            p.Name,
		QuotaLimit:      p.QuotaLimit,
		ClearQuotaLimit: p.ClearQuotaLimit,
		AllowedModels:   p.AllowedModels,
		AllowedIPs:      p.AllowedIPs,
		RPMLimit:        p.RPMLimit,
		TPMLimit:        p.TPMLimit,
		ExpiresAt:       p.ExpiresAt,
		ClearExpiresAt:  p.ClearExpiresAt,
		Status:          p.Status,
	}
}
