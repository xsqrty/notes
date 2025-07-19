package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/xsqrty/notes/internal/metrics"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
)

// Metrics creates a middleware function for collecting HTTP request metrics.
// It tracks total requests and their durations using the provided HttpMetrics instance.
func Metrics(metrics *metrics.HttpMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res := httpio.NewResponseDecorator(w)
			start := time.Now()
			defer func() {
				statusCode := strconv.Itoa(res.StatusCode())
				metrics.RequestsTotal.WithLabelValues(r.URL.String(), r.Method, statusCode).Inc()
				metrics.RequestsDuration.WithLabelValues(r.URL.String(), r.Method, statusCode).
					Observe(time.Since(start).Seconds())
			}()

			next.ServeHTTP(res, r)
		})
	}
}
