package log

import "github.com/prometheus/client_golang/prometheus"

var (
	droppedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proapi_log_dropped_total",
			Help: "Total number of log events dropped due to full channel.",
		}, []string{"kind"},
	)
	writeFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proapi_log_write_failures_total",
			Help: "Total number of log batches that failed to write after retries.",
		}, []string{"kind"},
	)
	flushDurationSec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proapi_log_flush_duration_seconds",
			Help:    "Duration of log batch flush operations.",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
		}, []string{"kind"},
	)
	queueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "proapi_log_queue_depth",
			Help: "Current depth of in-memory log channels.",
		}, []string{"kind"},
	)
	partitionEnsureTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proapi_log_partition_ensure_total",
			Help: "Partition ensure attempts.",
		}, []string{"result"},
	)
)

func init() {
	prometheus.MustRegister(
		droppedTotal,
		writeFailuresTotal,
		flushDurationSec,
		queueDepth,
		partitionEnsureTotal,
	)
}
