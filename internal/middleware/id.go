package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
)

// rid represents a type alias for string, typically used for identification such as request identifiers.
type rid string

// requestIdKey is a constant used as a context key to store and retrieve the unique request ID for HTTP requests.
const (
	requestIdKey = rid("request_id")
)

// RequestID is middleware that assigns a unique request ID to each HTTP request and adds it to the request context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.NewV7()
		if err != nil {
			httpio.Error(w, http.StatusInternalServerError, err)
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIdKey, id)))
	})
}

// GetRequestID extracts the request ID of type uuid.UUID from the request's context using the predefined requestIdKey.
// If no request ID is found, it returns an empty uuid.UUID.
func GetRequestID(r *http.Request) uuid.UUID {
	id := r.Context().Value(requestIdKey)
	if id == nil {
		return uuid.UUID{}
	}

	return id.(uuid.UUID)
}
