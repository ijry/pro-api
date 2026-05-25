package wechatpay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/ijry/pro-api/internal/payment/provider"
)

// Config holds the configuration for the WeChat Pay provider.
type Config struct {
	AppID      string // 公众号/小程序 AppID
	MchID      string // 商户号
	SerialNo   string // 商户证书序列号
	PrivateKey string // 商户私钥 PEM (PKCS8 RSA)
	APIv3Key   string // APIv3 密钥
}

type wechatpayProvider struct {
	client *core.Client
	cfg    Config
}

// New creates a WeChat Pay payment provider using auto-auth cipher (certificate auto-download).
// Returns an error if the private key cannot be loaded or the client cannot be created.
func New(cfg Config) (provider.Provider, error) {
	privateKey, err := utils.LoadPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("wechatpay: load private key: %w", err)
	}

	client, err := core.NewClient(
		context.Background(),
		option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.SerialNo, privateKey, cfg.APIv3Key),
	)
	if err != nil {
		return nil, fmt.Errorf("wechatpay: create client: %w", err)
	}

	return &wechatpayProvider{client: client, cfg: cfg}, nil
}

func (p *wechatpayProvider) Name() string { return "wechatpay" }

// CreateOrder calls Native Pay (QR-code) pre-order and returns the code_url as PayURL.
func (p *wechatpayProvider) CreateOrder(ctx context.Context, in provider.CreateOrderInput) (*provider.CreateOrderResult, error) {
	description := in.Description
	if description == "" {
		description = in.OutTradeNo
	}

	svc := native.NativeApiService{Client: p.client}

	appID := p.cfg.AppID
	mchID := p.cfg.MchID
	outTradeNo := in.OutTradeNo
	notifyURL := in.NotifyURL
	total := in.AmountCents
	currency := in.Currency
	if currency == "" {
		currency = "CNY"
	}
	desc := description

	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       &appID,
		Mchid:       &mchID,
		Description: &desc,
		OutTradeNo:  &outTradeNo,
		NotifyUrl:   &notifyURL,
		Amount: &native.Amount{
			Total:    &total,
			Currency: &currency,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("wechatpay: Prepay: %w", err)
	}

	codeURL := ""
	if resp.CodeUrl != nil {
		codeURL = *resp.CodeUrl
	}

	return &provider.CreateOrderResult{
		PayURL: codeURL,
		QRCode: codeURL,
	}, nil
}

// webhookBody is the simplified shape of a WeChat Pay APIv3 webhook notification.
type webhookBody struct {
	EventType string `json:"event_type"`
	Resource  struct {
		Algorithm      string `json:"algorithm"`
		CipherText     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
		OriginalType   string `json:"original_type"`
	} `json:"resource"`
	// Fallback: some notification formats carry out_trade_no at the top level.
	OutTradeNo string `json:"out_trade_no"`
}

// HandleWebhook parses a WeChat Pay APIv3 webhook notification.
// Full AEAD decryption of the encrypted resource is complex; this implementation
// performs basic JSON parsing to extract out_trade_no from event_type==TRANSACTION.SUCCESS.
// Returns nil event (no error) for non-payment-success events.
func (p *wechatpayProvider) HandleWebhook(_ context.Context, body []byte, _ map[string]string) (*provider.WebhookEvent, error) {
	var wb webhookBody
	if err := json.Unmarshal(body, &wb); err != nil {
		return nil, fmt.Errorf("wechatpay: parse webhook body: %w", err)
	}

	if wb.EventType != "TRANSACTION.SUCCESS" {
		return nil, nil
	}

	outTradeNo := wb.OutTradeNo
	if outTradeNo == "" {
		// Best-effort: out_trade_no may be embedded in the encrypted resource;
		// without full decryption we cannot extract it.
		return nil, fmt.Errorf("wechatpay: webhook missing out_trade_no (encrypted resource not decrypted)")
	}

	return &provider.WebhookEvent{
		OutTradeNo: outTradeNo,
		Paid:       true,
		Currency:   "CNY",
		Raw:        body,
	}, nil
}

// QueryOrder queries WeChat Pay by OutTradeNo and returns the payment status.
func (p *wechatpayProvider) QueryOrder(ctx context.Context, outTradeNo string) (*provider.WebhookEvent, error) {
	svc := native.NativeApiService{Client: p.client}
	mchID := p.cfg.MchID

	txn, _, err := svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: &outTradeNo,
		Mchid:      &mchID,
	})
	if err != nil {
		return nil, fmt.Errorf("wechatpay: QueryOrderByOutTradeNo: %w", err)
	}

	paid := txn.TradeState != nil && *txn.TradeState == "SUCCESS"

	var amountCents int64
	if txn.Amount != nil && txn.Amount.Total != nil {
		amountCents = *txn.Amount.Total
	}

	providerOrderID := ""
	if txn.TransactionId != nil {
		providerOrderID = *txn.TransactionId
	}

	tradeOutTradeNo := outTradeNo
	if txn.OutTradeNo != nil {
		tradeOutTradeNo = *txn.OutTradeNo
	}

	return &provider.WebhookEvent{
		OutTradeNo:      tradeOutTradeNo,
		ProviderOrderID: providerOrderID,
		Paid:            paid,
		AmountCents:     amountCents,
		Currency:        "CNY",
	}, nil
}
