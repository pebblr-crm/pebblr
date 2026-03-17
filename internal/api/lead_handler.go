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

// LeadServicer is the interface the LeadHandler depends on for business logic.
// Defined here so handler tests can provide stubs without a real database.
type LeadServicer interface {
	Create(ctx context.Context, actor *domain.User, lead *domain.Lead) (*domain.Lead, error)
	Get(ctx context.Context, actor *domain.User, id string) (*domain.Lead, error)
	List(ctx context.Context, actor *domain.User, filter store.LeadFilter, page, limit int) (*store.LeadPage, error)
	Update(ctx context.Context, actor *domain.User, lead *domain.Lead) (*domain.Lead, error)
	Delete(ctx context.Context, actor *domain.User, id string) error
	PatchStatus(ctx context.Context, actor *domain.User, id string, status domain.LeadStatus) (*domain.Lead, error)
}

// LeadHandler handles HTTP requests for lead CRUD operations.
type LeadHandler struct {
	svc LeadServicer
}

// NewLeadHandler constructs a LeadHandler backed by the given service.
func NewLeadHandler(svc LeadServicer) *LeadHandler {
	return &LeadHandler{svc: svc}
}

// NewLeadRouter returns an http.Handler with all lead sub-routes mounted.
// Mount at /api/v1/leads in the main router.
func NewLeadRouter(h *LeadHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Patch("/{id}/status", h.PatchStatus)
	return r
}

// leadRequest is the body for create and update operations.
type leadRequest struct {
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	AssigneeID   string              `json:"assignee_id"`
	TeamID       string              `json:"team_id"`
	CustomerID   string              `json:"customer_id"`
	CustomerType domain.CustomerType `json:"customer_type"`
	Status       domain.LeadStatus   `json:"status"`
}

// statusPatchRequest is the body for PATCH /leads/:id/status.
type statusPatchRequest struct {
	Status domain.LeadStatus `json:"status"`
}

// leadResponse is the JSON envelope for a single lead.
type leadResponse struct {
	Lead *domain.Lead `json:"lead"`
}

// leadListResponse is the JSON envelope for a paginated lead list.
type leadListResponse struct {
	Leads []*domain.Lead `json:"leads"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "lead not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	}
}

// List handles GET /api/v1/leads
func (h *LeadHandler) List(w http.ResponseWriter, r *http.Request) {
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

	var filter store.LeadFilter
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.LeadStatus(s)
		if !st.Valid() {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid status filter")
			return
		}
		filter.Status = &st
	}
	if a := r.URL.Query().Get("assignee"); a != "" {
		filter.Assignee = &a
	}
	if t := r.URL.Query().Get("team"); t != "" {
		filter.Team = &t
	}

	result, err := h.svc.List(r.Context(), actor, filter, page, limit)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	leads := result.Leads
	if leads == nil {
		leads = []*domain.Lead{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(leadListResponse{
		Leads: leads,
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})
}

// Create handles POST /api/v1/leads
func (h *LeadHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	var req leadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "title is required")
		return
	}
	if req.TeamID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "team_id is required")
		return
	}
	if req.CustomerID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "customer_id is required")
		return
	}

	lead := &domain.Lead{
		Title:        req.Title,
		Description:  req.Description,
		AssigneeID:   req.AssigneeID,
		TeamID:       req.TeamID,
		CustomerID:   req.CustomerID,
		CustomerType: req.CustomerType,
	}

	created, err := h.svc.Create(r.Context(), actor, lead)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(leadResponse{Lead: created})
}

// Get handles GET /api/v1/leads/{id}
func (h *LeadHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	lead, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(leadResponse{Lead: lead})
}

// Update handles PUT /api/v1/leads/{id}
func (h *LeadHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")

	var req leadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "title is required")
		return
	}

	lead := &domain.Lead{
		ID:           id,
		Title:        req.Title,
		Description:  req.Description,
		Status:       req.Status,
		AssigneeID:   req.AssigneeID,
		TeamID:       req.TeamID,
		CustomerID:   req.CustomerID,
		CustomerType: req.CustomerType,
	}

	updated, err := h.svc.Update(r.Context(), actor, lead)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(leadResponse{Lead: updated})
}

// Delete handles DELETE /api/v1/leads/{id}
func (h *LeadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), actor, id); err != nil {
		mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PatchStatus handles PATCH /api/v1/leads/{id}/status
func (h *LeadHandler) PatchStatus(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")

	var req statusPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	updated, err := h.svc.PatchStatus(r.Context(), actor, id, req.Status)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(leadResponse{Lead: updated})
}
