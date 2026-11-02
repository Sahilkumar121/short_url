package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxIDKey string

const (
	RequestIDKey    ctxIDKey = "request_id"
	RequestIDHeader string   = "X-Request-ID"
)

// Request id middleware
func RequestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqID := r.Header.Get(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.New().String()
		}

		w.Header().Add(RequestIDHeader, reqID)
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}

	return ""
}
