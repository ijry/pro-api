package relay_test

import (
	"context"
	"errors"
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
}

func (f *fakeAdapter) Name() string                   { return f.name }
func (f *fakeAdapter) Capabilities() adapter.Capability { return adapter.CapChat }
func (f *fakeAdapter) SupportedModels() []string       { return []string{"any"} }
func (f *fakeAdapter) Chat(_ context.Context, _ *ir.ChatRequest, cred adapter.Credential) (*ir.ChatResponse, error) {
	f.gotCred = cred
	if f.errToRet != nil {
		return nil, f.errToRet
	}
	return f.resp, nil
}
func (f *fakeAdapter) ChatStream(_ context.Context, _ *ir.ChatRequest, _ adapter.Credential) (adapter.StreamReader, error) {
	return nil, errors.New("stream not implemented in fake")
}
func (f *fakeAdapter) Embed(_ context.Context, _ *ir.EmbedRequest, _ adapter.Credential) (*ir.EmbedResponse, error) {
	return nil, errors.New("embed not implemented in fake")
}

// fakeFacade 记录 Select / ReportSuccess / ReportFailure 调用次数。
type fakeFacade struct {
	acc           *account.Account
	selectErr     error
	successCalls  int
	failureCalls  int
	lastReportID  int64
	lastReportErr error
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
