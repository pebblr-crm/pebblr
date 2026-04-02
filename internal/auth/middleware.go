package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const claimsKey contextKey = "claims"

// Middleware returns an HTTP middleware that validates Bearer tokens and
// attaches the extracted UserClaims to the request context.
func Middleware(authenticator Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
				return
			}

			claims, err := authenticator.ValidateToken(r.Context(), token)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}

			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeJSONError writes a structured JSON error response with the correct
// Content-Type header. This avoids http.Error which sets text/plain.
func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":{"code":"` + code + `","message":"` + message + `"}}`))
}

// WithClaims stores UserClaims in the context.
func WithClaims(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext retrieves UserClaims from the context.
// Returns nil if no claims are present.
func ClaimsFromContext(ctx context.Context) *UserClaims {
	claims, _ := ctx.Value(claimsKey).(*UserClaims)
	return claims
}

// bearerToken extracts the Bearer token from the Authorization header.
func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	return parts[1], true
}
