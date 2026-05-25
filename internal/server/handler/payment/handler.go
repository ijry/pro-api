// Package payment provides HTTP handlers for online payment endpoints.
// Routes:
//   POST /api/user/payment/create      — create order (requires session auth)
//   GET  /api/user/payment/orders      — list orders  (requires session auth)
//   POST /api/pay/webhook/:provider    — receive provider webhook (no auth)
package payment

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/payment/online"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

const maxWebhookBodyBytes = 1 * 1024 * 1024 // 1 MiB

// Deps holds handler dependencies.
type Deps struct {
	Online *online.Service
	Log    *zap.Logger
}

// Handler handles payment HTTP endpoints.
type Handler struct{ deps Deps }

// New creates a payment Handler.
func New(deps Deps) *Handler { return &Handler{deps: deps} }

type createOrderReq struct {
	Provider    string `json:"provider" binding:"required"`
	AmountCents int64  `json:"amount_cents" binding:"required,min=1"`
	Credits     int64  `json:"credits" binding:"required,min=1"`
	Currency    string `json:"currency"`
	NotifyURL   string `json:"notify_url"`
	ReturnURL   string `json:"return_url"`
}

// CreateOrder POST /api/user/payment/create
func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	userID := middleware.UserID(c)
	res, err := h.deps.Online.CreateOrder(c.Request.Context(), online.CreateOrderInput{
		UserID:      userID,
		Provider:    req.Provider,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		Credits:     req.Credits,
		NotifyURL:   req.NotifyURL,
		ReturnURL:   req.ReturnURL,
		Description: "账户充值",
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// ListOrders GET /api/user/payment/orders
func (h *Handler) ListOrders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"orders": []any{}})
}

// Webhook POST /api/pay/webhook/:provider
func (h *Handler) Webhook(c *gin.Context) {
	providerName := c.Param("provider")
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBodyBytes))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	if err := h.deps.Online.HandleWebhook(c.Request.Context(), providerName, body, headers); err != nil {
		h.deps.Log.Warn("payment webhook error",
			zap.String("provider", providerName), zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusOK)
}
