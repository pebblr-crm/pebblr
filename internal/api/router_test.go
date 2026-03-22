package api_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/auth"
)

func testRouter() http.Handler {
	return api.NewRouter(api.RouterConfig{
		Logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Authenticator: auth.NewStaticAuthenticator("test-token"),
	})
}

func TestHealthzEndpoint(t *testing.T) {
	t.Parallel()
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /healthz, got %d", w.Code)
	}
}

func TestAuthenticatedHealthEndpoint(t *testing.T) {
	t.Parallel()
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUnauthenticatedReturns401(t *testing.T) {
	t.Parallel()
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}
}

func TestWrongTokenReturns401(t *testing.T) {
	t.Parallel()
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong token, got %d", w.Code)
	}
}
