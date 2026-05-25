package online

import (
	"context"
	"fmt"

	payprovider "github.com/ijry/pro-api/internal/payment/provider"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"go.uber.org/zap"
)

// WalletCredit is the wallet interface (avoids direct dependency on wallet package).
type WalletCredit interface {
	Credit(ctx context.Context, userID, credits int64, note string) error
}

// InviteRebate is the invite rebate hook interface.
type InviteRebate interface {
	OnOrderPaid(ctx context.Context, orderID, userID int64) error
}

// Deps holds dependencies for NewService.
type Deps struct {
	Repo      Repository
	Providers []payprovider.Provider
	Wallet    WalletCredit
	Invite    InviteRebate // optional
	IDGen     *idgen.Generator
	Clock     clock.Clock
	Log       *zap.Logger
}

// Service handles online payments.
type Service struct {
	deps Deps
}

func NewService(deps Deps) *Service {
	if deps.Clock == nil {
		deps.Clock = clock.Real
	}
	if deps.Log == nil {
		deps.Log = zap.NewNop()
	}
	return &Service{deps: deps}
}

func (s *Service) provider(name string) (payprovider.Provider, bool) {
	for _, p := range s.deps.Providers {
		if p.Name() == name {
			return p, true
		}
	}
	return nil, false
}

// CreateOrderInput is the input for CreateOrder.
type CreateOrderInput struct {
	UserID      int64
	Provider    string
	AmountCents int64
	Currency    string
	Credits     int64
	Description string
	NotifyURL   string
	ReturnURL   string
}

// CreateOrderResult is the output of CreateOrder.
type CreateOrderResult struct {
	OrderID      int64
	OutTradeNo   string
	ClientSecret string
	PayURL       string
}

func (s *Service) CreateOrder(ctx context.Context, in CreateOrderInput) (*CreateOrderResult, error) {
	p, ok := s.provider(in.Provider)
	if !ok {
		return nil, fmt.Errorf("online payment: unknown provider: %s", in.Provider)
	}
	orderID := s.deps.IDGen.Generate()
	outTradeNo := fmt.Sprintf("PRO%d", orderID)
	o := &Order{
		ID:          orderID,
		UserID:      in.UserID,
		Provider:    in.Provider,
		OutTradeNo:  outTradeNo,
		AmountCents: in.AmountCents,
		Currency:    in.Currency,
		Status:      StatusPending,
		Credits:     in.Credits,
	}
	if err := s.deps.Repo.Create(ctx, o); err != nil {
		return nil, fmt.Errorf("online payment: create order: %w", err)
	}
	res, err := p.CreateOrder(ctx, payprovider.CreateOrderInput{
		OutTradeNo:  outTradeNo,
		AmountCents: in.AmountCents,
		Currency:    in.Currency,
		Description: in.Description,
		NotifyURL:   in.NotifyURL,
		ReturnURL:   in.ReturnURL,
	})
	if err != nil {
		return nil, fmt.Errorf("online payment: provider create: %w", err)
	}
	return &CreateOrderResult{
		OrderID:      orderID,
		OutTradeNo:   outTradeNo,
		ClientSecret: res.ClientSecret,
		PayURL:       res.PayURL,
	}, nil
}

// HandleWebhook processes a provider webhook callback (idempotent).
func (s *Service) HandleWebhook(ctx context.Context, providerName string, body []byte, headers map[string]string) error {
	p, ok := s.provider(providerName)
	if !ok {
		return fmt.Errorf("online payment: unknown provider: %s", providerName)
	}
	ev, err := p.HandleWebhook(ctx, body, headers)
	if err != nil {
		return err
	}
	if ev == nil || !ev.Paid {
		return nil
	}
	o, err := s.deps.Repo.FindByOutTradeNo(ctx, ev.OutTradeNo)
	if err != nil {
		return fmt.Errorf("online payment: find order %s: %w", ev.OutTradeNo, err)
	}
	if o.Status == StatusPaid {
		return nil // idempotent
	}
	now := s.deps.Clock.Now()
	if err := s.deps.Repo.UpdateStatus(ctx, o.ID, StatusPaid, ev.ProviderOrderID, &now); err != nil {
		return fmt.Errorf("online payment: update order: %w", err)
	}
	if s.deps.Wallet != nil {
		if err := s.deps.Wallet.Credit(ctx, o.UserID, o.Credits, fmt.Sprintf("充值订单 %s", o.OutTradeNo)); err != nil {
			s.deps.Log.Error("online payment: wallet credit failed", zap.Error(err), zap.Int64("order_id", o.ID))
		}
	}
	if s.deps.Invite != nil {
		if err := s.deps.Invite.OnOrderPaid(ctx, o.ID, o.UserID); err != nil {
			s.deps.Log.Warn("online payment: invite rebate failed", zap.Error(err), zap.Int64("order_id", o.ID))
		}
	}
	return nil
}
