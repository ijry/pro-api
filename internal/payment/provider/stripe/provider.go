package stripe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ijry/pro-api/internal/payment/provider"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

// Config is the Stripe provider config.
type Config struct {
	SecretKey     string
	WebhookSecret string
}

type stripeProvider struct {
	cfg Config
}

// New creates a Stripe payment provider.
func New(cfg Config) provider.Provider {
	stripe.Key = cfg.SecretKey
	return &stripeProvider{cfg: cfg}
}

func (p *stripeProvider) Name() string { return "stripe" }

func (p *stripeProvider) CreateOrder(_ context.Context, in provider.CreateOrderInput) (*provider.CreateOrderResult, error) {
	currency := in.Currency
	if currency == "" {
		currency = "usd"
	}
	// Stripe amount is in smallest currency unit (cents for USD)
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(in.AmountCents),
		Currency:    stripe.String(currency),
		Description: stripe.String(in.Description),
	}
	params.AddMetadata("out_trade_no", in.OutTradeNo)
	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe: create PaymentIntent: %w", err)
	}
	return &provider.CreateOrderResult{
		ProviderOrderID: pi.ID,
		ClientSecret:    pi.ClientSecret,
	}, nil
}

func (p *stripeProvider) HandleWebhook(_ context.Context, body []byte, headers map[string]string) (*provider.WebhookEvent, error) {
	sig := headers["Stripe-Signature"]
	if sig == "" {
		sig = headers["stripe-signature"]
	}
	event, err := webhook.ConstructEvent(body, sig, p.cfg.WebhookSecret)
	if err != nil {
		return nil, fmt.Errorf("stripe: webhook verify: %w", err)
	}
	if event.Type != "payment_intent.succeeded" {
		return nil, nil
	}
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return nil, fmt.Errorf("stripe: decode PaymentIntent: %w", err)
	}
	return &provider.WebhookEvent{
		OutTradeNo:      pi.Metadata["out_trade_no"],
		ProviderOrderID: pi.ID,
		Paid:            true,
		AmountCents:     pi.Amount,
		Currency:        string(pi.Currency),
		Raw:             body,
	}, nil
}

func (p *stripeProvider) QueryOrder(_ context.Context, outTradeNo string) (*provider.WebhookEvent, error) {
	params := &stripe.PaymentIntentSearchParams{}
	params.Query = fmt.Sprintf("metadata['out_trade_no']:'%s'", outTradeNo)
	iter := paymentintent.Search(params)
	for iter.Next() {
		pi := iter.PaymentIntent()
		return &provider.WebhookEvent{
			OutTradeNo:      outTradeNo,
			ProviderOrderID: pi.ID,
			Paid:            pi.Status == stripe.PaymentIntentStatusSucceeded,
			AmountCents:     pi.Amount,
			Currency:        string(pi.Currency),
		}, nil
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("stripe: search PaymentIntent: %w", err)
	}
	return nil, fmt.Errorf("stripe: order not found: %s", outTradeNo)
}
