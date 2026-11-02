package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Sahilkumar121/short_url/internal/auth"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		// strings.Fields splits the data into slice for (consecutive whitespace, \t, \n)
		parts := strings.Fields(authHeader)

		// strings.EqualFold is a built-in Go (Golang) function used for case-insensitive string comparison
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// validate token string
		claims, err := auth.ValidateAccessToken(tokenString, secret)
		if err != nil || claims == nil {
			http.Error(w, "invalid or expire token", http.StatusUnauthorized)
			return
		}

		if claims.Subject == "" {
			http.Error(w, "invalid token claim", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func OptionalAuthMiddleware(secret []byte, next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		// if no authHeader then it is guest
		if strings.TrimSpace(authHeader) == "" {
			next.ServeHTTP(w, r)
			return
		}

		AuthMiddleware(secret, next).ServeHTTP(w, r)
	})
}

func GetUserId(ctx context.Context) string {

	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}
