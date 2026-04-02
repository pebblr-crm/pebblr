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

// TerritoryServicer is the interface the TerritoryHandler depends on.
type TerritoryServicer interface {
	List(ctx context.Context, actor *domain.User) ([]*domain.Territory, error)
	Get(ctx context.Context, actor *domain.User, id string) (*domain.Territory, error)
	Create(ctx context.Context, actor *domain.User, t *domain.Territory) (*domain.Territory, error)
	Update(ctx context.Context, actor *domain.User, t *domain.Territory) (*domain.Territory, error)
	Delete(ctx context.Context, actor *domain.User, id string) error
}

// TerritoryHandler handles HTTP requests for territories.
type TerritoryHandler struct {
	svc TerritoryServicer
}

// NewTerritoryHandler constructs a TerritoryHandler.
func NewTerritoryHandler(svc TerritoryServicer) *TerritoryHandler {
	return &TerritoryHandler{svc: svc}
}

// NewTerritoryRouter returns an http.Handler with all territory sub-routes.
func NewTerritoryRouter(h *TerritoryHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

type territoryRequest struct {
	Name     string         `json:"name"`
	TeamID   string         `json:"teamId"`
	Region   string         `json:"region"`
	Boundary map[string]any `json:"boundary"`
}

func mapTerritoryServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "territory not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
	}
}

// List handles GET /api/v1/territories
func (h *TerritoryHandler) List(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	territories, err := h.svc.List(r.Context(), actor)
	if err != nil {
		mapTerritoryServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, map[string]any{
		"items": territories,
		"total": len(territories),
	})
}

// Get handles GET /api/v1/territories/{id}
func (h *TerritoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	id := chi.URLParam(r, "id")
	territory, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		mapTerritoryServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, territory)
}

// Create handles POST /api/v1/territories
func (h *TerritoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	var req territoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return
	}

	t := &domain.Territory{
		Name:     req.Name,
		TeamID:   req.TeamID,
		Region:   req.Region,
		Boundary: req.Boundary,
	}

	created, err := h.svc.Create(r.Context(), actor, t)
	if err != nil {
		mapTerritoryServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, r, created)
}

// Update handles PUT /api/v1/territories/{id}
func (h *TerritoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	id := chi.URLParam(r, "id")

	var req territoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return
	}

	t := &domain.Territory{
		ID:       id,
		Name:     req.Name,
		TeamID:   req.TeamID,
		Region:   req.Region,
		Boundary: req.Boundary,
	}

	updated, err := h.svc.Update(r.Context(), actor, t)
	if err != nil {
		mapTerritoryServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, updated)
}

// Delete handles DELETE /api/v1/territories/{id}
func (h *TerritoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), actor, id); err != nil {
		mapTerritoryServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
