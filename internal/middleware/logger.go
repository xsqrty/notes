package middleware

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/xsqrty/notes/internal/logger"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"net/http"
	"time"
)

type lk string

const (
	loggerKey = lk("logger")
)

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

func Log(r *http.Request) *zerolog.Logger {
	return r.Context().Value(loggerKey).(*zerolog.Logger)
}
