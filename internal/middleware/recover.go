package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
)

// Recover is a middleware that captures and handles panics during request processing, logging errors and returning HTTP 500.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(rec)
				}

				Log(r).Error().Err(errors.New(fmt.Sprint(rec))).Bytes("stack", debug.Stack()).Msg("Recovered")
				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
