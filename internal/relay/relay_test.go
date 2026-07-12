package relay_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/account"
	"github.com/ijry/pro-api/internal/adapter"
	"github.com/ijry/pro-api/internal/channel"
	"github.com/ijry/pro-api/internal/protocol/ir"
	"github.com/ijry/pro-api/internal/relay"
	"github.com/stretchr/testify/require"
)

// fakeAdapter 记录最近一次 Chat 收到的 cred,便于断言。
type fakeAdapter struct {
	name     string
	gotCred  adapter.Credential
	resp     *ir.ChatResponse
	errToRet error
	// chunks 非空时,ChatStream 返回一个依次吐这些 chunk、末尾 io.EOF 的 reader。
	chunks []*ir.ChatChunk
}

func (f *fakeAdapter) Name() string                     { return f.name }
func (f *fakeAdapter) Capabilities() adapter.Capability { return adapter.CapChat }
func (f *fakeAdapter) SupportedModels() []string        { return []string{"any"} }
func (f *fakeAdapter) Chat(_ context.Context, _ *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	f.gotCred = cred
	if f.errToRet != nil {
		return nil, f.errToRet
	}
	return f.resp, nil
}
func (f *fakeAdapter) ChatStream(_ context.Context, _ *ir.ChatRequest, cred adapter.Credential) (adapter.StreamReader, error) {
	f.gotCred = cred
	if f.errToRet != nil {
		return nil, f.errToRet
	}
	return &fakeStreamReader{chunks: f.chunks}, nil
}
func (f *fakeAdapter) Embed(_ context.Context, _ *ir.EmbedRequest, _ adapter.Credential) (*ir.EmbedResponse, error) {
	return nil, errors.New("embed not implemented in fake")
}

// fakeStreamReader 依次吐 chunks,读完返回 io.EOF。模拟 OpenAI 系"usage 仅在末尾 chunk"。
type fakeStreamReader struct {
	chunks []*ir.ChatChunk
	i      int
	closed bool
}

func (r *fakeStreamReader) Next(_ context.Context) (*ir.ChatChunk, error) {
	if r.i >= len(r.chunks) {
		return nil, io.EOF
	}
	c := r.chunks[r.i]
	r.i++
	return c, nil
}
func (r *fakeStreamReader) Close() error { r.closed = true; return nil }

// fakeFacade 记录 Select / ReportSuccess / ReportFailure 调用次数。
type fakeFacade struct {
	acc            *account.Account
	selectErr      error
	successCalls   int
	failureCalls   int
	lastReportID   int64
	lastReportErr  error
	usageCalls     int
	lastUsageID    int64
	lastUsageToken int64
}

func (f *fakeFacade) Select(_ context.Context, _ *channel.Channel, _ account.SelectHint) (*account.Account, error) {
	if f.selectErr != nil {
		return nil, f.selectErr
	}
	return f.acc, nil
}
func (f *fakeFacade) ReportSuccess(id int64, _ time.Duration) {
	f.successCalls++
	f.lastReportID = id
}
func (f *fakeFacade) ReportFailure(id int64, err error, _ http.Header) {
	f.failureCalls++
	f.lastReportID = id
	f.lastReportErr = err
}
func (f *fakeFacade) ReportUsage(id int64, tokens int64) {
	f.usageCalls++
	f.lastUsageID = id
	f.lastUsageToken = tokens
}

func newRegistry(t *testing.T, a adapter.Adapter) adapter.Registry {
	t.Helper()
	reg := adapter.NewRegistry()
	reg.Register(a)
	return reg
}

