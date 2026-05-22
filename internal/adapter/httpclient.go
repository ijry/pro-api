package adapter

import (
	"net"
	"net/http"
	"time"
)

// ClientConfig 是每家 adapter 的 HTTP client 配置。
type ClientConfig struct {
	Provider            string
	Timeout             time.Duration // 0 = 不设端到端超时(流式必须 0)
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	DisableKeepAlives   bool
}

// NewHTTPClient 构造一个用于 adapter 的 *http.Client。
//
//   - Timeout=0 时只设 dial / response header 超时,适合流式
//   - CheckRedirect=ErrUseLastResponse → 不跟随 3xx(避免凭证泄漏)
//   - 共享 Transport 复用连接池
func NewHTTPClient(cfg ClientConfig) *http.Client {
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 100
	}
	if cfg.MaxIdleConnsPerHost <= 0 {
		cfg.MaxIdleConnsPerHost = 32
	}
	if cfg.IdleConnTimeout <= 0 {
		cfg.IdleConnTimeout = 90 * time.Second
	}
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		DisableKeepAlives:   cfg.DisableKeepAlives,
	}
	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: tr,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
