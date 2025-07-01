package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/xsqrty/notes/internal/config"
)

type HttpMetrics struct {
	RequestsDuration *prometheus.HistogramVec
	RequestsTotal    *prometheus.CounterVec
}

func NewHttpMetrics(cfg config.MetricsConfig) *HttpMetrics {
	httpMetrics := &HttpMetrics{}
	httpMetrics.RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests",
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
	}, []string{"url", "method", "statusCode"})

	httpMetrics.RequestsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "http_request_duration_seconds",
		Help:      "Duration of HTTP requests",
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
	}, []string{"url", "method", "statusCode"})

	return httpMetrics
}