func TestService_Chat_UsesAccountWhenPoolEnabled(t *testing.T) {
	fakeAd := &fakeAdapter{name: "anthropic", resp: &ir.ChatResponse{ID: "r1"}}
	reg := newRegistry(t, fakeAd)
	facade := &fakeFacade{acc: &account.Account{
		ID: 77,
		Cred: account.AccountCred{
			AccessToken: "acct-access-token",
		},
	}}
	s := relay.New(reg, facade)

	ch := &channel.Channel{
		Provider:    "anthropic",
		BaseURL:     "https://api.example",
		Cred:        channel.Credential{APIKey: "channel-direct-key"},
		PoolEnabled: 1,
	}
	_, accID, err := s.Chat(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.NoError(t, err)
	require.Equal(t, int64(77), accID)
	require.Equal(t, "acct-access-token", fakeAd.gotCred.APIKey, "pool branch must override cred with account.Cred")
	require.Equal(t, 1, facade.successCalls)
	require.Equal(t, 0, facade.failureCalls)
}

func TestService_Chat_PoolOverlayPreservesChannelExtra(t *testing.T) {
	// 当 channel 携带 provider 级配置(如 Azure deployment id 在 Extra,或 Region)时,
	// 账号池命中只应覆盖 APIKey/AccessToken,其它字段必须保留。
	fakeAd := &fakeAdapter{name: "anthropic", resp: &ir.ChatResponse{ID: "r"}}
	reg := newRegistry(t, fakeAd)
	facade := &fakeFacade{acc: &account.Account{
		ID:   42,
		Cred: account.AccountCred{AccessToken: "acct-token"},
	}}
	s := relay.New(reg, facade)

	ch := &channel.Channel{
		Provider: "anthropic",
		BaseURL:  "https://api.example",
		Cred: channel.Credential{
			APIKey: "channel-key-should-be-overridden",
			Secret: "channel-secret",
			Region: "us-east-1",
			Extra:  map[string]string{"deployment": "gpt-4"},
		},
		PoolEnabled: 1,
	}
	_, _, err := s.Chat(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.NoError(t, err)
	require.Equal(t, "acct-token", fakeAd.gotCred.APIKey, "account token must override APIKey")
	require.Equal(t, "channel-secret", fakeAd.gotCred.Secret, "channel Secret must survive overlay")
	require.Equal(t, "us-east-1", fakeAd.gotCred.Region, "channel Region must survive overlay")
	require.Equal(t, "gpt-4", fakeAd.gotCred.Extra["deployment"], "channel Extra must survive overlay")
}

func TestService_Chat_PoolDisabledUsesChannelCred(t *testing.T) {
	fakeAd := &fakeAdapter{name: "openai", resp: &ir.ChatResponse{ID: "r2"}}
	reg := newRegistry(t, fakeAd)
	facade := &fakeFacade{acc: &account.Account{ID: 99, Cred: account.AccountCred{APIKey: "should-not-be-used"}}}
	s := relay.New(reg, facade)

	ch := &channel.Channel{
		Provider:    "openai",
		BaseURL:     "https://api.openai.example",
		Cred:        channel.Credential{APIKey: "channel-direct-key", Extra: map[string]string{"x": "1"}},
		PoolEnabled: 0,
	}
	_, accID, err := s.Chat(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.NoError(t, err)
	require.Equal(t, int64(0), accID, "PoolEnabled=0 must not consume an account")
	require.Equal(t, "channel-direct-key", fakeAd.gotCred.APIKey)
	require.Equal(t, "1", fakeAd.gotCred.Extra["x"], "channel.Extra must propagate through ToAdapter")
	require.Equal(t, 0, facade.successCalls, "facade must not be touched when pool disabled")
	require.Equal(t, 0, facade.failureCalls)
}

func TestService_Chat_PoolNilFacadeFallsBackToChannel(t *testing.T) {
	fakeAd := &fakeAdapter{name: "openai", resp: &ir.ChatResponse{ID: "r3"}}
	reg := newRegistry(t, fakeAd)
	s := relay.New(reg, nil)
	ch := &channel.Channel{
		Provider:    "openai",
		Cred:        channel.Credential{APIKey: "channel-key"},
		PoolEnabled: 1,
	}
	_, accID, err := s.Chat(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.NoError(t, err)
	require.Equal(t, int64(0), accID)
	require.Equal(t, "channel-key", fakeAd.gotCred.APIKey)
}

func TestService_Chat_ReportsFailureOnAdapterError(t *testing.T) {
	upstreamErr := errors.New("429 rate limit")
	fakeAd := &fakeAdapter{name: "anthropic", errToRet: upstreamErr}
	reg := newRegistry(t, fakeAd)
	facade := &fakeFacade{acc: &account.Account{ID: 88, Cred: account.AccountCred{APIKey: "k"}}}
	s := relay.New(reg, facade)
	ch := &channel.Channel{Provider: "anthropic", Cred: channel.Credential{}, PoolEnabled: 1}

	_, accID, err := s.Chat(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.ErrorIs(t, err, upstreamErr)
	require.Equal(t, int64(88), accID)
	require.Equal(t, 1, facade.failureCalls)
	require.Equal(t, 0, facade.successCalls)
	require.Equal(t, int64(88), facade.lastReportID)
}

func TestService_Chat_SelectErrorPropagates(t *testing.T) {
	fakeAd := &fakeAdapter{name: "anthropic"}
	reg := newRegistry(t, fakeAd)
	selErr := errors.New("pool empty")
	facade := &fakeFacade{selectErr: selErr}
	s := relay.New(reg, facade)
	ch := &channel.Channel{Provider: "anthropic", PoolEnabled: 1}

	_, _, err := s.Chat(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.ErrorIs(t, err, selErr)
	require.Equal(t, 0, facade.successCalls)
	require.Equal(t, 0, facade.failureCalls, "Select error must not trigger ReportFailure (account not selected)")
}

// TestService_Chat_ReportsUsageOnSuccess 验证非流式调用成功后按 TotalTokens 上报用量
// (供 manual 账号扣减手动额度)。
func TestService_Chat_ReportsUsageOnSuccess(t *testing.T) {
	fakeAd := &fakeAdapter{name: "openai", resp: &ir.ChatResponse{
		ID:    "r",
		Usage: ir.Usage{PromptTokens: 30, CompletionTokens: 70, TotalTokens: 100},
	}}
	reg := newRegistry(t, fakeAd)
	facade := &fakeFacade{acc: &account.Account{ID: 55, Cred: account.AccountCred{APIKey: "k"}}}
	s := relay.New(reg, facade)
	ch := &channel.Channel{Provider: "openai", Cred: channel.Credential{}, PoolEnabled: 1}

	_, accID, err := s.Chat(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.NoError(t, err)
	require.Equal(t, int64(55), accID)
	require.Equal(t, 1, facade.usageCalls)
	require.Equal(t, int64(55), facade.lastUsageID)
	require.Equal(t, int64(100), facade.lastUsageToken)
}

// TestService_Chat_NoUsageReportWhenPoolDisabled 验证未启用账号池时不上报用量。
func TestService_Chat_NoUsageReportWhenPoolDisabled(t *testing.T) {
	fakeAd := &fakeAdapter{name: "openai", resp: &ir.ChatResponse{
		ID:    "r",
		Usage: ir.Usage{TotalTokens: 100},
	}}
	reg := newRegistry(t, fakeAd)
	facade := &fakeFacade{acc: &account.Account{ID: 55}}
	s := relay.New(reg, facade)
	ch := &channel.Channel{Provider: "openai", Cred: channel.Credential{APIKey: "k"}, PoolEnabled: 0}

	_, _, err := s.Chat(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.NoError(t, err)
	require.Equal(t, 0, facade.usageCalls, "pool disabled 不应上报用量")
}

// TestService_ChatStream_ReportsUsageAtEOF 验证流式调用在流末尾(EOF)按末尾 chunk 的
// usage 上报一次用量。usage 仅在最后一个 chunk 出现(OpenAI 系语义)。
func TestService_ChatStream_ReportsUsageAtEOF(t *testing.T) {
	fakeAd := &fakeAdapter{name: "openai", chunks: []*ir.ChatChunk{
		{ID: "c1", Delta: ir.Delta{Content: "hel"}},
		{ID: "c2", Delta: ir.Delta{Content: "lo"}},
		{ID: "c3", Usage: &ir.Usage{PromptTokens: 12, CompletionTokens: 8, TotalTokens: 20}},
	}}
	reg := newRegistry(t, fakeAd)
	facade := &fakeFacade{acc: &account.Account{ID: 66, Cred: account.AccountCred{APIKey: "k"}}}
	s := relay.New(reg, facade)
	ch := &channel.Channel{Provider: "openai", Cred: channel.Credential{}, PoolEnabled: 1}

	reader, accID, err := s.ChatStream(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.NoError(t, err)
	require.Equal(t, int64(66), accID)

	// 未消费完流之前不应上报。
	require.Equal(t, 0, facade.usageCalls)

	ctx := context.Background()
	for {
		_, nerr := reader.Next(ctx)
		if nerr == io.EOF {
			break
		}
		require.NoError(t, nerr)
	}
	require.NoError(t, reader.Close())

	require.Equal(t, 1, facade.usageCalls, "流末尾应恰好上报一次")
	require.Equal(t, int64(66), facade.lastUsageID)
	require.Equal(t, int64(20), facade.lastUsageToken)
}

// TestService_ChatStream_NoUsageReportWhenPoolDisabled 验证未启用账号池时流式不包装、不上报。
func TestService_ChatStream_NoUsageReportWhenPoolDisabled(t *testing.T) {
	fakeAd := &fakeAdapter{name: "openai", chunks: []*ir.ChatChunk{
		{ID: "c1", Usage: &ir.Usage{TotalTokens: 20}},
	}}
	reg := newRegistry(t, fakeAd)
	facade := &fakeFacade{acc: &account.Account{ID: 66}}
	s := relay.New(reg, facade)
	ch := &channel.Channel{Provider: "openai", Cred: channel.Credential{APIKey: "k"}, PoolEnabled: 0}

	reader, accID, err := s.ChatStream(context.Background(), &ir.ChatRequest{Model: "m"}, ch)
	require.NoError(t, err)
	require.Equal(t, int64(0), accID)
	ctx := context.Background()
	for {
		_, nerr := reader.Next(ctx)
		if nerr == io.EOF {
			break
		}
		require.NoError(t, nerr)
	}
	require.NoError(t, reader.Close())
	require.Equal(t, 0, facade.usageCalls)
}
