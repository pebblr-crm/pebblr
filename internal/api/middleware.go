package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// devAuthMiddleware injects a default admin user into the request context,
// bypassing token validation. Only for local development.
func devAuthMiddleware(next http.Handler) http.Handler {
	devUser := &domain.User{
		ID:       "a0000000-0000-0000-0000-000000000001",
		Email:    "admin@pebblr.dev",
		Name:     "Alex Admin",
		Role:     domain.RoleAdmin,
		TeamIDs:  []string{},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := rbac.WithUser(r.Context(), devUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authMiddleware validates Bearer tokens from the Authorization header.
// TODO: wire up the auth.Authenticator once implemented.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format")
			return
		}

		// TODO: validate token via auth.Authenticator and attach claims to context.
		_ = parts[1]

		next.ServeHTTP(w, r)
	})
}

type contextKey string

const loggerKey contextKey = "logger"

// requestLogger returns middleware that logs each request using slog.
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := middleware.GetReqID(r.Context())

			l := logger.With(
				"request_id", reqID,
				"method", r.Method,
				"path", r.URL.Path,
			)

			ctx := context.WithValue(r.Context(), loggerKey, l)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r.WithContext(ctx))

			l.Info("request completed",
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// LoggerFromContext retrieves the request-scoped logger from context.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
