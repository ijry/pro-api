package account

import (
	"context"
	"net/http"
	"time"

	"github.com/ijry/pro-api/internal/channel"
)

// QuotaWindow 是单个时间窗的额度信息。
type QuotaWindow struct {
	Total     *int64
	Remaining *int64
	ResetAt   *time.Time
	SyncedAt  *time.Time
}

// QuotaSnapshot 是上游响应解出的额度快照。
type QuotaSnapshot struct {
	Quota5h   QuotaWindow
	QuotaWeek QuotaWindow
}

// SelectHint 是 Selector 入参。
type SelectHint struct {
	Excluded []int64
	UserID   int64
	TokenID  int64
	Model    string
}

// Repo 是数据访问接口。
type Repo interface {
	Create(ctx context.Context, a *Account) error
	Update(ctx context.Context, a *Account) error
	Get(ctx context.Context, id int64) (*Account, error)
	ListByChannel(ctx context.Context, channelID int64) ([]*Account, error)
	ListByShareTag(ctx context.Context, tag string) ([]*Account, error)
	ListForRefresher(ctx context.Context, before time.Time, limit int) ([]*Account, error)
	ListForReaper(ctx context.Context, now time.Time, limit int) ([]*Account, error)
	Delete(ctx context.Context, id int64) error
	AppendEvent(ctx context.Context, accountID int64, eventType string, payload any) error
}

// Selector 从 channel 的账号池中挑一个 account。
type Selector interface {
	Select(ctx context.Context, ch *channel.Channel, hint SelectHint) (*Account, error)
	ReportSuccess(accountID int64, latency time.Duration)
	ReportFailure(accountID int64, err error, upstreamHeaders http.Header)
}

// Breaker 处理短熔断 / 凭证失败 / 后台恢复。
type Breaker interface {
	MarkCooldown(ctx context.Context, accountID int64, until time.Time, reason string) error
	MarkExpired(ctx context.Context, accountID int64, reason string) error
	MarkInvalid(ctx context.Context, accountID int64, reason string) error
	IncConsecFailure(ctx context.Context, accountID int64) (int, error)
	Run(ctx context.Context) error
	Close() error
}

// QuotaTracker 嗅探上游响应头中的额度。
type QuotaTracker interface {
	ExtractFromResponse(provider string, headers http.Header) *QuotaSnapshot
	UpdateAccount(ctx context.Context, accountID int64, snap *QuotaSnapshot) error
}

// Refresher 后台刷 OAuth access_token。
type Refresher interface {
	Run(ctx context.Context) error
	RefreshOne(ctx context.Context, accountID int64) error
}

// Probe 入池探测。
type Probe interface {
	Run(ctx context.Context, account *Account) error
}

// Importer 解析粘贴/文件多种格式。
type Importer interface {
	Detect(payload []byte) (format string, ok bool)
	Parse(payload []byte, format string) ([]*Account, error)
}

// OAuthFlow PKCE 流程(M2 完整,M1 仅 ExchangeRefreshToken)。
type OAuthFlow interface {
	Start(ctx context.Context, provider string, channelID int64) (authURL, state string, err error)
	Callback(ctx context.Context, state, code string) (*Account, error)
	ExchangeRefreshToken(ctx context.Context, provider, refreshToken string) (*AccountCred, error)
}
