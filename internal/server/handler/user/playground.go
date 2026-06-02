// Package user 集中用户侧 HTTP handler。
package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/ijry/pro-api/internal/protocol/openai"
	"github.com/ijry/pro-api/internal/relay"
	"github.com/ijry/pro-api/internal/setting"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
	"go.uber.org/zap"
)

// PlaygroundDeps holds dependencies for the Playground handler.
type PlaygroundDeps struct {
	Relay   *relay.Service
	Channel *channel.Facade
	Setting setting.Store
	Log     *zap.Logger
}

// PlaygroundHandler 是 Playground HTTP handler。
type PlaygroundHandler struct{ deps PlaygroundDeps }

// NewPlaygroundHandler 构造。
func NewPlaygroundHandler(deps PlaygroundDeps) *PlaygroundHandler {
	return &PlaygroundHandler{deps: deps}
}

// Chat POST /api/user/playground/chat
//
// 请求体与 OpenAI /v1/chat/completions 相同。
// 使用 setting playground.default_channel_id 指向的渠道。
func (h *PlaygroundHandler) Chat(c *gin.Context) {
	req, err := openai.DecodeChat(c.Request.Body, openai.DecodeOptions{AllowUnknownFields: true})
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}

	// 读取默认渠道 ID
	channelID := int64(h.deps.Setting.GetInt(c.Request.Context(), "playground.default_channel_id", 0))
	if channelID <= 0 {
		middleware.SetErr(c, apierr.New(apierr.CodeInternal, "playground: default channel not configured (playground.default_channel_id)"))
		return
	}

	// 从渠道缓存取渠道（含解密后的 Cred）
	ch, ok := h.deps.Channel.Cache.GetByID(channelID)
	if !ok {
		middleware.SetErr(c, apierr.New(apierr.CodeInternal, "playground: default channel not found or decryption failed"))
		return
	}

	// 非流式调用(relay 内部根据 ch.PoolEnabled 自动走号池或直调)
	resp, _, err := h.deps.Relay.Chat(c.Request.Context(), req, ch)
	if err != nil {
		var ae *apierr.Error
		if errors.As(err, &ae) {
			middleware.SetErr(c, ae)
		} else {
			middleware.SetErr(c, apierr.New(apierr.CodeUpstreamError, err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, openai.EncodeChat(resp))
}
