// Package metrics 集中 Prometheus 指标定义与注册。
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry 是应用全局 metrics 注册表。
var Registry = prometheus.NewRegistry()

// HTTPHandler 返回 /metrics 的 http.Handler。
func HTTPHandler() http.Handler {
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{Registry: Registry})
}

// Init 注册 Go 进程/runtime 默认采集器。
func Init() {
	Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}
