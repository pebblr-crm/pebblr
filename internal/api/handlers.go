package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
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

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{
		Error: errorDetail{Code: code, Message: message},
	}); err != nil {
		// At this point the status code and headers are already written. The
		// request context (and therefore the request-scoped logger) is not
		// available in this helper, so fall back to the default logger.
		slog.Default().Error("failed to encode error response", "err", err)
	}
}
