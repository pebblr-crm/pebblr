package api

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

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
	ConfigHandler    *ConfigHandler
	CollectionHandler *CollectionHandler
	TerritoryHandler *TerritoryHandler
	AuditHandler     *AuditHandler
	DemoHandler      *demo.Handler
	WebDistPath      string
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
		r.Use(maxBodySize(defaultMaxBodySize))
		r.Use(auth.Middleware(cfg.Authenticator))
		r.Use(auth.ClaimsBridge)

		r.Get("/health", healthHandler)
		r.Get("/me", meHandler)

		mountAPIRoutes(r, cfg)
	})

	mountSPA(r, cfg.WebDistPath)

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

// mountSPA serves the frontend SPA from the given dist directory.
// Static files are served directly; all other paths fall back to index.html
// for client-side routing.
func mountSPA(r *chi.Mux, distPath string) {
	if distPath == "" {
		return
	}

	fileServer := http.FileServer(http.Dir(distPath))

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		path := filepath.Clean(req.URL.Path)
		if _, err := fs.Stat(os.DirFS(distPath), path[1:]); err == nil {
			fileServer.ServeHTTP(w, req)
			return
		}
		req.URL.Path = "/"
		fileServer.ServeHTTP(w, req)
	})
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
