// Package relay provides HTTP handlers for the OpenAI-compatible proxy endpoints.
// Routes: POST /v1/chat/completions, POST /v1/completions, POST /v1/embeddings
// M2a adds: POST /v1/messages (Anthropic), POST /v1beta/models/:model/generateContent (Gemini)
package relay

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/ijry/pro-api/internal/protocol/anthropic"
	"github.com/ijry/pro-api/internal/protocol/gemini"
	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/internal/protocol/openai"
	relaySvc "github.com/ijry/pro-api/internal/relay"
	"github.com/ijry/pro-api/internal/server/middleware"
	"github.com/ijry/pro-api/pkg/apierr"
)

// Handler handles relay (proxy) endpoints.
type Handler struct {
	relay *relaySvc.Service
}

// New constructs a relay Handler.
func New(r *relaySvc.Service) *Handler {
	return &Handler{relay: r}
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

	resp, _, err := h.relay.Chat(c.Request.Context(), req, ch)
	if err != nil {
		middleware.SetErr(c, mapErr(err))
		return
	}
	c.JSON(http.StatusOK, openai.EncodeChat(resp))
}

func (h *Handler) streamChat(c *gin.Context, req *ir.ChatRequest, ch *channel.Channel) {
	reader, _, err := h.relay.ChatStream(c.Request.Context(), req, ch)
	if err != nil {
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

	for {
		chunk, err := reader.Next(c.Request.Context())
		if err != nil {
			if err == io.EOF {
				_ = sw.WriteDone()
				return
			}
			_, _ = fmt.Fprintf(c.Writer, "data: {\"error\":\"stream error: %s\"}\n\n", err.Error())
			if hasFlusher {
				flusher.Flush()
			}
			return
		}
		if err := sw.WriteChunk(chunk); err != nil {
			return
		}
	}
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

	resp, _, err := h.relay.Chat(c.Request.Context(), chatReq, ch)
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

	resp, _, err := h.relay.Embed(c.Request.Context(), req, ch)
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
		reader, _, relayErr := h.relay.ChatStream(c.Request.Context(), req, ch)
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

	irResp, _, relayErr := h.relay.Chat(c.Request.Context(), req, ch)
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

	irResp, _, relayErr := h.relay.Chat(c.Request.Context(), req, ch)
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

	reader, _, relayErr := h.relay.ChatStream(c.Request.Context(), req, ch)
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
	resp, _, err := h.relay.GenerateImage(c.Request.Context(), &req, ch)
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
	resp, _, err := h.relay.TextToSpeech(c.Request.Context(), &req, ch)
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
	resp, _, err := h.relay.Transcribe(c.Request.Context(), &req, ch)
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
