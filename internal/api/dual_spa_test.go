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

// setupSPADirs creates temp web and legacy dist directories with distinct index.html files.
func setupSPADirs(t *testing.T) (webDir, legacyDir string) {
	t.Helper()

	webDir = t.TempDir()
	if err := os.WriteFile(filepath.Join(webDir, "index.html"), []byte("web-index"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(webDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(webDir, "assets", "app.js"), []byte("web-js"), 0o644); err != nil {
		t.Fatal(err)
	}

	legacyDir = t.TempDir()
	if err := os.WriteFile(filepath.Join(legacyDir, "index.html"), []byte("legacy-index"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(legacyDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "assets", "app.js"), []byte("legacy-js"), 0o644); err != nil {
		t.Fatal(err)
	}

	return webDir, legacyDir
}

func dualSPARouter(web, legacy string) http.Handler {
	return api.NewRouter(api.RouterConfig{
		Logger:            slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Authenticator:     auth.NewStaticAuthenticator("test-token"),
		WebDistPath:       web,
		WebLegacyDistPath: legacy,
	})
}

func TestDualSPA_DefaultServesWeb(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "web-index" {
		t.Errorf("expected web-index, got %q", w.Body.String())
	}
}

func TestDualSPA_LegacyQueryParamSetsCookieAndRedirects(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/?ui=legacy", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("expected redirect to /, got %q", location)
	}

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "pebblr_ui" && c.Value == "legacy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pebblr_ui=legacy cookie to be set")
	}
}

func TestDualSPA_V1QueryParamSetsLegacyCookie(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/?ui=v1", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "pebblr_ui" && c.Value == "legacy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pebblr_ui=legacy cookie for ?ui=v1 (backward compat)")
	}
}

func TestDualSPA_LegacyCookieServesLegacy(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "legacy"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "legacy-index" {
		t.Errorf("expected legacy-index, got %q", w.Body.String())
	}
}

func TestDualSPA_OldV1CookieServesLegacy(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "v1"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "legacy-index" {
		t.Errorf("expected legacy-index for old v1 cookie, got %q", w.Body.String())
	}
}

func TestDualSPA_DefaultCookieServesWeb(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "default"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "web-index" {
		t.Errorf("expected web-index, got %q", w.Body.String())
	}
}

func TestDualSPA_OldV2CookieServesWeb(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "v2"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "web-index" {
		t.Errorf("expected web-index for old v2 cookie, got %q", w.Body.String())
	}
}

func TestDualSPA_SPAFallbackWithLegacyCookie(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/planner", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "legacy"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "legacy-index" {
		t.Errorf("expected legacy-index for SPA fallback, got %q", w.Body.String())
	}
}

func TestDualSPA_StaticAssetsFromCorrectDist(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	// Default static asset (web)
	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Body.String() != "web-js" {
		t.Errorf("expected web-js, got %q", w.Body.String())
	}

	// Legacy static asset
	req2 := httptest.NewRequest(http.MethodGet, "/assets/app.js", http.NoBody)
	req2.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "legacy"})
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Body.String() != "legacy-js" {
		t.Errorf("expected legacy-js, got %q", w2.Body.String())
	}
}

func TestDualSPA_EmptyLegacyPathIgnoresCookie(t *testing.T) {
	t.Parallel()
	webDir, _ := setupSPADirs(t)
	router := dualSPARouter(webDir, "") // no legacy path

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "legacy"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "web-index" {
		t.Errorf("expected web-index when legacy path empty, got %q", w.Body.String())
	}
}

func TestDualSPA_SwitchBackToDefault(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/?ui=default", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "legacy"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "pebblr_ui" && c.Value == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pebblr_ui=default cookie to be set")
	}
}

func TestDualSPA_V2QueryParamSwitchesToDefault(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/?ui=v2", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "legacy"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "pebblr_ui" && c.Value == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pebblr_ui=default cookie for ?ui=v2 (backward compat)")
	}
}

func TestDualSPA_QueryParamPreservesOtherParams(t *testing.T) {
	t.Parallel()
	webDir, legacyDir := setupSPADirs(t)
	router := dualSPARouter(webDir, legacyDir)

	req := httptest.NewRequest(http.MethodGet, "/targets?ui=legacy&page=3", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/targets?page=3" {
		t.Errorf("expected /targets?page=3, got %q", location)
	}
}
