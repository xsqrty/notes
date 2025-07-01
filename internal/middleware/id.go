package middleware

import (
	"context"
	"github.com/google/uuid"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"net/http"
)

type rid string

const (
	requestIdKey = rid("request_id")
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.NewV7()
		if err != nil {
			httpio.Error(w, http.StatusInternalServerError, err)
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIdKey, id)))
	})
}

func GetRequestID(r *http.Request) uuid.UUID {
	id := r.Context().Value(requestIdKey)
	if id == nil {
		return uuid.UUID{}
	}

	return id.(uuid.UUID)
}
