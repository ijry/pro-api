// Package relay provides HTTP handlers for the OpenAI-compatible proxy endpoints.
// Routes: POST /v1/chat/completions, POST /v1/completions, POST /v1/embeddings
// M2a adds: POST /v1/messages (Anthropic), POST /v1beta/models/:model/generateContent (Gemini)
package relay

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/billing"
	"github.com/ijry/pro-api/internal/channel"
	ilog "github.com/ijry/pro-api/internal/log"
	"github.com/ijry/pro-api/internal/pricing"
	"github.com/ijry/pro-api/internal/protocol/anthropic"
	"github.com/ijry/pro-api/internal/protocol/gemini"
	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/internal/protocol/openai"
	relaySvc "github.com/ijry/pro-api/internal/relay"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/internal/token"
	"github.com/ijry/pro-api/pkg/apierr"
)

// Deps holds all dependencies for the relay Handler.
// Pricing, Biller, and LogSvc are optional; nil = skip that feature.
type Deps struct {
	Relay   *relaySvc.Service
	Pricing pricing.Pricing  // optional; nil = skip billing
	Biller  billing.Biller   // optional; nil = skip billing
	LogSvc  ilog.Store       // optional; nil = skip logging
}

// Handler handles relay (proxy) endpoints.
type Handler struct {
	deps Deps
}

// New constructs a relay Handler.
func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

// Register attaches OpenAI-format routes to the given router group.
func (h *Handler) Register(g *gin.RouterGroup) {
	g.POST("/chat/completions", h.ChatCompletions)
	g.POST("/completions", h.Completions)
	g.POST("/embeddings", h.Embeddings)
	g.POST("/images/generations", h.Images)
	g.POST("/audio/speech", h.Speech)
	g.POST("/audio/transcriptions", h.Transcriptions)
}

// RegisterAnthropicRoutes attaches Anthropic-format routes.
func (h *Handler) RegisterAnthropicRoutes(g *gin.RouterGroup) {
	g.POST("/messages", h.AnthropicMessages)
}

// RegisterGeminiRoutes attaches Gemini-format routes.
func (h *Handler) RegisterGeminiRoutes(g *gin.RouterGroup) {
	g.POST("/models/:model/generateContent", h.GeminiGenerateContent)
	g.POST("/models/:model/streamGenerateContent", h.GeminiStreamGenerateContent)
}

// channelFromContext 从 gin context 中提取选定的 *channel.Channel。
// 优先取上层中间件注入的 "relay:channel";否则用 "relay:provider" + "relay:credential"
// 合成一个 PoolEnabled=0 的临时 Channel(兼容尚未接入 channel middleware 的链路)。
// Authorization: Bearer ... header 是最后兜底。
func channelFromContext(c *gin.Context) *channel.Channel {
	if v, ok := c.Get("relay:channel"); ok {
		if ch, ok := v.(*channel.Channel); ok && ch != nil {
			return ch
		}
	}

	provider := "openai"
	if v, ok := c.Get("relay:provider"); ok {
		if s, ok := v.(string); ok && s != "" {
			provider = s
		}
	}

	var cred adapter.Credential
	if v, ok := c.Get("relay:credential"); ok {
		if cr, ok := v.(adapter.Credential); ok {
			cred = cr
		}
	}
	if cred.APIKey == "" {
		auth := c.GetHeader("Authorization")
		if len(auth) > 7 && auth[:7] == "Bearer " {
			cred.APIKey = auth[7:]
		}
	}

	return &channel.Channel{
		Provider: provider,
		BaseURL:  cred.BaseURL,
		Cred: channel.Credential{
			APIKey:  cred.APIKey,
			Secret:  cred.Secret,
			Region:  cred.Region,
			BaseURL: cred.BaseURL,
			Extra:   cred.Extra,
		},
	}
}

