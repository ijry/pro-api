// Package relay 提供统一的 LLM 调用编排层。
//
// M2b 起 Service 直接接收 *channel.Channel(而不是 cred+provider),
// 当 ch.PoolEnabled == 1 且 accountFacade 非 nil 时,从账号池选号、上报成功/失败;
// 否则走 channel.Cred 直调(原 M1 行为)。
package relay

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/pkg/apierr"
)

// AccountFacade 是 account 模块向 relay 暴露的最小接口,避免循环 import。
// 实现见 internal/account/facade.go。
type AccountFacade interface {
	Select(ctx context.Context, ch *channel.Channel, hint account.SelectHint) (*account.Account, error)
	ReportSuccess(accountID int64, latency time.Duration)
	ReportFailure(accountID int64, err error, headers http.Header)
	// ReportUsage 按 token 用量扣减账号手动额度(仅 quota_mode='manual' 生效)。
	ReportUsage(accountID int64, tokens int64)
}

// Service 是 relay 编排服务。
type Service struct {
	registry      adapter.Registry
	accountFacade AccountFacade // 可为 nil(不启用账号池)
}

// New 构造 relay.Service。accountFacade 可为 nil。
func New(reg adapter.Registry, accountFacade AccountFacade) *Service {
	return &Service{registry: reg, accountFacade: accountFacade}
}

// resolveCred 决定本次调用使用的凭证。返回 (cred, accountID, err)。
// 当 channel 启用账号池时,从池中选号并把账号凭证 *叠加* 到 channel 凭证之上;
// 否则直接用 channel 凭证。
//
// 叠加而不是替换的原因:channel.Cred 可能携带账号无关的 provider 级配置
// (Azure deployment id 在 Extra、Region、Secret 等),账号池只接管"谁来发请求"
// 的部分(APIKey / AccessToken)。
func (s *Service) resolveCred(ctx context.Context, ch *channel.Channel, hint account.SelectHint) (adapter.Credential, int64, error) {
	cred := ch.Cred.ToAdapter()
	if cred.BaseURL == "" {
		cred.BaseURL = ch.BaseURL
	}
	if ch.PoolEnabled != 1 || s.accountFacade == nil {
		return cred, 0, nil
	}
	a, err := s.accountFacade.Select(ctx, ch, hint)
	if err != nil {
		return adapter.Credential{}, 0, err
	}
	// 账号池命中:用账号的访问凭证覆盖 APIKey/AccessToken,其余字段保留。
	// APIKey 优先,其次 AccessToken(OAuth 账号典型只有 AccessToken)。
	if a.Cred.APIKey != "" {
		cred.APIKey = a.Cred.APIKey
	} else if a.Cred.AccessToken != "" {
		cred.APIKey = a.Cred.AccessToken
	}
	return cred, a.ID, nil
}

// report 在调用结束后给 facade 上报本次结果(成功/失败)。
// accID == 0 表示未启用账号池,直接跳过。
func (s *Service) report(accID int64, err error, latency time.Duration, headers http.Header) {
	if accID == 0 || s.accountFacade == nil {
		return
	}
	if err != nil {
		s.accountFacade.ReportFailure(accID, err, headers)
		return
	}
	s.accountFacade.ReportSuccess(accID, latency)
}

// reportUsage 在调用成功后上报 token 用量(用于 manual 模式账号的手动额度扣减)。
// accID==0(未启用账号池)或 tokens<=0 时为 no-op。
func (s *Service) reportUsage(accID int64, tokens int) {
	if accID == 0 || tokens <= 0 || s.accountFacade == nil {
		return
	}
	s.accountFacade.ReportUsage(accID, int64(tokens))
}

// Chat 执行非流式 chat completion。
func (s *Service) Chat(ctx context.Context, req *ir.ChatRequest, ch *channel.Channel) (*ir.ChatResponse, int64, error) {
	cred, accID, err := s.resolveCred(ctx, ch, account.SelectHint{Model: req.Model})
	if err != nil {
		return nil, 0, err
	}
	a, ok := s.registry.Get(ch.Provider)
	if !ok {
		return nil, accID, fmt.Errorf("relay: unknown provider %q", ch.Provider)
	}
	start := time.Now()
	resp, err := a.Chat(ctx, req, cred)
	s.report(accID, err, time.Since(start), nil)
	if err == nil && resp != nil {
		s.reportUsage(accID, resp.Usage.TotalTokens)
	}
	return resp, accID, err
}

// ChatStream 执行流式 chat completion。
// 注意:reader 构造成功 ≠ 调用成功(stream 内可能 401/429/超时)。本方法只在
// 构造失败时调 facade.ReportFailure;成功时不调 ReportSuccess,避免误把"刚发
// 起请求"上报成"成功",欺骗 selector 给出错分增信。后续把流内结果回传给
// facade 需要包装 StreamReader,留待 M2b+。
func (s *Service) ChatStream(ctx context.Context, req *ir.ChatRequest, ch *channel.Channel) (adapter.StreamReader, int64, error) {
	cred, accID, err := s.resolveCred(ctx, ch, account.SelectHint{Model: req.Model})
	if err != nil {
		return nil, 0, err
	}
	a, ok := s.registry.Get(ch.Provider)
	if !ok {
		return nil, accID, fmt.Errorf("relay: unknown provider %q", ch.Provider)
	}
	reader, err := a.ChatStream(ctx, req, cred)
	if err != nil {
		s.report(accID, err, 0, nil)
		return nil, accID, err
	}
	// 账号池命中时包装 reader:在流末尾(EOF)按累计 usage 扣减 manual 额度。
	// 未启用账号池(accID==0)直接透传,零开销。
	if accID != 0 {
		return &usageDeductReader{inner: reader, svc: s, accID: accID}, accID, nil
	}
	return reader, accID, nil
}

