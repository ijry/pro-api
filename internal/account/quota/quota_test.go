package quota_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ijry/pro-api/internal/account/quota"
	"github.com/stretchr/testify/require"
)

func TestTracker_DispatchByProvider(t *testing.T) {
	tk := quota.NewTracker(nil) // repo 暂传 nil,只验 Extract 分发
	h := http.Header{}
	h.Set("anthropic-ratelimit-tokens-remaining", "5000")
	snap := tk.ExtractFromResponse("anthropic", h)
	require.NotNil(t, snap)
	require.Nil(t, tk.ExtractFromResponse("unknown", h))
	_ = context.Background()
}