// ChatCompletions handles POST /v1/chat/completions
func (h *Handler) ChatCompletions(c *gin.Context) {
	req, err := openai.DecodeChat(c.Request.Body, openai.DecodeOptions{AllowUnknownFields: true})
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	ch := channelFromContext(c)

	if req.Stream {
		h.streamChat(c, req, ch)
		return
	}

	ctx := c.Request.Context()
	userID := token.UserIDFromContext(ctx)
	groupID := token.GroupIDFromContext(ctx)
	var tokenID int64
	if tv, ok := token.FromContext(ctx); ok {
		tokenID = tv.ID
	}

	// Pre-deduct
	var reservationID string
	if h.deps.Pricing != nil && h.deps.Biller != nil {
		var br float64
		if v, ok := c.Get(middleware.CtxKeyBillingGroupRatio); ok {
			if f, ok := v.(float64); ok {
				br = f
			}
		}
		estCost := h.deps.Pricing.EstimateMax(ctx, req.Model, pricing.EstimateInput{
			InputTokens:       countInputTokens(req),
			MaxOutTokens:      h.deps.Pricing.DefaultMaxOut(ctx, req.Model),
			Stream:            false,
			BillingGroupRatio: br,
		})
		if estCost > 0 {
			var err2 error
			reservationID, err2 = h.deps.Biller.Reserve(ctx, userID, tokenID, estCost)
			if err2 != nil {
				middleware.SetErr(c, apierr.New(apierr.CodeBalanceInsufficient, "insufficient quota"))
				return
			}
		}
	}

	start := time.Now()
	resp, _, err := h.deps.Relay.Chat(ctx, req, ch)
	latencyMS := int(time.Since(start).Milliseconds())

	if err != nil {
		if reservationID != "" && h.deps.Biller != nil {
			_ = h.deps.Biller.Refund(ctx, reservationID)
		}
		h.writeLog(ctx, c, req.Model, groupID, tokenID, latencyMS, false,
			0, 0, 0, 0, pricing.Ratios{}, http.StatusBadGateway, err.Error())
		middleware.SetErr(c, mapErr(err))
		return
	}

	var actualCost int64
	var ratios pricing.Ratios
	if h.deps.Pricing != nil && resp.Usage.PromptTokens+resp.Usage.CompletionTokens > 0 {
		result := h.deps.Pricing.Compute(ctx, pricing.ComputeInput{
			Model:        req.Model,
			GroupID:      groupID,
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			CachedTokens: resp.Usage.CachedTokens,
		})
		actualCost = result.Quota
		ratios = result.Ratios
		if reservationID != "" && h.deps.Biller != nil {
			_ = h.deps.Biller.Commit(ctx, reservationID, actualCost)
		}
	} else if reservationID != "" && h.deps.Biller != nil {
		_ = h.deps.Biller.Refund(ctx, reservationID)
	}

	h.writeLog(ctx, c, req.Model, groupID, tokenID, latencyMS, false,
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.CachedTokens,
		actualCost, ratios, http.StatusOK, "")
	c.JSON(http.StatusOK, openai.EncodeChat(resp))
}

