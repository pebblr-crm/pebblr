package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
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
	PartialUpdate(ctx context.Context, actor *domain.User, id string, patch *domain.ActivityPatch) (*domain.Activity, error)
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
	r.Post("/batch", h.BatchCreate)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Patch("/{id}", h.Patch)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/submit", h.Submit)
	r.Patch("/{id}/status", h.PatchStatus)
	return r
}

type activityRequest struct {
	ActivityType string         `json:"activityType"`
	Label        string         `json:"label"`
	Status       string         `json:"status"`
	DueDate      string         `json:"dueDate"`
	Duration     string         `json:"duration"`
	Routing      string         `json:"routing"`
	Fields       map[string]any `json:"fields"`
	TargetID     string         `json:"targetId"`
}

// hoistJointVisitUID extracts "joint_visit_user_id" from a fields map and
// returns the value. The key is removed from the map so it is not stored in
// the JSONB column — the real DB column is used instead.
func hoistJointVisitUID(fields map[string]any) string {
	v, ok := fields["joint_visit_user_id"]
	if !ok {
		return ""
	}
	delete(fields, "joint_visit_user_id")
	s, _ := v.(string)
	return s
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

// prepareActivities injects hoisted column values back into Fields for
// all activities before they are serialized to JSON.
func prepareActivities(activities ...*domain.Activity) {
	for _, a := range activities {
		if a != nil {
			a.PrepareForResponse()
		}
	}
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
		if encErr := json.NewEncoder(w).Encode(validationErrorResponse{
			Error:  errorDetail{Code: "VALIDATION_ERROR", Message: "validation failed"},
			Fields: ve.Errors,
		}); encErr != nil {
			slog.Default().Error("failed to encode validation error response", "err", encErr)
		}
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
	case errors.Is(err, service.ErrBlockedDay):
		writeError(w, http.StatusConflict, "CONFLICT", "day is blocked by a non-field activity")
	case errors.Is(err, service.ErrTargetRequired):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Please select a target before saving")
	case errors.Is(err, service.ErrInvalidJointVisitor):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid joint visit user")
	case errors.Is(err, service.ErrStatusNotSubmittable):
		writeError(w, http.StatusConflict, "CONFLICT", "set status to completed or cancelled before submitting")
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
	prepareActivities(activities...)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityListResponse{
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
		Label:         req.Label,
		Status:        req.Status,
		DueDate:       dueDate,
		Duration:      req.Duration,
		Routing:       req.Routing,
		Fields:        fields,
		TargetID:      req.TargetID,
		JointVisitUID: hoistJointVisitUID(fields),
	}

	created, err := h.svc.Create(r.Context(), actor, activity)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	prepareActivities(created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, r, activityResponse{Activity: created})
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

	prepareActivities(activity)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: activity})
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
		Label:         req.Label,
		Status:        req.Status,
		DueDate:       dueDate,
		Duration:      req.Duration,
		Routing:       req.Routing,
		Fields:        fields,
		TargetID:      req.TargetID,
		JointVisitUID: hoistJointVisitUID(fields),
	}

	updated, err := h.svc.Update(r.Context(), actor, id, activity)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	prepareActivities(updated)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: updated})
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

	prepareActivities(activity)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: activity})
}

// Patch handles PATCH /api/v1/activities/{id} with server-side apply semantics.
// Only fields present in the request body are updated; absent fields are left untouched.
// When the "fields" key is present, its sub-keys are merged individually.
func (h *ActivityHandler) Patch(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	patch := &domain.ActivityPatch{}

	if v, ok := raw["status"]; ok {
		if string(v) == "null" {
			s := ""
			patch.Status = &s
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid status value")
				return
			}
			patch.Status = &s
		}
	}

	if v, ok := raw["dueDate"]; ok {
		if string(v) == "null" {
			zero := time.Time{}
			patch.DueDate = &zero
		} else {
			var ds string
			if err := json.Unmarshal(v, &ds); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid dueDate value")
				return
			}
			t, err := time.Parse("2006-01-02", ds)
			if err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "dueDate must be in YYYY-MM-DD format")
				return
			}
			patch.DueDate = &t
		}
	}

	if v, ok := raw["duration"]; ok {
		if string(v) == "null" {
			s := ""
			patch.Duration = &s
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid duration value")
				return
			}
			patch.Duration = &s
		}
	}

	if v, ok := raw["routing"]; ok {
		if string(v) == "null" {
			s := ""
			patch.Routing = &s
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid routing value")
				return
			}
			patch.Routing = &s
		}
	}

	if v, ok := raw["fields"]; ok {
		patch.FieldsPresent = true
		if string(v) != "null" {
			if err := json.Unmarshal(v, &patch.Fields); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid fields value")
				return
			}
		}
		// Hoist joint_visit_user_id from fields into the dedicated column.
		if jv, has := patch.Fields["joint_visit_user_id"]; has {
			delete(patch.Fields, "joint_visit_user_id")
			if jv == nil {
				s := ""
				patch.JointVisitUID = &s
			} else if s, ok := jv.(string); ok {
				patch.JointVisitUID = &s
			}
		}
	}

	if v, ok := raw["targetId"]; ok {
		if string(v) == "null" {
			s := ""
			patch.TargetID = &s
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid targetId value")
				return
			}
			patch.TargetID = &s
		}
	}

	updated, err := h.svc.PartialUpdate(r.Context(), actor, id, patch)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	prepareActivities(updated)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: updated})
}

// BatchCreate handles POST /api/v1/activities/batch
// Creates multiple visit activities from a list of {targetId, dueDate} pairs.
func (h *ActivityHandler) BatchCreate(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	var req struct {
		Items []struct {
			TargetID string `json:"targetId"`
			DueDate  string `json:"dueDate"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "items is required")
		return
	}

	var created []*domain.Activity
	var batchErrors []map[string]string

	for _, item := range req.Items {
		dueDate, err := time.Parse("2006-01-02", item.DueDate)
		if err != nil {
			batchErrors = append(batchErrors, map[string]string{"targetId": item.TargetID, "error": "invalid date format"})
			continue
		}
		activity := &domain.Activity{
			ActivityType: "visit",
			Status:       "",
			DueDate:      dueDate,
			Fields:       map[string]any{"visit_type": "f2f"},
			TargetID:     item.TargetID,
		}
		result, err := h.svc.Create(r.Context(), actor, activity)
		if err != nil {
			batchErrors = append(batchErrors, map[string]string{"targetId": item.TargetID, "error": err.Error()})
			continue
		}
		prepareActivities(result)
		created = append(created, result)
	}

	if created == nil {
		created = []*domain.Activity{}
	}

	w.Header().Set("Content-Type", "application/json")
	status := http.StatusCreated
	if len(batchErrors) > 0 && len(created) == 0 {
		status = http.StatusBadRequest
	} else if len(batchErrors) > 0 {
		status = http.StatusMultiStatus
	}
	w.WriteHeader(status)
	writeJSON(w, r, map[string]any{
		"created": created,
		"errors":  batchErrors,
	})
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

	prepareActivities(activity)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: activity})
}
