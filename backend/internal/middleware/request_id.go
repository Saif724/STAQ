package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string
const (
	requestIDKey contextKey = "request_id"
	RequestIDHeader = "X-Request-ID"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)

		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set(RequestIDHeader, requestID)

		ctx := context.WithValue(
			r.Context(),
			requestIDKey,
			requestID,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}

	return requestID
}