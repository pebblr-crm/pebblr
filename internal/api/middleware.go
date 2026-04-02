package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// defaultMaxBodySize is the default request body limit (1 MB).
// Specific endpoints (e.g., import) can override this per-route.
const defaultMaxBodySize = 1 << 20 // 1 MB

// maxBodySize returns middleware that limits the size of request bodies.
// It wraps r.Body with http.MaxBytesReader so the JSON decoder will fail
// with a clear error rather than consuming unbounded memory.
func maxBodySize(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
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
