package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// RouterConfig holds dependencies for the HTTP router.
type RouterConfig struct {
	Logger      *slog.Logger
	LeadHandler *LeadHandler
}

// NewRouter constructs and returns the application HTTP router.
// All API routes are mounted under /api/v1/.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(requestLogger(cfg.Logger))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/health", healthHandler)

		// Lead routes
		r.Route("/leads", func(r chi.Router) {
			if cfg.LeadHandler != nil {
				r.Mount("/", NewLeadRouter(cfg.LeadHandler))
			} else {
				r.Get("/", notImplementedHandler)
				r.Post("/", notImplementedHandler)
				r.Get("/{id}", notImplementedHandler)
				r.Put("/{id}", notImplementedHandler)
				r.Delete("/{id}", notImplementedHandler)
				r.Patch("/{id}/status", notImplementedHandler)
			}
		})

		// User routes — placeholder
		r.Route("/users", func(r chi.Router) {
			r.Get("/", notImplementedHandler)
			r.Get("/{id}", notImplementedHandler)
		})

		// Team routes — placeholder
		r.Route("/teams", func(r chi.Router) {
			r.Get("/", notImplementedHandler)
			r.Post("/", notImplementedHandler)
			r.Get("/{id}", notImplementedHandler)
		})

		// Metrics routes — placeholder
		r.Route("/metrics", func(r chi.Router) {
			r.Get("/pipeline", notImplementedHandler)
			r.Get("/rep/{id}", notImplementedHandler)
			r.Get("/team/{id}", notImplementedHandler)
		})
	})

	return r
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck // writing to http.ResponseWriter never returns a meaningful error
	w.Write([]byte(`{"status":"ok"}`))
}

func notImplementedHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	//nolint:errcheck // writing to http.ResponseWriter never returns a meaningful error
	w.Write([]byte(`{"error":{"code":"NOT_IMPLEMENTED","message":"endpoint not yet implemented"}}`))
}
