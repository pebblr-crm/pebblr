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

// setupSPADirs creates temp v1 and v2 dist directories with distinct index.html files.
func setupSPADirs(t *testing.T) (v1Dir, v2Dir string) {
	t.Helper()

	v1Dir = t.TempDir()
	if err := os.WriteFile(filepath.Join(v1Dir, "index.html"), []byte("v1-index"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(v1Dir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(v1Dir, "assets", "app.js"), []byte("v1-js"), 0644); err != nil {
		t.Fatal(err)
	}

	v2Dir = t.TempDir()
	if err := os.WriteFile(filepath.Join(v2Dir, "index.html"), []byte("v2-index"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(v2Dir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(v2Dir, "assets", "app.js"), []byte("v2-js"), 0644); err != nil {
		t.Fatal(err)
	}

	return v1Dir, v2Dir
}

func dualSPARouter(v1, v2 string) http.Handler {
	return api.NewRouter(api.RouterConfig{
		Logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Authenticator: auth.NewStaticAuthenticator("test-token"),
		WebDistPath:   v1,
		WebV2DistPath: v2,
	})
}

func TestDualSPA_DefaultServesV1(t *testing.T) {
	t.Parallel()
	v1Dir, v2Dir := setupSPADirs(t)
	router := dualSPARouter(v1Dir, v2Dir)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "v1-index" {
		t.Errorf("expected v1-index, got %q", w.Body.String())
	}
}

func TestDualSPA_QueryParamSetsCookieAndRedirects(t *testing.T) {
	t.Parallel()
	v1Dir, v2Dir := setupSPADirs(t)
	router := dualSPARouter(v1Dir, v2Dir)

	req := httptest.NewRequest(http.MethodGet, "/?ui=v2", http.NoBody)
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
		if c.Name == "pebblr_ui" && c.Value == "v2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pebblr_ui=v2 cookie to be set")
	}
}

func TestDualSPA_V2CookieServesV2(t *testing.T) {
	t.Parallel()
	v1Dir, v2Dir := setupSPADirs(t)
	router := dualSPARouter(v1Dir, v2Dir)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "v2"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "v2-index" {
		t.Errorf("expected v2-index, got %q", w.Body.String())
	}
}

func TestDualSPA_V1CookieServesV1(t *testing.T) {
	t.Parallel()
	v1Dir, v2Dir := setupSPADirs(t)
	router := dualSPARouter(v1Dir, v2Dir)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "v1"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "v1-index" {
		t.Errorf("expected v1-index, got %q", w.Body.String())
	}
}

func TestDualSPA_SPAFallbackWithV2Cookie(t *testing.T) {
	t.Parallel()
	v1Dir, v2Dir := setupSPADirs(t)
	router := dualSPARouter(v1Dir, v2Dir)

	// Request a client-side route — should fall back to v2 index.html
	req := httptest.NewRequest(http.MethodGet, "/planner", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "v2"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "v2-index" {
		t.Errorf("expected v2-index for SPA fallback, got %q", w.Body.String())
	}
}

func TestDualSPA_StaticAssetsFromCorrectDist(t *testing.T) {
	t.Parallel()
	v1Dir, v2Dir := setupSPADirs(t)
	router := dualSPARouter(v1Dir, v2Dir)

	// v1 static asset
	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Body.String() != "v1-js" {
		t.Errorf("expected v1-js, got %q", w.Body.String())
	}

	// v2 static asset
	req2 := httptest.NewRequest(http.MethodGet, "/assets/app.js", http.NoBody)
	req2.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "v2"})
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Body.String() != "v2-js" {
		t.Errorf("expected v2-js, got %q", w2.Body.String())
	}
}

func TestDualSPA_EmptyV2PathIgnoresCookie(t *testing.T) {
	t.Parallel()
	v1Dir, _ := setupSPADirs(t)
	router := dualSPARouter(v1Dir, "") // no v2 path

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "v2"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Body.String() != "v1-index" {
		t.Errorf("expected v1-index when v2 path empty, got %q", w.Body.String())
	}
}

func TestDualSPA_SwitchBackToV1(t *testing.T) {
	t.Parallel()
	v1Dir, v2Dir := setupSPADirs(t)
	router := dualSPARouter(v1Dir, v2Dir)

	req := httptest.NewRequest(http.MethodGet, "/?ui=v1", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "pebblr_ui", Value: "v2"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "pebblr_ui" && c.Value == "v1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pebblr_ui=v1 cookie to be set")
	}
}

func TestDualSPA_QueryParamPreservesOtherParams(t *testing.T) {
	t.Parallel()
	v1Dir, v2Dir := setupSPADirs(t)
	router := dualSPARouter(v1Dir, v2Dir)

	req := httptest.NewRequest(http.MethodGet, "/targets?ui=v2&page=3", http.NoBody)
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