// usageDeductReader 包装上游 StreamReader,在流正常结束时按最终 usage 扣减
// manual 模式账号的手动额度。它对每个 chunk 记录最后一次非空 Usage(OpenAI 系
// 仅末尾 chunk 带 usage),在 Next 返回 io.EOF 时上报一次;非 EOF 错误不扣减。
// 只上报一次由 reported 保证(Close 与 EOF 竞态下不重复扣)。
type usageDeductReader struct {
	inner    adapter.StreamReader
	svc      *Service
	accID    int64
	lastTok  int
	reported bool
}

func (r *usageDeductReader) Next(ctx context.Context) (*ir.ChatChunk, error) {
	chunk, err := r.inner.Next(ctx)
	if err == io.EOF {
		r.flush()
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if chunk != nil && chunk.Usage != nil && chunk.Usage.TotalTokens > 0 {
		r.lastTok = chunk.Usage.TotalTokens
	}
	return chunk, nil
}

// flush 上报累计 usage,幂等(仅第一次生效)。
func (r *usageDeductReader) flush() {
	if r.reported {
		return
	}
	r.reported = true
	r.svc.reportUsage(r.accID, r.lastTok)
}

func (r *usageDeductReader) Close() error { return r.inner.Close() }

// Embed 执行 embedding 请求。
func (s *Service) Embed(ctx context.Context, req *ir.EmbedRequest, ch *channel.Channel) (*ir.EmbedResponse, int64, error) {
	cred, accID, err := s.resolveCred(ctx, ch, account.SelectHint{Model: req.Model})
	if err != nil {
		return nil, 0, err
	}
	a, ok := s.registry.Get(ch.Provider)
	if !ok {
		return nil, accID, fmt.Errorf("relay: unknown provider %q", ch.Provider)
	}
	start := time.Now()
	resp, err := a.Embed(ctx, req, cred)
	s.report(accID, err, time.Since(start), nil)
	if err == nil && resp != nil {
		s.reportUsage(accID, resp.Usage.TotalTokens)
	}
	return resp, accID, err
}

// GenerateImage 执行图片生成请求。
func (s *Service) GenerateImage(ctx context.Context, req *ir.ImageRequest, ch *channel.Channel) (*ir.ImageResponse, int64, error) {
	cred, accID, err := s.resolveCred(ctx, ch, account.SelectHint{Model: req.Model})
	if err != nil {
		return nil, 0, err
	}
	a, ok := s.registry.Get(ch.Provider)
	if !ok {
		return nil, accID, fmt.Errorf("relay: unknown provider %q", ch.Provider)
	}
	ia, ok := a.(adapter.ImageAdapter)
	if !ok {
		return nil, accID, apierr.New(apierr.CodeInvalidParam, "provider does not support image generation")
	}
	start := time.Now()
	resp, err := ia.GenerateImage(ctx, req, cred)
	s.report(accID, err, time.Since(start), nil)
	return resp, accID, err
}

// TextToSpeech 执行文字转语音请求。
func (s *Service) TextToSpeech(ctx context.Context, req *ir.SpeechRequest, ch *channel.Channel) (*ir.SpeechResponse, int64, error) {
	cred, accID, err := s.resolveCred(ctx, ch, account.SelectHint{Model: req.Model})
	if err != nil {
		return nil, 0, err
	}
	a, ok := s.registry.Get(ch.Provider)
	if !ok {
		return nil, accID, fmt.Errorf("relay: unknown provider %q", ch.Provider)
	}
	sa, ok := a.(adapter.SpeechAdapter)
	if !ok {
		return nil, accID, apierr.New(apierr.CodeInvalidParam, "provider does not support TTS")
	}
	start := time.Now()
	resp, err := sa.TextToSpeech(ctx, req, cred)
	s.report(accID, err, time.Since(start), nil)
	return resp, accID, err
}

// Transcribe 执行语音转文字请求。
func (s *Service) Transcribe(ctx context.Context, req *ir.TranscribeRequest, ch *channel.Channel) (*ir.TranscribeResponse, int64, error) {
	cred, accID, err := s.resolveCred(ctx, ch, account.SelectHint{Model: req.Model})
	if err != nil {
		return nil, 0, err
	}
	a, ok := s.registry.Get(ch.Provider)
	if !ok {
		return nil, accID, fmt.Errorf("relay: unknown provider %q", ch.Provider)
	}
	ta, ok := a.(adapter.TranscribeAdapter)
	if !ok {
		return nil, accID, apierr.New(apierr.CodeInvalidParam, "provider does not support transcription")
	}
	start := time.Now()
	resp, err := ta.Transcribe(ctx, req, cred)
	s.report(accID, err, time.Since(start), nil)
	return resp, accID, err
}

// Rerank 执行文档重排请求。
func (s *Service) Rerank(ctx context.Context, req *ir.RerankRequest, ch *channel.Channel) (*ir.RerankResponse, int64, error) {
	cred, accID, err := s.resolveCred(ctx, ch, account.SelectHint{Model: req.Model})
	if err != nil {
		return nil, 0, err
	}
	a, ok := s.registry.Get(ch.Provider)
	if !ok {
		return nil, accID, fmt.Errorf("relay: unknown provider %q", ch.Provider)
	}
	ra, ok := a.(adapter.RerankAdapter)
	if !ok {
		return nil, accID, apierr.New(apierr.CodeInvalidParam, "provider does not support rerank")
	}
	start := time.Now()
	resp, err := ra.Rerank(ctx, req, cred)
	s.report(accID, err, time.Since(start), nil)
	return resp, accID, err
}
