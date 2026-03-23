package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// DashboardServicer is the interface the DashboardHandler depends on.
type DashboardServicer interface {
	ActivityStats(ctx context.Context, actor *domain.User, period string, filter store.DashboardFilter) (*service.DashboardStatsResponse, error)
	CoverageStats(ctx context.Context, actor *domain.User, period string, filter store.DashboardFilter) (*store.CoverageStats, error)
	UserStats(ctx context.Context, actor *domain.User, period string, filter store.DashboardFilter) ([]store.UserActivityStats, error)
}

// DashboardHandler handles HTTP requests for dashboard metrics.
type DashboardHandler struct {
	svc DashboardServicer
}

// NewDashboardHandler constructs a DashboardHandler backed by the given service.
func NewDashboardHandler(svc DashboardServicer) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// NewDashboardRouter returns an http.Handler with all dashboard sub-routes mounted.
func NewDashboardRouter(h *DashboardHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/stats", h.Stats)
	r.Get("/coverage", h.Coverage)
	r.Get("/user-stats", h.UserStats)
	return r
}

func parseDashboardFilter(r *http.Request) store.DashboardFilter {
	var filter store.DashboardFilter
	if v := r.URL.Query().Get("creatorId"); v != "" {
		filter.CreatorID = &v
	}
	if v := r.URL.Query().Get("teamId"); v != "" {
		filter.TeamID = &v
	}
	return filter
}

func mapDashboardServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	}
}

// Stats handles GET /api/v1/dashboard/stats?period=YYYY-MM&teamId=...&creatorId=...
func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "period query parameter is required (format: YYYY-MM)")
		return
	}

	filter := parseDashboardFilter(r)
	result, err := h.svc.ActivityStats(r.Context(), actor, period, filter)
	if err != nil {
		mapDashboardServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// Coverage handles GET /api/v1/dashboard/coverage?period=YYYY-MM&teamId=...&creatorId=...
func (h *DashboardHandler) Coverage(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "period query parameter is required (format: YYYY-MM)")
		return
	}

	filter := parseDashboardFilter(r)
	result, err := h.svc.CoverageStats(r.Context(), actor, period, filter)
	if err != nil {
		mapDashboardServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// UserStats handles GET /api/v1/dashboard/user-stats?period=YYYY-MM&teamId=...
func (h *DashboardHandler) UserStats(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "period query parameter is required (format: YYYY-MM)")
		return
	}

	filter := parseDashboardFilter(r)
	result, err := h.svc.UserStats(r.Context(), actor, period, filter)
	if err != nil {
		mapDashboardServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"users": result})
}
