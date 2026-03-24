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
	ConfigHandler      *ConfigHandler
	CollectionHandler  *CollectionHandler
	DemoHandler        *demo.Handler
	WebDistPath        string
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

// mountAPIRoutes registers all /api/v1 resource routes.
func mountAPIRoutes(r chi.Router, cfg RouterConfig) {
	// Target routes
	var targetRouter http.Handler
	if cfg.TargetHandler != nil {
		targetRouter = NewTargetRouter(cfg.TargetHandler)
	}
	mountOrStub(r, "/targets", targetRouter, []string{"/", "/{id}"})

	// Activity routes
	var activityRouter http.Handler
	if cfg.ActivityHandler != nil {
		activityRouter = NewActivityRouter(cfg.ActivityHandler)
	}
	mountOrStub(r, "/activities", activityRouter, []string{"/", "/{id}", "/{id}/submit", "/{id}/status"})

	// Dashboard routes
	var dashboardRouter http.Handler
	if cfg.DashboardHandler != nil {
		dashboardRouter = NewDashboardRouter(cfg.DashboardHandler)
	}
	mountOrStub(r, "/dashboard", dashboardRouter, []string{"/activities", "/coverage", "/frequency"})

	// User routes
	var userRouter http.Handler
	if cfg.UserHandler != nil {
		userRouter = NewUserRouter(cfg.UserHandler)
	}
	mountOrStub(r, "/users", userRouter, []string{"/", "/{id}"})

	// Team routes
	var teamRouter http.Handler
	if cfg.TeamHandler != nil {
		teamRouter = NewTeamRouter(cfg.TeamHandler)
	}
	mountOrStub(r, "/teams", teamRouter, []string{"/", "/{id}"})

	// Collection routes
	var collectionRouter http.Handler
	if cfg.CollectionHandler != nil {
		collectionRouter = NewCollectionRouter(cfg.CollectionHandler)
	}
	mountOrStub(r, "/collections", collectionRouter, []string{"/", "/{id}"})

	// Config route
	if cfg.ConfigHandler != nil {
		r.Get("/config", cfg.ConfigHandler.Get)
	} else {
		r.Get("/config", notImplementedHandler)
	}
}

// mountSPA serves the frontend SPA from the given directory if configured.
func mountSPA(r *chi.Mux, webDistPath string) {
	if webDistPath == "" {
		return
	}
	staticFS := http.Dir(webDistPath)
	fileServer := http.FileServer(staticFS)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the requested file. If it doesn't exist,
		// fall back to index.html for client-side routing.
		path := filepath.Clean(r.URL.Path)
		if _, err := fs.Stat(os.DirFS(webDistPath), path[1:]); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	user, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, user)
}

func notImplementedHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":{"code":"NOT_IMPLEMENTED","message":"endpoint not yet implemented"}}`))
}
