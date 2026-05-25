package provider

import "context"

// CreateOrderInput is the input for creating a payment order.
type CreateOrderInput struct {
	OutTradeNo  string
	AmountCents int64
	Currency    string
	Description string
	NotifyURL   string
	ReturnURL   string
	Extra       map[string]string
}

// CreateOrderResult is returned after creating an order with a provider.
type CreateOrderResult struct {
	ProviderOrderID string
	PayURL          string
	QRCode          string
	ClientSecret    string // Stripe PaymentIntent client_secret
}

// WebhookEvent is the normalized webhook event from a provider.
type WebhookEvent struct {
	OutTradeNo      string
	ProviderOrderID string
	Paid            bool
	AmountCents     int64
	Currency        string
	Raw             []byte
}

// Provider is the payment provider abstraction.
type Provider interface {
	Name() string
	CreateOrder(ctx context.Context, in CreateOrderInput) (*CreateOrderResult, error)
	// HandleWebhook parses the webhook body+headers and verifies the signature.
	// Returns nil event (no error) for non-payment-success events.
	HandleWebhook(ctx context.Context, body []byte, headers map[string]string) (*WebhookEvent, error)
	QueryOrder(ctx context.Context, outTradeNo string) (*WebhookEvent, error)
}
