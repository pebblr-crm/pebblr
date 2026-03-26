package api

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/pebblr/pebblr/internal/auth"
	"github.com/pebblr/pebblr/internal/auth/demo"
	"github.com/pebblr/pebblr/internal/rbac"
)

// RouterConfig holds dependencies for the HTTP router.
type RouterConfig struct {
	Logger           *slog.Logger
	Authenticator    auth.Authenticator
	TargetHandler    *TargetHandler
	ActivityHandler  *ActivityHandler
	DashboardHandler *DashboardHandler
	TeamHandler      *TeamHandler
	UserHandler      *UserHandler
	ConfigHandler      *ConfigHandler
	CollectionHandler  *CollectionHandler
	TerritoryHandler   *TerritoryHandler
	AuditHandler       *AuditHandler
	DemoHandler        *demo.Handler
	WebDistPath        string
	WebLegacyDistPath  string
}

// NewRouter constructs and returns the application HTTP router.
// All API routes are mounted under /api/v1/.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(requestLogger(cfg.Logger))

	// Kubernetes probe endpoints — outside auth middleware.
	r.Get("/healthz", healthHandler)
	r.Get("/readyz", healthHandler)

	mountDemoRoutes(r, cfg)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware(cfg.Authenticator))
		r.Use(auth.ClaimsBridge)

		r.Get("/health", healthHandler)
		r.Get("/me", meHandler)

		mountAPIRoutes(r, cfg)
	})

	mountDualSPA(r, cfg.WebDistPath, cfg.WebLegacyDistPath)

	return r
}

// mountDemoRoutes registers demo auth endpoints outside the auth middleware.
func mountDemoRoutes(r chi.Router, cfg RouterConfig) {
	if cfg.DemoHandler == nil {
		return
	}
	r.Route("/demo", func(r chi.Router) {
		r.Get("/accounts", cfg.DemoHandler.ListAccounts)
		r.Post("/token", cfg.DemoHandler.IssueToken)
	})
}

// mountOrStub mounts a real router when the handler is non-nil, or registers
// not-implemented stubs for each path.
func mountOrStub(r chi.Router, pattern string, router http.Handler, stubs []string) {
	r.Route(pattern, func(r chi.Router) {
		if router != nil {
			r.Mount("/", router)
		} else {
			for _, s := range stubs {
				r.HandleFunc(s, notImplementedHandler)
			}
		}
	})
}

// routeSpec describes a resource route with its pattern, optional router constructor, and stub paths.
type routeSpec struct {
	pattern string
	router  http.Handler
	stubs   []string
}

// buildRouteSpecs creates the list of resource routes from the config.
func buildRouteSpecs(cfg RouterConfig) []routeSpec {
	return []routeSpec{
		{"/targets", newRouterIfNotNil(cfg.TargetHandler, NewTargetRouter), []string{"/", "/{id}"}},
		{"/activities", newRouterIfNotNil(cfg.ActivityHandler, NewActivityRouter), []string{"/", "/{id}", "/{id}/submit", "/{id}/status"}},
		{"/dashboard", newRouterIfNotNil(cfg.DashboardHandler, NewDashboardRouter), []string{"/activities", "/coverage", "/frequency"}},
		{"/users", newRouterIfNotNil(cfg.UserHandler, NewUserRouter), []string{"/", "/{id}"}},
		{"/teams", newRouterIfNotNil(cfg.TeamHandler, NewTeamRouter), []string{"/", "/{id}"}},
		{"/collections", newRouterIfNotNil(cfg.CollectionHandler, NewCollectionRouter), []string{"/", "/{id}"}},
		{"/territories", newRouterIfNotNil(cfg.TerritoryHandler, NewTerritoryRouter), []string{"/", "/{id}"}},
		{"/audit", newRouterIfNotNil(cfg.AuditHandler, NewAuditRouter), []string{"/", "/{id}/status"}},
	}
}

// newRouterIfNotNil calls the constructor if the handler is non-nil, returning nil otherwise.
func newRouterIfNotNil[T any](handler *T, constructor func(*T) http.Handler) http.Handler {
	if handler == nil {
		return nil
	}
	return constructor(handler)
}

