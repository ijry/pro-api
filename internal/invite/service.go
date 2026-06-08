package invite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WalletCredit is the wallet interface (avoids direct dependency).
type WalletCredit interface {
	Credit(ctx context.Context, userID, credits int64, note string) error
}

// Deps holds dependencies for NewService.
type Deps struct {
	Repo    Repository
	DB      *gorm.DB // raw DB for cross-package queries without import cycles
	Wallet  WalletCredit
	Setting setting.Store
	IDGen   *idgen.Generator
	Clock   clock.Clock
	Log     *zap.Logger
}

// Service handles invite rebate logic.
type Service struct{ deps Deps }

func NewService(deps Deps) *Service {
	if deps.Clock == nil {
		deps.Clock = clock.Real
	}
	if deps.Log == nil {
		deps.Log = zap.NewNop()
	}
	return &Service{deps: deps}
}

// OnOrderPaid implements online.InviteRebate.
// Called after a payment order is marked paid.
func (s *Service) OnOrderPaid(ctx context.Context, orderID, userID int64) error {
	// 1. Find inviter via users.invited_by
	var invitedBy int64
	if err := s.deps.DB.WithContext(ctx).
		Table("users").
		Select("invited_by").
		Where("id = ?", userID).
		Scan(&invitedBy).Error; err != nil {
		return fmt.Errorf("invite: query inviter for user %d: %w", userID, err)
	}
	if invitedBy == 0 {
		return nil // no inviter
	}

	// 2. Get order amount from payment_orders
	var order struct {
		AmountCents int64
		Credits     int64
	}
	if err := s.deps.DB.WithContext(ctx).
		Table("payment_orders").
		Select("amount_cents, credits").
		Where("id = ?", orderID).
		Scan(&order).Error; err != nil {
		return fmt.Errorf("invite: query order %d: %w", orderID, err)
	}

	// 3. Calculate rebate
	ratio := s.getRebateRatio(ctx)
	creditPerCent := s.getCreditPerCent(ctx)
	rebateCents := int64(float64(order.AmountCents) * ratio)
	rebateCredits := int64(float64(rebateCents) * creditPerCent)

	// 4. Create invite record
	rec := &Record{
		ID:            s.deps.IDGen.Generate(),
		InviterID:     invitedBy,
		InviteeID:     userID,
		OrderID:       orderID,
		RebateCents:   rebateCents,
		RebateCredits: rebateCredits,
		CreatedAt:     s.deps.Clock.Now(),
	}
	if err := s.deps.Repo.Create(ctx, rec); err != nil {
		return fmt.Errorf("invite: create record: %w", err)
	}

	// 5. Credit wallet
	if rebateCredits > 0 && s.deps.Wallet != nil {
		if err := s.deps.Wallet.Credit(ctx, invitedBy, rebateCredits,
			fmt.Sprintf("邀请返佣(订单 %d)", orderID)); err != nil {
			s.deps.Log.Warn("invite: wallet credit failed",
				zap.Int64("inviter", invitedBy), zap.Error(err))
		}
	}
	return nil
}

func (s *Service) getRebateRatio(ctx context.Context) float64 {
	if s.deps.Setting == nil {
		return 0.1
	}
	return s.deps.Setting.GetFloat(ctx, "invite.rebate_ratio", 0.1)
}

func (s *Service) getCreditPerCent(ctx context.Context) float64 {
	if s.deps.Setting == nil {
		return 1.0
	}
	return s.deps.Setting.GetFloat(ctx, "invite.credit_per_cent", 1.0)
}

// --- View types (match frontend JSON contract) ---

// SummaryResp is the response for GET /api/user/invites/me.
type SummaryResp struct {
	InviteCode  string    `json:"invite_code"`
	ShareURL    string    `json:"share_url"`
	RebateRatio float64   `json:"rebate_ratio"`
	Stats       StatsResp `json:"stats"`
}

// StatsResp holds the three counters inside SummaryResp.
type StatsResp struct {
	InviteeCount       int64 `json:"invitee_count"`
	RebateCreditsTotal int64 `json:"rebate_credits_total"`
	RebateCreditsMonth int64 `json:"rebate_credits_month"`
}

// InviteeView is one row in GET /api/user/invites/invitees.
type InviteeView struct {
	UserID             int64  `json:"user_id"`
	DisplayName        string `json:"display_name"`
	EmailMasked        string `json:"email_masked"`
	RegisteredAt       string `json:"registered_at"`
	TotalRebateCredits int64  `json:"total_rebate_credits"`
}

// RecordView is one row in GET /api/user/invites/records.
type RecordView struct {
	ID                 int64  `json:"id"`
	InviteeID          int64  `json:"invitee_id"`
	InviteeDisplayName string `json:"invitee_display_name"`
	OrderID            int64  `json:"order_id"`
	RebateCents        int64  `json:"rebate_cents"`
	RebateCredits      int64  `json:"rebate_credits"`
	CreatedAt          string `json:"created_at"`
}

// maskEmail converts "john@example.com" → "j***@example.com".
func maskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return "***"
	}
	return string(email[0]) + "***" + email[at:]
}

