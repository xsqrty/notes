package prometheus

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xsqrty/notes/internal/config"
)

// NewPrometheusServer initializes and returns an HTTP server configured for exposing Prometheus metrics.
func NewPrometheusServer(cfg config.MetricsConfig) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	router := chi.NewRouter()

	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 10,
	}
}
