package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// CollectionServicer is the interface the CollectionHandler depends on.
type CollectionServicer interface {
	Create(ctx context.Context, actor *domain.User, name string, targetIDs []string) (*domain.Collection, error)
	List(ctx context.Context, actor *domain.User) ([]*domain.Collection, error)
	Get(ctx context.Context, actor *domain.User, id string) (*domain.Collection, error)
	Update(ctx context.Context, actor *domain.User, id string, name string, targetIDs []string) (*domain.Collection, error)
	Delete(ctx context.Context, actor *domain.User, id string) error
}

// CollectionHandler handles HTTP requests for target collections.
type CollectionHandler struct {
	svc CollectionServicer
}

// NewCollectionHandler constructs a CollectionHandler.
func NewCollectionHandler(svc CollectionServicer) *CollectionHandler {
	return &CollectionHandler{svc: svc}
}

// NewCollectionRouter returns an http.Handler with all collection sub-routes.
func NewCollectionRouter(h *CollectionHandler) http.Handler {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

type collectionRequest struct {
	Name      string   `json:"name"`
	TargetIDs []string `json:"targetIds"`
}

func mapCollectionServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "collection not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
	}
}

// Create handles POST /api/v1/collections
func (h *CollectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	var req collectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return
	}

	c, err := h.svc.Create(r.Context(), actor, req.Name, req.TargetIDs)
	if err != nil {
		mapCollectionServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, r, c)
}

// List handles GET /api/v1/collections
func (h *CollectionHandler) List(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	result, err := h.svc.List(r.Context(), actor)
	if err != nil {
		mapCollectionServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, map[string]any{"items": result})
}

// Get handles GET /api/v1/collections/{id}
func (h *CollectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")
	c, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		mapCollectionServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, c)
}

// Update handles PUT /api/v1/collections/{id}
func (h *CollectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")

	var req collectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return
	}

	c, err := h.svc.Update(r.Context(), actor, id, req.Name, req.TargetIDs)
	if err != nil {
		mapCollectionServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, c)
}

// Delete handles DELETE /api/v1/collections/:id
func (h *CollectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), actor, id); err != nil {
		mapCollectionServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
