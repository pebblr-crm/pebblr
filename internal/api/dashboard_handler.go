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

// parseDashboardDateRange parses the date range from query parameters.
// Supports "period" (YYYY-MM) or explicit "dateFrom"/"dateTo" (YYYY-MM-DD).
func parseDashboardDateRange(r *http.Request) (from, to time.Time, err error) {
	if period := r.URL.Query().Get("period"); period != "" {
		t, parseErr := time.Parse("2006-01", period)
		if parseErr != nil {
			return time.Time{}, time.Time{}, parseErr
		}
		return t, t.AddDate(0, 1, -1), nil
	}

	from, err = parseDateFromParam(r, "dateFrom")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if from.IsZero() {
		now := time.Now()
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	to, err = parseDateFromParam(r, "dateTo")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if to.IsZero() {
		to = from.AddDate(0, 1, -1)
	}

	return from, to, nil
}

// parseDateFromParam parses a YYYY-MM-DD date from the named query parameter.
// Returns zero time if the parameter is absent or empty.
func parseDateFromParam(r *http.Request, param string) (time.Time, error) {
	v := r.URL.Query().Get(param)
	if v == "" {
		return time.Time{}, nil
	}
	return time.Parse(dateFormat, v)
}

// parseDashboardFilter extracts period and scope from query parameters.
// period is expected as YYYY-MM (returns first and last day of that month).
// Alternatively dateFrom/dateTo can be specified as YYYY-MM-DD.
func parseDashboardFilter(r *http.Request) (store.DashboardFilter, error) {
	var filter store.DashboardFilter

	from, to, err := parseDashboardDateRange(r)
	if err != nil {
		return filter, err
	}
	filter.DateFrom = from
	filter.DateTo = to

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
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidPeriod)
		return
	}

	stats, err := h.svc.ActivityStats(r.Context(), actor, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, stats)
}

// Coverage handles GET /api/v1/dashboard/coverage
func (h *DashboardHandler) Coverage(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidPeriod)
		return
	}

	stats, err := h.svc.Coverage(r.Context(), actor, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, stats)
}

// RecoveryBalance handles GET /api/v1/dashboard/recovery
func (h *DashboardHandler) RecoveryBalance(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidPeriod)
		return
	}

	result, err := h.svc.RecoveryBalance(r.Context(), actor, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, result)
}

// Frequency handles GET /api/v1/dashboard/frequency
func (h *DashboardHandler) Frequency(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	filter, err := parseDashboardFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidPeriod)
		return
	}

	stats, err := h.svc.Frequency(r.Context(), actor, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, stats)
}
