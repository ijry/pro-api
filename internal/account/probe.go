package account

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ProviderProbe 是单 provider 探测器接口;具体实现见 internal/account/probe/。
type ProviderProbe interface {
	Probe(ctx context.Context, cred AccountCred) (http.Header, error)
}

// ProbeConfig 控制后台定时探测循环。零值字段由 NewProbe 填默认值。
type ProbeConfig struct {
	// Tick 是后台扫描间隔。<=0 时默认 5 分钟。
	Tick time.Duration
	// StaleAfter 是额度"陈旧"判定阈值:quota_synced_at 早于 now-StaleAfter
	// (或为空)的 active 账号会被探测。<=0 时默认 10 分钟。
	StaleAfter time.Duration
	// BatchLimit 是单轮最多探测的账号数。<=0 时默认 100。
	BatchLimit int
	// Concurrency 是单轮内的最大并发探测数。<=0 时默认 8。
	Concurrency int
}

func (c *ProbeConfig) withDefaults() ProbeConfig {
	out := *c
	if out.Tick <= 0 {
		out.Tick = 5 * time.Minute
	}
	if out.StaleAfter <= 0 {
		out.StaleAfter = 10 * time.Minute
	}
	if out.BatchLimit <= 0 {
		out.BatchLimit = 100
	}
	if out.Concurrency <= 0 {
		out.Concurrency = 8
	}
	return out
}

type probeImpl struct {
	repo         Repo
	quotaTracker QuotaTracker
	breaker      Breaker
	providers    map[string]ProviderProbe
	cfg          ProbeConfig
	log          *zap.Logger
	stop         chan struct{}
}

// NewProbe 构造 Probe;providers 由 wire 注入(provider 名 → 实现)。
// breaker 用于后台循环的失败标记(可为 nil:此时后台探测失败只记事件不改状态,
// 手动 ProbeOne 路径不受影响)。cfg 控制后台循环节奏。log 可为 nil。
func NewProbe(r Repo, q QuotaTracker, b Breaker, providers map[string]ProviderProbe, cfg ProbeConfig, log *zap.Logger) Probe {
	return &probeImpl{
		repo:         r,
		quotaTracker: q,
		breaker:      b,
		providers:    providers,
		cfg:          cfg.withDefaults(),
		log:          log,
		stop:         make(chan struct{}),
	}
}

// ProbeOne 探测单个账号:回填额度、追加 probed 事件。不做状态标记。
func (p *probeImpl) ProbeOne(ctx context.Context, a *Account) error {
	pp, ok := p.providers[a.Provider]
	if !ok {
		return fmt.Errorf("probe: unknown provider %q", a.Provider)
	}
	h, err := pp.Probe(ctx, a.Cred)
	if err != nil {
		_ = p.repo.AppendEvent(ctx, a.ID, "probed", map[string]any{"err": err.Error()})
		return err
	}
	snap := p.quotaTracker.ExtractFromResponse(a.Provider, h)
	if snap != nil {
		_ = p.quotaTracker.UpdateAccount(ctx, a.ID, snap)
	}
	_ = p.repo.AppendEvent(ctx, a.ID, "probed", snap)
	return nil
}

// Run 后台循环:周期扫描额度陈旧的 active 账号并探测。
func (p *probeImpl) Run(ctx context.Context) error {
	tick := time.NewTicker(p.cfg.Tick)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-p.stop:
			return nil
		case <-tick.C:
			n, err := p.RunOnce(ctx)
			if err != nil && p.log != nil {
				p.log.Warn("probe loop", zap.Error(err))
			}
			if n > 0 && p.log != nil {
				p.log.Info("probe loop", zap.Int("probed", n))
			}
		}
	}
}

// RunOnce 执行一轮定时探测:取陈旧 active 账号,并发探测,失败按类型标记。
// 返回本轮探测的账号数。供 Run() 调用,也供测试直接驱动。
func (p *probeImpl) RunOnce(ctx context.Context) (int, error) {
	staleBefore := time.Now().UTC().Add(-p.cfg.StaleAfter)
	list, err := p.repo.ListForProbe(ctx, staleBefore, p.cfg.BatchLimit)
	if err != nil {
		return 0, err
	}
	if len(list) == 0 {
		return 0, nil
	}
	sem := make(chan struct{}, p.cfg.Concurrency)
	var wg sync.WaitGroup
	for _, a := range list {
		wg.Add(1)
		sem <- struct{}{}
		go func(a *Account) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := p.probeAndMark(ctx, a); err != nil && p.log != nil {
				p.log.Debug("probe account",
					zap.Int64("account_id", a.ID), zap.Error(err))
			}
		}(a)
	}
	wg.Wait()
	return len(list), nil
}

// probeAndMark 探测单账号并在失败时标记状态(仅后台循环用)。
func (p *probeImpl) probeAndMark(ctx context.Context, a *Account) error {
	pp, ok := p.providers[a.Provider]
	if !ok {
		return fmt.Errorf("probe: unknown provider %q", a.Provider)
	}
	h, err := pp.Probe(ctx, a.Cred)
	if err != nil {
		_ = p.repo.AppendEvent(ctx, a.ID, "probed", map[string]any{"err": err.Error()})
		p.mark(ctx, a.ID, err, h)
		return err
	}
	snap := p.quotaTracker.ExtractFromResponse(a.Provider, h)
	if snap != nil {
		_ = p.quotaTracker.UpdateAccount(ctx, a.ID, snap)
	}
	_ = p.repo.AppendEvent(ctx, a.ID, "probed", snap)
	return nil
}

// mark 复用 selector 的 classifyFailure / parseResetAt,按失败类型打标。
// 与转发链路的 ReportFailure 语义一致,但不走连续失败计数(探测失败不叠加熔断)。
func (p *probeImpl) mark(ctx context.Context, id int64, err error, h http.Header) {
	if p.breaker == nil {
		return
	}
	switch classifyFailure(err) {
	case failRateLimited:
		reset := parseResetAt(h)
		if reset.IsZero() {
			reset = time.Now().Add(5 * time.Minute)
		}
		_ = p.breaker.MarkCooldown(ctx, id, reset, "probe: "+err.Error())
	case failTokenExpired:
		_ = p.breaker.MarkExpired(ctx, id, "probe: "+err.Error())
	case failInvalidCred:
		_ = p.breaker.MarkInvalid(ctx, id, "probe: "+err.Error())
	default:
		// 未知失败(如网络超时)不改状态:可能是探测端暂时性问题,
		// 误标记会把有效账号踢出池子。
	}
}

// Close 关闭后台 Run 循环。
func (p *probeImpl) Close() error {
	select {
	case <-p.stop:
	default:
		close(p.stop)
	}
	return nil
}
