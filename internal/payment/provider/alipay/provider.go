package alipay

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	alipaysdk "github.com/smartwalle/alipay/v3"

	"github.com/ijry/pro-api/internal/payment/provider"
)

// Config is the Alipay provider config.
type Config struct {
	AppID      string
	PrivateKey string // PKCS1 RSA private key (PEM)
	PublicKey  string // Alipay public key (PEM)
	Sandbox    bool
}

type alipayProvider struct {
	client *alipaysdk.Client
}

// New creates an Alipay payment provider.
// Returns error if client creation or public key loading fails.
func New(cfg Config) (provider.Provider, error) {
	// production=true means real environment; Sandbox=true means test environment
	production := !cfg.Sandbox
	c, err := alipaysdk.New(cfg.AppID, cfg.PrivateKey, production)
	if err != nil {
		return nil, fmt.Errorf("alipay: create client: %w", err)
	}
	if cfg.PublicKey != "" {
		if err := c.LoadAliPayPublicKey(cfg.PublicKey); err != nil {
			return nil, fmt.Errorf("alipay: load public key: %w", err)
		}
	}
	return &alipayProvider{client: c}, nil
}

func (p *alipayProvider) Name() string { return "alipay" }

// CreateOrder builds a TradePagePay URL (网页支付) and returns it as PayURL.
func (p *alipayProvider) CreateOrder(_ context.Context, in provider.CreateOrderInput) (*provider.CreateOrderResult, error) {
	totalAmount := fmt.Sprintf("%.2f", float64(in.AmountCents)/100)
	subject := in.Description
	if subject == "" {
		subject = in.OutTradeNo
	}

	param := alipaysdk.TradePagePay{}
	param.NotifyURL = in.NotifyURL
	param.ReturnURL = in.ReturnURL
	param.Subject = subject
	param.OutTradeNo = in.OutTradeNo
	param.TotalAmount = totalAmount
	param.ProductCode = "FAST_INSTANT_TRADE_PAY"

	payURL, err := p.client.TradePagePay(param)
	if err != nil {
		return nil, fmt.Errorf("alipay: TradePagePay: %w", err)
	}
	return &provider.CreateOrderResult{
		PayURL: payURL.String(),
	}, nil
}

// HandleWebhook verifies an Alipay async notification and returns a WebhookEvent on
// TRADE_SUCCESS or TRADE_FINISHED. Returns nil event (no error) for other statuses.
func (p *alipayProvider) HandleWebhook(ctx context.Context, body []byte, _ map[string]string) (*provider.WebhookEvent, error) {
	// Alipay sends async notifications as application/x-www-form-urlencoded POST bodies.
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("alipay: parse notification body: %w", err)
	}

	notification, err := p.client.DecodeNotification(ctx, values)
	if err != nil {
		return nil, fmt.Errorf("alipay: decode notification: %w", err)
	}

	// Only treat TRADE_SUCCESS and TRADE_FINISHED as paid.
	if notification.TradeStatus != alipaysdk.TradeStatusSuccess &&
		notification.TradeStatus != alipaysdk.TradeStatusFinished {
		return nil, nil
	}

	amountCents, err := yuanToCents(notification.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("alipay: parse total_amount %q: %w", notification.TotalAmount, err)
	}

	return &provider.WebhookEvent{
		OutTradeNo:      notification.OutTradeNo,
		ProviderOrderID: notification.TradeNo,
		Paid:            true,
		AmountCents:     amountCents,
		Currency:        "CNY",
		Raw:             body,
	}, nil
}

// QueryOrder queries Alipay for the current state of an order by OutTradeNo.
func (p *alipayProvider) QueryOrder(ctx context.Context, outTradeNo string) (*provider.WebhookEvent, error) {
	param := alipaysdk.TradeQuery{}
	param.OutTradeNo = outTradeNo

	rsp, err := p.client.TradeQuery(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("alipay: TradeQuery: %w", err)
	}

	paid := rsp.TradeStatus == alipaysdk.TradeStatusSuccess ||
		rsp.TradeStatus == alipaysdk.TradeStatusFinished

	amountCents, _ := yuanToCents(rsp.TotalAmount) // best effort; zero on parse error

	return &provider.WebhookEvent{
		OutTradeNo:      rsp.OutTradeNo,
		ProviderOrderID: rsp.TradeNo,
		Paid:            paid,
		AmountCents:     amountCents,
		Currency:        "CNY",
	}, nil
}

// yuanToCents converts a yuan string (e.g. "12.34") to integer cents (1234).
func yuanToCents(yuan string) (int64, error) {
	f, err := strconv.ParseFloat(yuan, 64)
	if err != nil {
		return 0, err
	}
	return int64(f * 100), nil
}
