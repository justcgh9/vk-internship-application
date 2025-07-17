package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/justcgh9/vk-internship-application/internal/service/auth"
)

type contextKey string

const userIDKey contextKey = "user_id"
const authHeader = "Authorization"
const bearerPrefix = "Bearer "

func AuthMiddleware(authSvc auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeaderVal := r.Header.Get(authHeader)
			if authHeaderVal == "" || !strings.HasPrefix(authHeaderVal, bearerPrefix) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeaderVal, bearerPrefix)
			userID, err := authSvc.VerifyToken(token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuthMiddleware(authSvc auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeaderVal := r.Header.Get(authHeader)
			if authHeaderVal != "" {
				if !strings.HasPrefix(authHeaderVal, bearerPrefix) {
					http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
					return
				}

				token := strings.TrimPrefix(authHeaderVal, bearerPrefix)
				userID, err := authSvc.VerifyToken(token)
				if err != nil {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				ctx := context.WithValue(r.Context(), userIDKey, userID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}



func GetUserID(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(userIDKey).(int64)
	return uid, ok
}
