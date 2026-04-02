package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// AuditServicer is the interface the AuditHandler depends on.
type AuditServicer interface {
	List(ctx context.Context, actor *domain.User, filter store.AuditFilter) ([]*domain.AuditEntry, int, error)
	UpdateStatus(ctx context.Context, actor *domain.User, id, status string) error
}

// AuditHandler handles HTTP requests for audit log operations.
type AuditHandler struct {
	svc AuditServicer
}

// NewAuditHandler constructs an AuditHandler.
func NewAuditHandler(svc AuditServicer) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// NewAuditRouter returns an http.Handler with all audit sub-routes.
func NewAuditRouter(h *AuditHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Patch("/{id}/status", h.UpdateStatus)
	return r
}

func mapAuditServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "audit entry not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
	}
}

// List handles GET /api/v1/audit
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	filter := store.AuditFilter{
		Page:  1,
		Limit: 50,
	}

	if v := r.URL.Query().Get("entityType"); v != "" {
		filter.EntityType = &v
	}
	if v := r.URL.Query().Get("actorId"); v != "" {
		filter.ActorID = &v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filter.Status = &v
	}
	if v := r.URL.Query().Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 && l <= 200 {
			filter.Limit = l
		}
	}

	entries, total, err := h.svc.List(r.Context(), actor, filter)
	if err != nil {
		mapAuditServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	writeJSON(w, r, map[string]any{
		"items": entries,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateStatus handles PATCH /api/v1/audit/{id}/status
func (h *AuditHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	id := chi.URLParam(r, "id")

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), actor, id, req.Status); err != nil {
		mapAuditServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
