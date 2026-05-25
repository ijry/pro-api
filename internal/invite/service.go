package invite

import (
	"context"
	"fmt"

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
	DB      *gorm.DB       // raw DB for cross-package queries without import cycles
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
		Scan(&invitedBy).Error; err != nil || invitedBy == 0 {
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
