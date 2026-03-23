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

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware(cfg.Authenticator))
		r.Use(auth.ClaimsBridge)

		r.Get("/health", healthHandler)

		// Current user endpoint
		r.Get("/me", meHandler)

		// Target routes
		r.Route("/targets", func(r chi.Router) {
			if cfg.TargetHandler != nil {
				r.Mount("/", NewTargetRouter(cfg.TargetHandler))
			} else {
				r.Get("/", notImplementedHandler)
				r.Post("/", notImplementedHandler)
				r.Get("/{id}", notImplementedHandler)
				r.Put("/{id}", notImplementedHandler)
			}
		})

		// Activity routes
		r.Route("/activities", func(r chi.Router) {
			if cfg.ActivityHandler != nil {
				r.Mount("/", NewActivityRouter(cfg.ActivityHandler))
			} else {
				r.Get("/", notImplementedHandler)
				r.Post("/", notImplementedHandler)
				r.Get("/{id}", notImplementedHandler)
				r.Put("/{id}", notImplementedHandler)
				r.Delete("/{id}", notImplementedHandler)
				r.Post("/{id}/submit", notImplementedHandler)
				r.Patch("/{id}/status", notImplementedHandler)
			}
		})

		// Dashboard routes
		r.Route("/dashboard", func(r chi.Router) {
			if cfg.DashboardHandler != nil {
				r.Mount("/", NewDashboardRouter(cfg.DashboardHandler))
			} else {
				r.Get("/activities", notImplementedHandler)
				r.Get("/coverage", notImplementedHandler)
				r.Get("/frequency", notImplementedHandler)
			}
		})

		// User routes
		r.Route("/users", func(r chi.Router) {
			if cfg.UserHandler != nil {
				r.Mount("/", NewUserRouter(cfg.UserHandler))
			} else {
				r.Get("/", notImplementedHandler)
				r.Get("/{id}", notImplementedHandler)
			}
		})

		// Team routes
		r.Route("/teams", func(r chi.Router) {
			if cfg.TeamHandler != nil {
				r.Mount("/", NewTeamRouter(cfg.TeamHandler))
			} else {
				r.Get("/", notImplementedHandler)
				r.Get("/{id}", notImplementedHandler)
			}
		})

		// Collection routes
		r.Route("/collections", func(r chi.Router) {
			if cfg.CollectionHandler != nil {
				r.Mount("/", NewCollectionRouter(cfg.CollectionHandler))
			} else {
				r.Get("/", notImplementedHandler)
				r.Post("/", notImplementedHandler)
				r.Get("/{id}", notImplementedHandler)
				r.Put("/{id}", notImplementedHandler)
				r.Delete("/{id}", notImplementedHandler)
			}
		})

		// Config route
		if cfg.ConfigHandler != nil {
			r.Get("/config", cfg.ConfigHandler.Get)
		} else {
			r.Get("/config", notImplementedHandler)
		}

	})

	// Serve frontend SPA from WebDistPath if configured.
	if cfg.WebDistPath != "" {
		staticFS := http.Dir(cfg.WebDistPath)
		fileServer := http.FileServer(staticFS)
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the requested file. If it doesn't exist,
			// fall back to index.html for client-side routing.
			path := filepath.Clean(r.URL.Path)
			if _, err := fs.Stat(os.DirFS(cfg.WebDistPath), path[1:]); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	}

	return r
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