func (h *Handler) streamChat(c *gin.Context, req *ir.ChatRequest, ch *channel.Channel) {
	ctx := c.Request.Context()
	userID := token.UserIDFromContext(ctx)
	groupID := token.GroupIDFromContext(ctx)
	var tokenID int64
	if tv, ok := token.FromContext(ctx); ok {
		tokenID = tv.ID
	}

	// Pre-deduct
	var reservationID string
	if h.deps.Pricing != nil && h.deps.Biller != nil {
		var br float64
		if v, ok := c.Get(middleware.CtxKeyBillingGroupRatio); ok {
			if f, ok := v.(float64); ok {
				br = f
			}
		}
		estCost := h.deps.Pricing.EstimateMax(ctx, req.Model, pricing.EstimateInput{
			InputTokens:       countInputTokens(req),
			MaxOutTokens:      h.deps.Pricing.DefaultMaxOut(ctx, req.Model),
			Stream:            true,
			BillingGroupRatio: br,
		})
		if estCost > 0 {
			var err2 error
			reservationID, err2 = h.deps.Biller.Reserve(ctx, userID, tokenID, estCost)
			if err2 != nil {
				middleware.SetErr(c, apierr.New(apierr.CodeBalanceInsufficient, "insufficient quota"))
				return
			}
		}
	}

	start := time.Now()
	reader, _, err := h.deps.Relay.ChatStream(ctx, req, ch)
	if err != nil {
		if reservationID != "" && h.deps.Biller != nil {
			_ = h.deps.Biller.Refund(ctx, reservationID)
		}
		middleware.SetErr(c, mapErr(err))
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")

	flusher, hasFlusher := c.Writer.(http.Flusher)
	var flushFn func()
	if hasFlusher {
		flushFn = flusher.Flush
	}
	sw := openai.NewStreamWriter(c.Writer, flushFn)

	var finalUsage *ir.Usage
	for {
		chunk, err := reader.Next(ctx)
		if err != nil {
			if err == io.EOF {
				_ = sw.WriteDone()
				break
			}
			_, _ = fmt.Fprintf(c.Writer, "data: {\"error\":\"stream error: %s\"}\n\n", err.Error())
			if hasFlusher {
				flusher.Flush()
			}
			if reservationID != "" && h.deps.Biller != nil {
				_ = h.deps.Biller.Refund(ctx, reservationID)
			}
			return
		}
		// Capture usage from final chunk (non-nil only on final chunk)
		if chunk.Usage != nil {
			finalUsage = chunk.Usage
		}
		if err := sw.WriteChunk(chunk); err != nil {
			return
		}
	}

	latencyMS := int(time.Since(start).Milliseconds())

	// Commit or refund after streaming completes
	var actualCost int64
	var ratios pricing.Ratios
	if finalUsage != nil && h.deps.Pricing != nil && finalUsage.PromptTokens+finalUsage.CompletionTokens > 0 {
		result := h.deps.Pricing.Compute(ctx, pricing.ComputeInput{
			Model:        req.Model,
			GroupID:      groupID,
			InputTokens:  finalUsage.PromptTokens,
			OutputTokens: finalUsage.CompletionTokens,
			CachedTokens: finalUsage.CachedTokens,
		})
		actualCost = result.Quota
		ratios = result.Ratios
		if reservationID != "" && h.deps.Biller != nil {
			_ = h.deps.Biller.Commit(ctx, reservationID, actualCost)
		}
	} else if reservationID != "" && h.deps.Biller != nil {
		_ = h.deps.Biller.Refund(ctx, reservationID)
	}

	var inTok, outTok, cachedTok int
	if finalUsage != nil {
		inTok = finalUsage.PromptTokens
		outTok = finalUsage.CompletionTokens
		cachedTok = finalUsage.CachedTokens
	}
	h.writeLog(ctx, c, req.Model, groupID, tokenID, latencyMS, true,
		inTok, outTok, cachedTok, actualCost, ratios, http.StatusOK, "")
}

// Completions handles POST /v1/completions (legacy)
func (h *Handler) Completions(c *gin.Context) {
	completionReq, err := openai.DecodeCompletion(c.Request.Body, openai.DecodeOptions{AllowUnknownFields: true})
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	chatReq := completionReq.ToChat()
	ch := channelFromContext(c)

	resp, _, err := h.deps.Relay.Chat(c.Request.Context(), chatReq, ch)
	if err != nil {
		middleware.SetErr(c, mapErr(err))
		return
	}
	c.JSON(http.StatusOK, openai.EncodeChatAsCompletion(resp))
}

// Embeddings handles POST /v1/embeddings
func (h *Handler) Embeddings(c *gin.Context) {
	req, err := openai.DecodeEmbed(c.Request.Body, openai.DecodeOptions{AllowUnknownFields: true})
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	ch := channelFromContext(c)

	resp, _, err := h.deps.Relay.Embed(c.Request.Context(), req, ch)
	if err != nil {
		middleware.SetErr(c, mapErr(err))
		return
	}
	c.JSON(http.StatusOK, openai.EncodeEmbed(resp, req.EncodingFormat))
}

// AnthropicMessages handles POST /v1/messages (Anthropic Messages API inlet)
func (h *Handler) AnthropicMessages(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 4*1024*1024))
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "read body: "+err.Error()))
		return
	}

	req, err := anthropic.DecodeRequest(body)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}

	ch := channelFromContext(c)

	if req.Stream {
		reader, _, relayErr := h.deps.Relay.ChatStream(c.Request.Context(), req, ch)
		if relayErr != nil {
			middleware.SetErr(c, mapErr(relayErr))
			return
		}
		defer reader.Close()

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("X-Accel-Buffering", "no")
		w := c.Writer
		flusher, hasFlusher := w.(http.Flusher)

		var idx int
		for {
			chunk, nextErr := reader.Next(c.Request.Context())
			if nextErr == io.EOF {
				break
			}
			if nextErr != nil {
				break
			}
			isStop := chunk.FinishReason != ""
			events, _ := anthropic.EncodeChunk(chunk, idx, isStop)
			for _, ev := range events {
				_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, ev.Data)
			}
			if hasFlusher {
				flusher.Flush()
			}
			idx++
			if isStop {
				break
			}
		}
		return
	}

	irResp, _, relayErr := h.deps.Relay.Chat(c.Request.Context(), req, ch)
	if relayErr != nil {
		middleware.SetErr(c, mapErr(relayErr))
		return
	}
	b, _ := anthropic.EncodeResponse(irResp)
	c.Data(http.StatusOK, "application/json", b)
}

