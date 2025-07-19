package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/xsqrty/notes/internal/logger"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
)

// lk is a custom string type used to define specific keys for context or other application-specific purposes.
type lk string

// loggerKey is a context key used to store and retrieve the logger instance associated with an HTTP request.
const (
	loggerKey = lk("logger")
)

// Logger is an HTTP middleware that logs request and response details, including method, URL, status, and duration.
func Logger(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.With().Str("req_id", GetRequestID(r).String()).Logger()
			res := httpio.NewResponseDecorator(w)
			start := time.Now()
			next.ServeHTTP(res, r.WithContext(context.WithValue(r.Context(), loggerKey, &logger)))

			logger.Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Int("status", res.StatusCode()).
				Int("bytes", res.BytesWritten()).
				Str("completed_at", time.Since(start).String()).
				Msg(fmt.Sprintf("HTTP %s %s %d", r.Method, r.URL, res.StatusCode()))
		})
	}
}

// Log retrieves the logger from the request's context associated with the loggerKey and returns it as a zerolog.Logger.
func Log(r *http.Request) *zerolog.Logger {
	return r.Context().Value(loggerKey).(*zerolog.Logger)
}
