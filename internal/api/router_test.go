package api_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pebblr/pebblr/internal/api"
)

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()
	router := api.NewRouter(api.RouterConfig{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
	// Health check bypasses auth middleware
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Auth middleware will reject the fake token in a real implementation.
	// For now, the placeholder auth middleware passes through.
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestNotFoundReturns404(t *testing.T) {
	t.Parallel()
	router := api.NewRouter(api.RouterConfig{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
