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

// TargetServicer is the interface the TargetHandler depends on for business logic.
type TargetServicer interface {
	Create(ctx context.Context, actor *domain.User, target *domain.Target) (*domain.Target, error)
	Get(ctx context.Context, actor *domain.User, id string) (*domain.Target, error)
	List(ctx context.Context, actor *domain.User, filter store.TargetFilter, page, limit int) (*store.TargetPage, error)
	Update(ctx context.Context, actor *domain.User, target *domain.Target) (*domain.Target, error)
	Import(ctx context.Context, actor *domain.User, targets []*domain.Target) (*store.ImportResult, error)
	VisitStatus(ctx context.Context, actor *domain.User) ([]store.TargetVisitStatus, error)
}

// TargetHandler handles HTTP requests for target CRUD operations.
type TargetHandler struct {
	svc TargetServicer
}

// NewTargetHandler constructs a TargetHandler backed by the given service.
func NewTargetHandler(svc TargetServicer) *TargetHandler {
	return &TargetHandler{svc: svc}
}

// NewTargetRouter returns an http.Handler with all target sub-routes mounted.
func NewTargetRouter(h *TargetHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/visit-status", h.VisitStatus)
	r.Post("/", h.Create)
	r.Post("/import", h.Import)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	return r
}

type targetRequest struct {
	TargetType string         `json:"targetType"`
	Name       string         `json:"name"`
	Fields     map[string]any `json:"fields"`
	AssigneeID string         `json:"assigneeId"`
	TeamID     string         `json:"teamId"`
}

type targetResponse struct {
	Target *domain.Target `json:"target"`
}

type targetListResponse struct {
	Items []*domain.Target `json:"items"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

func mapTargetServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "target not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	}
}

// List handles GET /api/v1/targets
func (h *TargetHandler) List(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	var filter store.TargetFilter
	if t := r.URL.Query().Get("type"); t != "" {
		filter.TargetType = &t
	}
	if a := r.URL.Query().Get("assignee"); a != "" {
		filter.AssigneeID = &a
	}
	if q := r.URL.Query().Get("q"); q != "" {
		filter.Query = &q
	}

	result, err := h.svc.List(r.Context(), actor, filter, page, limit)
	if err != nil {
		mapTargetServiceError(w, err)
		return
	}

	targets := result.Targets
	if targets == nil {
		targets = []*domain.Target{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, targetListResponse{
		Items: targets,
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})
}

// Create handles POST /api/v1/targets
func (h *TargetHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	var req targetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required")
		return
	}

	fields := req.Fields
	if fields == nil {
		fields = map[string]any{}
	}

	target := &domain.Target{
		TargetType: req.TargetType,
		Name:       req.Name,
		Fields:     fields,
		AssigneeID: req.AssigneeID,
		TeamID:     req.TeamID,
	}

	created, err := h.svc.Create(r.Context(), actor, target)
	if err != nil {
		mapTargetServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, r, targetResponse{Target: created})
}

// Get handles GET /api/v1/targets/{id}
func (h *TargetHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	target, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		mapTargetServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, targetResponse{Target: target})
}

// Update handles PUT /api/v1/targets/{id}
func (h *TargetHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")

	var req targetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required")
		return
	}

	fields := req.Fields
	if fields == nil {
		fields = map[string]any{}
	}

	target := &domain.Target{
		ID:         id,
		TargetType: req.TargetType,
		Name:       req.Name,
		Fields:     fields,
		AssigneeID: req.AssigneeID,
		TeamID:     req.TeamID,
	}

	updated, err := h.svc.Update(r.Context(), actor, target)
	if err != nil {
		mapTargetServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, targetResponse{Target: updated})
}

type importTargetItem struct {
	ExternalID string         `json:"externalId"`
	TargetType string         `json:"targetType"`
	Name       string         `json:"name"`
	Fields     map[string]any `json:"fields"`
	AssigneeID string         `json:"assigneeId"`
	TeamID     string         `json:"teamId"`
}

type importRequest struct {
	Targets []importTargetItem `json:"targets"`
}

type importResponse struct {
	Created int              `json:"created"`
	Updated int              `json:"updated"`
	Targets []*domain.Target `json:"targets"`
}

// Import handles POST /api/v1/targets/import
func (h *TargetHandler) Import(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	var req importRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if len(req.Targets) == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "targets array is required and must not be empty")
		return
	}

	targets := make([]*domain.Target, len(req.Targets))
	for i, item := range req.Targets {
		fields := item.Fields
		if fields == nil {
			fields = map[string]any{}
		}
		targets[i] = &domain.Target{
			ExternalID: item.ExternalID,
			TargetType: item.TargetType,
			Name:       item.Name,
			Fields:     fields,
			AssigneeID: item.AssigneeID,
			TeamID:     item.TeamID,
		}
	}

	result, err := h.svc.Import(r.Context(), actor, targets)
	if err != nil {
		mapTargetServiceError(w, err)
		return
	}

	imported := result.Imported
	if imported == nil {
		imported = []*domain.Target{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, importResponse{
		Created: result.Created,
		Updated: result.Updated,
		Targets: imported,
	})
}

// VisitStatus handles GET /api/v1/targets/visit-status
func (h *TargetHandler) VisitStatus(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	result, err := h.svc.VisitStatus(r.Context(), actor)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to query visit status")
		return
	}
	if result == nil {
		result = []store.TargetVisitStatus{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, r, map[string]any{"items": result})
}
