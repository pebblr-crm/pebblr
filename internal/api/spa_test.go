package api_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/auth"
)

const webIndexContent = "web-index"

// setupSPADir creates a temp dist directory with an index.html and a static asset.
func setupSPADir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(webIndexContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("web-js"), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func spaRouter(webDir string) http.Handler {
	return api.NewRouter(api.RouterConfig{
		Logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Authenticator: auth.NewStaticAuthenticator("test-token"),
		WebDistPath:   webDir,
	})
}

func TestSPA_ServesIndex(t *testing.T) {
	t.Parallel()
	router := spaRouter(setupSPADir(t))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != webIndexContent {
		t.Errorf("expected %s, got %q", webIndexContent, w.Body.String())
	}
}

func TestSPA_ServesStaticAssets(t *testing.T) {
	t.Parallel()
	router := spaRouter(setupSPADir(t))

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "web-js" {
		t.Errorf("expected web-js, got %q", w.Body.String())
	}
}

func TestSPA_FallbackToIndex(t *testing.T) {
	t.Parallel()
	router := spaRouter(setupSPADir(t))

	// Client-side route should fall back to index.html
	req := httptest.NewRequest(http.MethodGet, "/planner", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != webIndexContent {
		t.Errorf("expected %s for SPA fallback, got %q", webIndexContent, w.Body.String())
	}
}

func TestSPA_EmptyDistPathNoOp(t *testing.T) {
	t.Parallel()
	router := spaRouter("") // no dist path

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// With no dist path, NotFound is not mounted — chi returns 404
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 with empty dist path, got %d", w.Code)
	}
}