// mountAPIRoutes registers all /api/v1 resource routes.
func mountAPIRoutes(r chi.Router, cfg RouterConfig) {
	for _, spec := range buildRouteSpecs(cfg) {
		mountOrStub(r, spec.pattern, spec.router, spec.stubs)
	}

	if cfg.ConfigHandler != nil {
		r.Get("/config", cfg.ConfigHandler.Get)
	} else {
		r.Get("/config", notImplementedHandler)
	}
}

const uiCookieName = "pebblr_ui"

// mountDualSPA serves the current frontend by default and the legacy frontend
// when the pebblr_ui cookie is set to "legacy" (or the old "v1" value).
// The ?ui=legacy (or ?ui=v1) query parameter switches to the legacy UI.
// The ?ui=default (or ?ui=v2) query parameter switches back to the current UI.
// If legacyDistPath is empty, only the current frontend is served.
func mountDualSPA(r *chi.Mux, webDistPath, legacyDistPath string) {
	if webDistPath == "" {
		return
	}

	webServer := http.FileServer(http.Dir(webDistPath))

	var legacyServer http.Handler
	if legacyDistPath != "" {
		legacyServer = http.FileServer(http.Dir(legacyDistPath))
	}

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		if handleUISwitch(w, req) {
			return
		}
		distPath, fileServer := selectSPA(req, webDistPath, webServer, legacyDistPath, legacyServer)
		serveSPA(w, req, distPath, fileServer)
	})
}

// handleUISwitch checks for the ?ui= query parameter, sets the preference
// cookie, and redirects. Returns true if a redirect was issued.
func handleUISwitch(w http.ResponseWriter, req *http.Request) bool {
	uiParam := req.URL.Query().Get("ui")

	var cookieValue string
	switch uiParam {
	case "legacy", "v1":
		cookieValue = "legacy"
	case "default", "v2":
		cookieValue = "default"
	default:
		return false
	}

	http.SetCookie(w, &http.Cookie{
		Name:     uiCookieName,
		Value:    cookieValue,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		HttpOnly: true,
	})

	dest := sanitizeRedirectPath(req)
	http.Redirect(w, req, dest, http.StatusFound)
	return true
}

// sanitizeRedirectPath builds a safe, same-origin redirect target from the
// request path, stripping the "ui" query parameter. Only relative paths
// (starting with "/") are allowed; anything else falls back to "/".
func sanitizeRedirectPath(req *http.Request) string {
	path := req.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/"
	}
	// Reject any path containing sequences that could be used for open redirects.
	if strings.HasPrefix(path, "//") || strings.Contains(path, "://") {
		path = "/"
	}
	path = filepath.Clean(path)

	q := req.URL.Query()
	q.Del("ui")
	if encoded := q.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}

// selectSPA picks the file server and dist path based on the UI preference cookie.
// The current web frontend is served by default; legacy is served only when the
// cookie is "legacy" or the old "v1" value.
func selectSPA(req *http.Request, webDist string, webServer http.Handler, legacyDist string, legacyServer http.Handler) (string, http.Handler) {
	if legacyServer != nil {
		if c, err := req.Cookie(uiCookieName); err == nil && (c.Value == "legacy" || c.Value == "v1") {
			return legacyDist, legacyServer
		}
	}
	return webDist, webServer
}

// serveSPA tries to serve a static file and falls back to index.html for
// client-side routing.
func serveSPA(w http.ResponseWriter, req *http.Request, distPath string, fileServer http.Handler) {
	path := filepath.Clean(req.URL.Path)
	if _, err := fs.Stat(os.DirFS(distPath), path[1:]); err == nil {
		fileServer.ServeHTTP(w, req)
		return
	}
	req.URL.Path = "/"
	fileServer.ServeHTTP(w, req)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	user, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, user)
}

func notImplementedHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":{"code":"NOT_IMPLEMENTED","message":"endpoint not yet implemented"}}`))
}
