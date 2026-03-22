package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
)

// DashboardServicer is the interface the DashboardHandler depends on for business logic.
type DashboardServicer interface {
	Stats(ctx context.Context) (*service.DashboardStats, error)
}

// DashboardHandler handles HTTP requests for dashboard analytics.
type DashboardHandler struct {
	svc DashboardServicer
}

// NewDashboardHandler constructs a DashboardHandler backed by the given service.
func NewDashboardHandler(svc DashboardServicer) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Stats handles GET /api/v1/dashboard/stats
func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	_, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stats)
}
