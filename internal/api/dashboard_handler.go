package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// DashboardServicer is the interface the DashboardHandler depends on.
type DashboardServicer interface {
	ActivityStats(ctx context.Context, actor *domain.User, filter store.DashboardFilter) (*service.ActivityStatsResponse, error)
	Coverage(ctx context.Context, actor *domain.User, filter store.DashboardFilter) (*service.CoverageResponse, error)
	Frequency(ctx context.Context, actor *domain.User, filter store.DashboardFilter) (*service.FrequencyResponse, error)
	RecoveryBalance(ctx context.Context, actor *domain.User, filter store.DashboardFilter) (*service.RecoveryBalanceResponse, error)
}

// DashboardHandler handles HTTP requests for dashboard analytics.
type DashboardHandler struct {
	svc DashboardServicer
}

// NewDashboardHandler constructs a DashboardHandler.
func NewDashboardHandler(svc DashboardServicer) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// NewDashboardRouter returns an http.Handler with all dashboard sub-routes mounted.
func NewDashboardRouter(h *DashboardHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/activities", h.ActivityStats)
	r.Get("/coverage", h.Coverage)
	r.Get("/frequency", h.Frequency)
	r.Get("/recovery", h.RecoveryBalance)
	return r
}

// parseDashboardFilter extracts period and scope from query parameters.
// period is expected as YYYY-MM (returns first and last day of that month).
// Alternatively dateFrom/dateTo can be specified as YYYY-MM-DD.
func parseDashboardFilter(r *http.Request) (store.DashboardFilter, error) {
	var filter store.DashboardFilter

	if period := r.URL.Query().Get("period"); period != "" {
		t, err := time.Parse("2006-01", period)
		if err != nil {
			return filter, err
		}
		filter.DateFrom = t
		filter.DateTo = t.AddDate(0, 1, -1)
	} else {
		if v := r.URL.Query().Get("dateFrom"); v != "" {
			t, err := time.Parse("2006-01-02", v)
			if err != nil {
				return filter, err
			}
			filter.DateFrom = t
		} else {
			// Default to current month.
			now := time.Now()
			filter.DateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		}
		if v := r.URL.Query().Get("dateTo"); v != "" {
			t, err := time.Parse("2006-01-02", v)
			if err != nil {
				return filter, err
			}
			filter.DateTo = t
		} else {
			filter.DateTo = filter.DateFrom.AddDate(0, 1, -1)
		}
	}

	if v := r.URL.Query().Get("userId"); v != "" {
		filter.UserID = &v
	}
	if v := r.URL.Query().Get("teamId"); v != "" {
		filter.TeamID = &v
	}

	return filter, nil
}

// ActivityStats handles GET /api/v1/dashboard/activities
func (h *DashboardHandler) ActivityStats(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid period or date format")
		return
	}

	stats, err := h.svc.ActivityStats(r.Context(), actor, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, stats)
}

// Coverage handles GET /api/v1/dashboard/coverage
func (h *DashboardHandler) Coverage(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid period or date format")
		return
	}

	stats, err := h.svc.Coverage(r.Context(), actor, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, stats)
}

// RecoveryBalance handles GET /api/v1/dashboard/recovery
func (h *DashboardHandler) RecoveryBalance(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid period or date format")
		return
	}

	result, err := h.svc.RecoveryBalance(r.Context(), actor, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, result)
}

// Frequency handles GET /api/v1/dashboard/frequency
func (h *DashboardHandler) Frequency(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid period or date format")
		return
	}

	stats, err := h.svc.Frequency(r.Context(), actor, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, stats)
}