// GetSummary returns the current user's invite summary (code, share URL, stats).
func (s *Service) GetSummary(ctx context.Context, userID int64) (*SummaryResp, error) {
	var u struct {
		InviteCode *string
	}
	if err := s.deps.DB.WithContext(ctx).Table("users").
		Select("invite_code").
		Where("id = ? AND deleted_at IS NULL", userID).
		Scan(&u).Error; err != nil {
		return nil, fmt.Errorf("invite: get user: %w", err)
	}
	inviteCode := ""
	if u.InviteCode != nil {
		inviteCode = *u.InviteCode
	}

	var inviteeCount int64
	if err := s.deps.DB.WithContext(ctx).Table("users").
		Where("invited_by = ? AND deleted_at IS NULL", userID).
		Count(&inviteeCount).Error; err != nil {
		return nil, fmt.Errorf("invite: count invitees: %w", err)
	}

	var totalCredits int64
	s.deps.DB.WithContext(ctx).Table("invite_records").
		Select("COALESCE(SUM(rebate_credits), 0)").
		Where("inviter_id = ?", userID).
		Scan(&totalCredits)

	now := s.deps.Clock.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var monthCredits int64
	s.deps.DB.WithContext(ctx).Table("invite_records").
		Select("COALESCE(SUM(rebate_credits), 0)").
		Where("inviter_id = ? AND created_at >= ?", userID, monthStart).
		Scan(&monthCredits)

	baseURL := ""
	if s.deps.Setting != nil {
		baseURL = s.deps.Setting.GetString(ctx, "site.base_url", "")
	}
	shareURL := ""
	if inviteCode != "" {
		shareURL = baseURL + "/register?invite_code=" + inviteCode
	}

	return &SummaryResp{
		InviteCode:  inviteCode,
		ShareURL:    shareURL,
		RebateRatio: s.getRebateRatio(ctx),
		Stats: StatsResp{
			InviteeCount:       inviteeCount,
			RebateCreditsTotal: totalCredits,
			RebateCreditsMonth: monthCredits,
		},
	}, nil
}

// ListInvitees returns paginated users who registered via this user's invite code.
func (s *Service) ListInvitees(ctx context.Context, inviterID int64, page, size int) ([]*InviteeView, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	offset := (page - 1) * size

	var total int64
	if err := s.deps.DB.WithContext(ctx).Table("users").
		Where("invited_by = ? AND deleted_at IS NULL", inviterID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("invite: count invitees: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	type row struct {
		ID                 int64
		DisplayName        *string
		Email              *string
		CreatedAt          time.Time
		TotalRebateCredits int64
	}
	var rows []row
	err := s.deps.DB.WithContext(ctx).Table("users").
		Select("users.id, users.display_name, users.email, users.created_at, COALESCE(SUM(ir.rebate_credits), 0) AS total_rebate_credits").
		Joins("LEFT JOIN invite_records ir ON ir.invitee_id = users.id AND ir.inviter_id = ?", inviterID).
		Where("users.invited_by = ? AND users.deleted_at IS NULL", inviterID).
		Group("users.id, users.display_name, users.email, users.created_at").
		Order("users.created_at DESC").
		Offset(offset).Limit(size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("invite: list invitees: %w", err)
	}

	views := make([]*InviteeView, len(rows))
	for i, r := range rows {
		dn := ""
		if r.DisplayName != nil {
			dn = *r.DisplayName
		}
		email := ""
		if r.Email != nil {
			email = *r.Email
		}
		views[i] = &InviteeView{
			UserID:             r.ID,
			DisplayName:        dn,
			EmailMasked:        maskEmail(email),
			RegisteredAt:       r.CreatedAt.UTC().Format(time.RFC3339),
			TotalRebateCredits: r.TotalRebateCredits,
		}
	}
	return views, total, nil
}

// ListRecords returns paginated rebate records where this user is the inviter.
func (s *Service) ListRecords(ctx context.Context, inviterID int64, page, size int) ([]*RecordView, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	offset := (page - 1) * size

	total, err := s.deps.Repo.CountByInviter(ctx, inviterID)
	if err != nil {
		return nil, 0, fmt.Errorf("invite: count records: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	records, err := s.deps.Repo.ListByInviter(ctx, inviterID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("invite: list records: %w", err)
	}
	if len(records) == 0 {
		return nil, total, nil
	}

	inviteeIDs := make([]int64, len(records))
	for i, r := range records {
		inviteeIDs[i] = r.InviteeID
	}
	type nameRow struct {
		ID          int64
		DisplayName *string
	}
	var nameRows []nameRow
	s.deps.DB.WithContext(ctx).Table("users").
		Select("id, display_name").
		Where("id IN ?", inviteeIDs).
		Scan(&nameRows)
	nameMap := make(map[int64]string, len(nameRows))
	for _, nr := range nameRows {
		if nr.DisplayName != nil {
			nameMap[nr.ID] = *nr.DisplayName
		}
	}

	views := make([]*RecordView, len(records))
	for i, r := range records {
		views[i] = &RecordView{
			ID:                 r.ID,
			InviteeID:          r.InviteeID,
			InviteeDisplayName: nameMap[r.InviteeID],
			OrderID:            r.OrderID,
			RebateCents:        r.RebateCents,
			RebateCredits:      r.RebateCredits,
			CreatedAt:          r.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return views, total, nil
}