// GeminiGenerateContent handles POST /v1beta/models/:model/generateContent
func (h *Handler) GeminiGenerateContent(c *gin.Context) {
	model := c.Param("model")
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 4*1024*1024))
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "read body: "+err.Error()))
		return
	}

	req, err := gemini.DecodeRequest(body, model)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}

	ch := channelFromContext(c)

	irResp, _, relayErr := h.deps.Relay.Chat(c.Request.Context(), req, ch)
	if relayErr != nil {
		middleware.SetErr(c, mapErr(relayErr))
		return
	}
	b, _ := gemini.EncodeResponse(irResp)
	c.Data(http.StatusOK, "application/json", b)
}

// GeminiStreamGenerateContent handles POST /v1beta/models/:model/streamGenerateContent
func (h *Handler) GeminiStreamGenerateContent(c *gin.Context) {
	model := c.Param("model")
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 4*1024*1024))
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "read body: "+err.Error()))
		return
	}

	req, err := gemini.DecodeRequest(body, model)
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	req.Stream = true

	ch := channelFromContext(c)

	reader, _, relayErr := h.deps.Relay.ChatStream(c.Request.Context(), req, ch)
	if relayErr != nil {
		middleware.SetErr(c, mapErr(relayErr))
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	w := c.Writer
	flusher, hasFlusher := w.(http.Flusher)

	for {
		chunk, nextErr := reader.Next(c.Request.Context())
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			break
		}
		isStop := chunk.FinishReason != ""
		events, _ := gemini.EncodeChunk(chunk, isStop)
		for _, ev := range events {
			_, _ = fmt.Fprintf(w, "data: %s\n\n", ev.Data)
		}
		if hasFlusher {
			flusher.Flush()
		}
		if isStop {
			break
		}
	}
}

// mapErr maps errors to apierr.Error.
func mapErr(err error) *apierr.Error {
	if e, ok := err.(*apierr.Error); ok {
		return e
	}
	return apierr.New(apierr.CodeUpstreamError, err.Error())
}

