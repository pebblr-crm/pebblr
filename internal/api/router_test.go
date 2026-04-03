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

const (
	pathHealth      = "/api/v1/health"
	testToken       = "test-token"
	bearerTestToken = "Bearer test-token"
)

func testRouter() http.Handler {
	return api.NewRouter(api.RouterConfig{
		Logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Authenticator: auth.NewStaticAuthenticator(testToken),
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

	req := httptest.NewRequest(http.MethodGet, pathHealth, http.NoBody)
	req.Header.Set("Authorization", bearerTestToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUnauthenticatedReturns401(t *testing.T) {
	t.Parallel()
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, pathHealth, http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}
}

func TestWrongTokenReturns401(t *testing.T) {
	t.Parallel()
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, pathHealth, http.NoBody)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong token, got %d", w.Code)
	}
}

func TestReadyzEndpoint(t *testing.T) {
	t.Parallel()
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/readyz", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /readyz, got %d", w.Code)
	}
}

func TestMeEndpointReturnsUser(t *testing.T) {
	t.Parallel()
	router := testRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", http.NoBody)
	req.Header.Set("Authorization", bearerTestToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Static auth creates a user, so /me should return 200
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for /me with valid auth, got %d", w.Code)
	}
}

func TestNotImplementedStubs(t *testing.T) {
	t.Parallel()

	// Router with no handlers — all routes should return 501
	router := api.NewRouter(api.RouterConfig{
		Logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Authenticator: auth.NewStaticAuthenticator(testToken),
	})

	paths := []string{
		"/api/v1/targets/",
		"/api/v1/activities/",
		"/api/v1/dashboard/activities",
		"/api/v1/users/",
		"/api/v1/teams/",
		"/api/v1/collections/",
		"/api/v1/territories/",
		"/api/v1/audit/",
		"/api/v1/config",
	}

	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, p, http.NoBody)
			req.Header.Set("Authorization", bearerTestToken)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotImplemented {
				t.Errorf("expected 501 for %s, got %d", p, w.Code)
			}
		})
	}
}

func TestSPAFallback(t *testing.T) {
	t.Parallel()

	// Create a temp directory with an index.html
	dir := t.TempDir()
	indexPath := dir + "/index.html"
	if err := os.WriteFile(indexPath, []byte("<html>SPA</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Also create a static asset
	if err := os.WriteFile(dir+"/style.css", []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	router := api.NewRouter(api.RouterConfig{
		Logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Authenticator: auth.NewStaticAuthenticator(testToken),
		WebDistPath:   dir,
	})

	// Unknown path should fall back to index.html
	req := httptest.NewRequest(http.MethodGet, "/some-spa-route", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for SPA fallback, got %d", w.Code)
	}

	// Known static file should be served directly
	req2 := httptest.NewRequest(http.MethodGet, "/style.css", http.NoBody)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 for static file, got %d", w2.Code)
	}
}
