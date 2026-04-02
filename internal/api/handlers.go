package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// errorResponse is the standard error envelope for all API error responses.
type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeJSON encodes v as JSON into the response body. If encoding fails, the
// error is logged — the status code and headers are already written at that
// point, so there is nothing else we can do for the client.
func writeJSON(w http.ResponseWriter, r *http.Request, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		LoggerFromContext(r.Context()).Error("failed to encode JSON response", "err", err)
	}
}

// requireActor extracts the authenticated user from the request context.
// If the user is missing, it writes a 401 JSON error and returns nil.
// Handlers should return early when nil is returned.
func requireActor(w http.ResponseWriter, r *http.Request) *domain.User {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return nil
	}
	return actor
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{
		Error: errorDetail{Code: code, Message: message},
	}); err != nil {
		slog.Default().Error("failed to encode error response", "err", err)
	}
}