// Images handles POST /v1/images/generations
func (h *Handler) Images(c *gin.Context) {
	var req ir.ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	ch := channelFromContext(c)
	resp, _, err := h.deps.Relay.GenerateImage(c.Request.Context(), &req, ch)
	if err != nil {
		middleware.SetErr(c, mapErr(err))
		return
	}
	type imageItem struct {
		URL           string `json:"url,omitempty"`
		B64JSON       string `json:"b64_json,omitempty"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	}
	type imageResp struct {
		Created int64       `json:"created"`
		Data    []imageItem `json:"data"`
	}
	out := imageResp{Created: resp.Created}
	for _, d := range resp.Data {
		out.Data = append(out.Data, imageItem{
			URL:           d.URL,
			B64JSON:       d.B64JSON,
			RevisedPrompt: d.RevisedPrompt,
		})
	}
	c.JSON(http.StatusOK, out)
}

// Speech handles POST /v1/audio/speech
func (h *Handler) Speech(c *gin.Context) {
	var req ir.SpeechRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, err.Error()))
		return
	}
	ch := channelFromContext(c)
	resp, _, err := h.deps.Relay.TextToSpeech(c.Request.Context(), &req, ch)
	if err != nil {
		middleware.SetErr(c, mapErr(err))
		return
	}
	ct := resp.ContentType
	if ct == "" {
		ct = "audio/mpeg"
	}
	c.Data(http.StatusOK, ct, resp.Data)
}

// Transcriptions handles POST /v1/audio/transcriptions
func (h *Handler) Transcriptions(c *gin.Context) {
	audioFile, header, err := c.Request.FormFile("file")
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "missing file: "+err.Error()))
		return
	}
	defer audioFile.Close()

	audio, err := io.ReadAll(io.LimitReader(audioFile, 25*1024*1024))
	if err != nil {
		middleware.SetErr(c, apierr.New(apierr.CodeInvalidParam, "read file: "+err.Error()))
		return
	}

	req := ir.TranscribeRequest{
		Model:          c.PostForm("model"),
		Audio:          audio,
		Filename:       header.Filename,
		Language:       c.PostForm("language"),
		Prompt:         c.PostForm("prompt"),
		ResponseFormat: c.PostForm("response_format"),
	}
	if req.Model == "" {
		req.Model = "whisper-1"
	}

	ch := channelFromContext(c)
	resp, _, err := h.deps.Relay.Transcribe(c.Request.Context(), &req, ch)
	if err != nil {
		middleware.SetErr(c, mapErr(err))
		return
	}

	if req.ResponseFormat == "text" {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(resp.Text))
		return
	}
	c.JSON(http.StatusOK, gin.H{"text": resp.Text})
}

// countInputTokens approximates the number of input tokens in a chat request.
func countInputTokens(req *ir.ChatRequest) int {
	n := 0
	for _, m := range req.Messages {
		for _, p := range m.Content {
			if p.Text != "" {
				n += len(p.Text) / 4
			}
		}
	}
	if n < 10 {
		n = 10
	}
	return n
}

// writeLog writes an API request log entry via LogSvc (no-op if LogSvc is nil).
func (h *Handler) writeLog(ctx context.Context, c *gin.Context, model string,
	groupID, tokenID int64, latencyMS int, stream bool,
	in, out, cached int, cost int64, ratios pricing.Ratios,
	statusCode int, errMsg string,
) {
	if h.deps.LogSvc == nil {
		return
	}
	userID := token.UserIDFromContext(ctx)
	gid := groupID
	e := ilog.Event{
		UserID:             userID,
		TokenID:            tokenID,
		GroupID:            &gid,
		ClientModel:        model,
		Protocol:           "openai",
		Endpoint:           c.FullPath(),
		IP:                 c.ClientIP(),
		UserAgent:          c.Request.UserAgent(),
		StatusCode:         statusCode,
		LatencyMS:          latencyMS,
		Stream:             stream,
		InputTokens:        in,
		OutputTokens:       out,
		CachedTokens:       cached,
		TotalQuota:         cost,
		BillingInputRatio:  ratios.Input,
		BillingOutputRatio: ratios.Output,
		BillingGroupRatio:  ratios.Group,
		ErrorMsg:           errMsg,
	}
	h.deps.LogSvc.Write(ctx, e)
}
