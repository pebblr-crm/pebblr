package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// ActivityServicer is the interface the ActivityHandler depends on for business logic.
type ActivityServicer interface {
	Create(ctx context.Context, actor *domain.User, activity *domain.Activity) (*domain.Activity, error)
	Get(ctx context.Context, actor *domain.User, id string) (*domain.Activity, error)
	List(ctx context.Context, actor *domain.User, filter store.ActivityFilter, page, limit int) (*store.ActivityPage, error)
	Update(ctx context.Context, actor *domain.User, id string, activity *domain.Activity) (*domain.Activity, error)
	Delete(ctx context.Context, actor *domain.User, id string) error
	Submit(ctx context.Context, actor *domain.User, id string) (*domain.Activity, error)
	PatchStatus(ctx context.Context, actor *domain.User, id, newStatus string) (*domain.Activity, error)
}

// ActivityHandler handles HTTP requests for activity CRUD operations.
type ActivityHandler struct {
	svc ActivityServicer
}

// NewActivityHandler constructs an ActivityHandler backed by the given service.
func NewActivityHandler(svc ActivityServicer) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

// NewActivityRouter returns an http.Handler with all activity sub-routes mounted.
func NewActivityRouter(h *ActivityHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/submit", h.Submit)
	r.Patch("/{id}/status", h.PatchStatus)
	return r
}

type activityRequest struct {
	ActivityType  string         `json:"activityType"`
	Status        string         `json:"status"`
	DueDate       string         `json:"dueDate"`
	Duration      string         `json:"duration"`
	Routing       string         `json:"routing"`
	Fields        map[string]any `json:"fields"`
	TargetID      string         `json:"targetId"`
	JointVisitUID string         `json:"jointVisitUserId"`
}

type activityResponse struct {
	Activity *domain.Activity `json:"activity"`
}

type activityListResponse struct {
	Items []*domain.Activity `json:"items"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

type statusPatchRequest struct {
	Status string `json:"status"`
}

type validationErrorResponse struct {
	Error  errorDetail          `json:"error"`
	Fields []config.FieldError  `json:"fields,omitempty"`
}

func mapActivityServiceError(w http.ResponseWriter, err error) {
	var ve *service.ValidationErrors
	if errors.As(err, &ve) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(validationErrorResponse{
			Error:  errorDetail{Code: "VALIDATION_ERROR", Message: "validation failed"},
			Fields: ve.Errors,
		})
		return
	}
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "activity not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	case errors.Is(err, service.ErrSubmitted):
		writeError(w, http.StatusConflict, "CONFLICT", "activity is submitted and locked")
	case errors.Is(err, service.ErrMaxActivities):
		writeError(w, http.StatusConflict, "CONFLICT", "maximum activities per day reached")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	}
}

// List handles GET /api/v1/activities
func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
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

	var filter store.ActivityFilter
	if v := r.URL.Query().Get("activityType"); v != "" {
		filter.ActivityType = &v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filter.Status = &v
	}
	if v := r.URL.Query().Get("creatorId"); v != "" {
		filter.CreatorID = &v
	}
	if v := r.URL.Query().Get("targetId"); v != "" {
		filter.TargetID = &v
	}
	if v := r.URL.Query().Get("teamId"); v != "" {
		filter.TeamID = &v
	}
	if v := r.URL.Query().Get("dateFrom"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.DateFrom = &t
		}
	}
	if v := r.URL.Query().Get("dateTo"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.DateTo = &t
		}
	}

	result, err := h.svc.List(r.Context(), actor, filter, page, limit)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	activities := result.Activities
	if activities == nil {
		activities = []*domain.Activity{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(activityListResponse{
		Items: activities,
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})
}

// Create handles POST /api/v1/activities
func (h *ActivityHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	var req activityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.ActivityType == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "activityType is required")
		return
	}
	if req.DueDate == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "dueDate is required")
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "dueDate must be in YYYY-MM-DD format")
		return
	}

	fields := req.Fields
	if fields == nil {
		fields = map[string]any{}
	}

	activity := &domain.Activity{
		ActivityType:  req.ActivityType,
		Status:        req.Status,
		DueDate:       dueDate,
		Duration:      req.Duration,
		Routing:       req.Routing,
		Fields:        fields,
		TargetID:      req.TargetID,
		JointVisitUID: req.JointVisitUID,
	}

	created, err := h.svc.Create(r.Context(), actor, activity)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(activityResponse{Activity: created})
}

// Get handles GET /api/v1/activities/{id}
func (h *ActivityHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	activity, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(activityResponse{Activity: activity})
}

// Update handles PUT /api/v1/activities/{id}
func (h *ActivityHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")

	var req activityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.ActivityType == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "activityType is required")
		return
	}
	if req.DueDate == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "dueDate is required")
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "dueDate must be in YYYY-MM-DD format")
		return
	}

	fields := req.Fields
	if fields == nil {
		fields = map[string]any{}
	}

	activity := &domain.Activity{
		ActivityType:  req.ActivityType,
		Status:        req.Status,
		DueDate:       dueDate,
		Duration:      req.Duration,
		Routing:       req.Routing,
		Fields:        fields,
		TargetID:      req.TargetID,
		JointVisitUID: req.JointVisitUID,
	}

	updated, err := h.svc.Update(r.Context(), actor, id, activity)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(activityResponse{Activity: updated})
}

// Delete handles DELETE /api/v1/activities/{id}
func (h *ActivityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), actor, id); err != nil {
		mapActivityServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Submit handles POST /api/v1/activities/{id}/submit
func (h *ActivityHandler) Submit(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	activity, err := h.svc.Submit(r.Context(), actor, id)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(activityResponse{Activity: activity})
}

// PatchStatus handles PATCH /api/v1/activities/{id}/status
func (h *ActivityHandler) PatchStatus(w http.ResponseWriter, r *http.Request) {
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

	if req.Status == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "status is required")
		return
	}

	activity, err := h.svc.PatchStatus(r.Context(), actor, id, req.Status)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(activityResponse{Activity: activity})
}
