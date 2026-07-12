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
	// Reactivate 把账号从 cooldown 恢复到 active,并清掉 cooldown_until / consec_failures。
	// 使用 map-based 更新,绕过 GORM 对零值字段的跳过语义。
	Reactivate(ctx context.Context, id int64) error
	// ResetFailures 把 consec_failures 重置为 0 并刷新 last_success_at / last_used_at。
	// 使用 map-based 更新,确保零值能写入。
	ResetFailures(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*Account, error)
	ListByChannel(ctx context.Context, channelID int64) ([]*Account, error)
	ListForRefresher(ctx context.Context, before time.Time, limit int) ([]*Account, error)
	ListForReaper(ctx context.Context, now time.Time, limit int) ([]*Account, error)
	// ListForProbe 返回额度陈旧的 active 账号:quota_synced_at 为空,或早于 staleBefore。
	// 供后台定时探测调度器周期拉取。
	ListForProbe(ctx context.Context, staleBefore time.Time, limit int) ([]*Account, error)
	// DeductManualQuota 原子扣减 quota_mode='manual' 账号的 quota_5h_remaining
	// (按 token 用量),下限截断到 0。非 manual 账号或 tokens<=0 时不做任何写入。
	// 用 SQL 表达式原地更新,避免读改写竞态。
	DeductManualQuota(ctx context.Context, accountID int64, tokens int64) error
	Delete(ctx context.Context, id int64) error
	AppendEvent(ctx context.Context, accountID int64, eventType string, payload any) error
	// ListEvents 返回某账号的事件,按 id DESC 排序,支持 page/size 分页。
	// 返回的 total 是该账号事件总数(忽略 page/size)。
	ListEvents(ctx context.Context, accountID int64, page, size int) ([]*AccountEvent, int64, error)
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
	// RunReaperOnce 执行一次 reaper:把 cooldown_until 已过期的账号恢复成 active,
	// 返回本轮处理数量。供 Run() 周期调用,也供 wire/tests 直接驱动。
	RunReaperOnce(ctx context.Context) (int, error)
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
	Close() error
}

// Probe 入池探测 + 后台定时探测。
type Probe interface {
	// ProbeOne 探测单个账号:请求上游、回填额度、追加 probed 事件。
	// 不做状态标记(供建号 / OAuth 入池 / 手动 Test 调用,失败仅返回 error)。
	ProbeOne(ctx context.Context, account *Account) error
	// Run 后台循环:周期扫描额度陈旧的 active 账号并探测,失败时按类型标记
	// (429→cooldown / 401→expired / 403→invalid)。
	Run(ctx context.Context) error
	Close() error
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
