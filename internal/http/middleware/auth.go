package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/justcgh9/vk-internship-application/internal/service/auth"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
)

type contextKey string

const userIDKey contextKey = "user_id"
const authHeader = "Authorization"
const bearerPrefix = "Bearer "

func AuthMiddleware(authSvc auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.
				FromContext(r.Context()).
				With("component", "middleware").
				With("function", "auth")

			authHeaderVal := r.Header.Get(authHeader)
			if authHeaderVal == "" || !strings.HasPrefix(authHeaderVal, bearerPrefix) {
				log.Warn("missing or malformed Authorization header")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeaderVal, bearerPrefix)
			userID, err := authSvc.VerifyToken(token)
			if err != nil {
				log.Warn("token verification failed", slog.String("err", err.Error()))
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			log.Info("user authenticated", slog.Int64("user_id", userID))
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuthMiddleware(authSvc auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.
				FromContext(r.Context()).
				With("component", "middleware").
				With("function", "optional_auth")

			authHeaderVal := r.Header.Get(authHeader)
			if authHeaderVal != "" {
				if !strings.HasPrefix(authHeaderVal, bearerPrefix) {
					log.Warn("malformed Authorization header")
					http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
					return
				}

				token := strings.TrimPrefix(authHeaderVal, bearerPrefix)
				userID, err := authSvc.VerifyToken(token)
				if err != nil {
					log.Warn("invalid token in optional auth", slog.String("err", err.Error()))
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				log.Info("optional auth: user authenticated", slog.Int64("user_id", userID))
				r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
			} else {
				log.Debug("no auth header, continuing unauthenticated")
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserID(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(userIDKey).(int64)
	return uid, ok
}

func WithUserID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}
