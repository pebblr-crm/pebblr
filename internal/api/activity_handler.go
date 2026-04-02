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
	CloneWeek(ctx context.Context, actor *domain.User, sourceWeekStart, targetWeekStart time.Time) (*service.CloneWeekResult, error)
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
	r.Post("/clone-week", h.CloneWeek)
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
// all activities before they are serialized to JSON. This is a presentation
// concern — the frontend expects hoisted columns to also appear in the
// dynamic fields map.
func prepareActivities(activities ...*domain.Activity) {
	for _, a := range activities {
		if a == nil {
			continue
		}
		if a.Fields == nil {
			a.Fields = map[string]any{}
		}
		if a.JointVisitUID != "" {
			a.Fields["joint_visit_user_id"] = a.JointVisitUID
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

// safeBatchError maps a service error to a client-safe message for batch responses.
// This prevents leaking internal error details (SQL errors, stack traces) to clients.
func safeBatchError(err error) string {
	switch {
	case errors.Is(err, service.ErrForbidden):
		return "access denied"
	case errors.Is(err, store.ErrNotFound):
		return "target not found"
	case errors.Is(err, service.ErrInvalidInput):
		return "invalid input"
	case errors.Is(err, service.ErrSubmitted):
		return "activity is submitted and locked"
	case errors.Is(err, service.ErrMaxActivities):
		return "maximum activities per day reached"
	case errors.Is(err, service.ErrBlockedDay):
		return "day is blocked by a non-field activity"
	case errors.Is(err, service.ErrTargetRequired):
		return "target is required"
	case errors.Is(err, service.ErrDuplicateActivity):
		return "activity for this target on this date already exists"
	default:
		return "an unexpected error occurred"
	}
}

func mapActivityServiceError(w http.ResponseWriter, err error) {
	var ve *service.ValidationErrors
	if errors.As(err, &ve) {
		w.Header().Set(headerContentType, contentTypeJSON)
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
	case errors.Is(err, service.ErrTargetNotAccessible):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "target not in your visible scope")
	case errors.Is(err, service.ErrStatusNotSubmittable):
		writeError(w, http.StatusConflict, "CONFLICT", "set status to completed or cancelled before submitting")
	case errors.Is(err, service.ErrNoRecoveryBalance):
		writeError(w, http.StatusConflict, "CONFLICT", "no recovery day balance available")
	case errors.Is(err, service.ErrDuplicateActivity):
		writeError(w, http.StatusConflict, "DUPLICATE", "activity for this target on this date already exists")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
	}
}

// parseActivityFilter builds an ActivityFilter from query parameters.
func parseActivityFilter(r *http.Request) store.ActivityFilter {
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
		if t, err := time.Parse(dateFormat, v); err == nil {
			filter.DateFrom = &t
		}
	}
	if v := r.URL.Query().Get("dateTo"); v != "" {
		if t, err := time.Parse(dateFormat, v); err == nil {
			filter.DateTo = &t
		}
	}
	return filter
}

// List handles GET /api/v1/activities
func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
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
	if limit > maxPaginationLimit {
		limit = maxPaginationLimit
	}

	filter := parseActivityFilter(r)

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

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityListResponse{
		Items: activities,
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})
}

// decodeActivityRequest decodes, validates, and converts an activity request body
// into a domain.Activity. Returns nil and writes an error response if validation fails.
func decodeActivityRequest(w http.ResponseWriter, r *http.Request) *domain.Activity {
	var req activityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return nil
	}

	if req.ActivityType == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "activityType is required")
		return nil
	}
	if req.DueDate == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "dueDate is required")
		return nil
	}

	dueDate, err := time.Parse(dateFormat, req.DueDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "dueDate must be in YYYY-MM-DD format")
		return nil
	}

	fields := req.Fields
	if fields == nil {
		fields = map[string]any{}
	}

	return &domain.Activity{
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
}

// Create handles POST /api/v1/activities
func (h *ActivityHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	activity := decodeActivityRequest(w, r)
	if activity == nil {
		return
	}

	created, err := h.svc.Create(r.Context(), actor, activity)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	prepareActivities(created)
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, r, activityResponse{Activity: created})
}

// Get handles GET /api/v1/activities/{id}
func (h *ActivityHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")
	activity, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	prepareActivities(activity)
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: activity})
}

// Update handles PUT /api/v1/activities/{id}
func (h *ActivityHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")

	activity := decodeActivityRequest(w, r)
	if activity == nil {
		return
	}

	updated, err := h.svc.Update(r.Context(), actor, id, activity)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	prepareActivities(updated)
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: updated})
}

// Delete handles DELETE /api/v1/activities/{id}
func (h *ActivityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
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
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")
	activity, err := h.svc.Submit(r.Context(), actor, id)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	prepareActivities(activity)
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: activity})
}

// parseRawString unmarshals a JSON value as a string pointer suitable for a patch field.
// If the raw value is "null", it returns a pointer to an empty string.
// Returns the string pointer and any parsing error message (empty if ok).
func parseRawString(raw json.RawMessage) (val *string, errMsg string) {
	if string(raw) == "null" {
		s := ""
		return &s, ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err.Error()
	}
	return &s, ""
}

// parseRawDueDate unmarshals a JSON value as a time pointer for the dueDate patch field.
// Returns the time pointer and any error message (empty if ok).
func parseRawDueDate(raw json.RawMessage) (val *time.Time, errMsg string) {
	if string(raw) == "null" {
		zero := time.Time{}
		return &zero, ""
	}
	var ds string
	if err := json.Unmarshal(raw, &ds); err != nil {
		return nil, "invalid dueDate value"
	}
	t, err := time.Parse(dateFormat, ds)
	if err != nil {
		return nil, "dueDate must be in YYYY-MM-DD format"
	}
	return &t, ""
}

// parsePatchFields unmarshals and processes the "fields" key from a patch request,
// hoisting joint_visit_user_id into the patch's dedicated column field.
func parsePatchFields(raw json.RawMessage, patch *domain.ActivityPatch) string {
	patch.FieldsPresent = true
	if string(raw) == "null" {
		return ""
	}
	if err := json.Unmarshal(raw, &patch.Fields); err != nil {
		return "invalid fields value"
	}
	// Hoist joint_visit_user_id from fields into the dedicated column.
	jv, has := patch.Fields["joint_visit_user_id"]
	if !has {
		return ""
	}
	delete(patch.Fields, "joint_visit_user_id")
	if jv == nil {
		s := ""
		patch.JointVisitUID = &s
	} else if s, ok := jv.(string); ok {
		patch.JointVisitUID = &s
	}
	return ""
}

// Patch handles PATCH /api/v1/activities/{id} with server-side apply semantics.
// Only fields present in the request body are updated; absent fields are left untouched.
// When the "fields" key is present, its sub-keys are merged individually.
func (h *ActivityHandler) Patch(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return
	}

	patch, errMsg := buildActivityPatch(raw)
	if errMsg != "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errMsg)
		return
	}

	updated, err := h.svc.PartialUpdate(r.Context(), actor, id, patch)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	prepareActivities(updated)
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: updated})
}

// parsePatchStringField extracts a string field from raw JSON and assigns it to dest.
// Returns an error message if parsing fails, empty string on success.
func parsePatchStringField(raw map[string]json.RawMessage, key, label string, dest **string) string {
	v, ok := raw[key]
	if !ok {
		return ""
	}
	s, errMsg := parseRawString(v)
	if errMsg != "" {
		return "invalid " + label + " value"
	}
	*dest = s
	return ""
}

// buildActivityPatch constructs an ActivityPatch from the raw JSON map.
// Returns the patch and an error message (empty if ok).
func buildActivityPatch(raw map[string]json.RawMessage) (result *domain.ActivityPatch, errMsg string) {
	patch := &domain.ActivityPatch{}

	stringFields := []struct {
		key, label string
		dest       **string
	}{
		{"status", "status", &patch.Status},
		{"duration", "duration", &patch.Duration},
		{"routing", "routing", &patch.Routing},
		{"targetId", "targetId", &patch.TargetID},
	}
	for _, sf := range stringFields {
		if errMsg := parsePatchStringField(raw, sf.key, sf.label, sf.dest); errMsg != "" {
			return nil, errMsg
		}
	}

	if v, ok := raw["dueDate"]; ok {
		t, errMsg := parseRawDueDate(v)
		if errMsg != "" {
			return nil, errMsg
		}
		patch.DueDate = t
	}

	if v, ok := raw["fields"]; ok {
		if errMsg := parsePatchFields(v, patch); errMsg != "" {
			return nil, errMsg
		}
	}

	return patch, ""
}

// BatchCreate handles POST /api/v1/activities/batch
// Creates multiple visit activities from a list of {targetId, dueDate} pairs.
func (h *ActivityHandler) BatchCreate(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	var req struct {
		Items []struct {
			TargetID string         `json:"targetId"`
			DueDate  string         `json:"dueDate"`
			Fields   map[string]any `json:"fields,omitempty"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "items is required")
		return
	}
	if len(req.Items) > maxBatchItems {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "batch limited to 100 items per request")
		return
	}

	var created []*domain.Activity
	var batchErrors []map[string]string

	for _, item := range req.Items {
		dueDate, err := time.Parse(dateFormat, item.DueDate)
		if err != nil {
			batchErrors = append(batchErrors, map[string]string{"targetId": item.TargetID, "error": "invalid date format"})
			continue
		}
		fields := map[string]any{"visit_type": "f2f"}
		for k, v := range item.Fields {
			fields[k] = v
		}
		activity := &domain.Activity{
			ActivityType: "visit",
			Status:       "",
			DueDate:      dueDate,
			Fields:       fields,
			TargetID:     item.TargetID,
		}
		result, err := h.svc.Create(r.Context(), actor, activity)
		if err != nil {
			// Map to safe error message -- never leak raw service/DB errors to clients.
			batchErrors = append(batchErrors, map[string]string{"targetId": item.TargetID, "error": safeBatchError(err)})
			continue
		}
		prepareActivities(result)
		created = append(created, result)
	}

	if created == nil {
		created = []*domain.Activity{}
	}

	w.Header().Set(headerContentType, contentTypeJSON)
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

// CloneWeek handles POST /api/v1/activities/clone-week
func (h *ActivityHandler) CloneWeek(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	var req struct {
		SourceWeekStart string `json:"sourceWeekStart"`
		TargetWeekStart string `json:"targetWeekStart"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
		return
	}

	source, err := time.Parse(dateFormat, req.SourceWeekStart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "sourceWeekStart must be YYYY-MM-DD")
		return
	}
	target, err := time.Parse(dateFormat, req.TargetWeekStart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "targetWeekStart must be YYYY-MM-DD")
		return
	}

	result, err := h.svc.CloneWeek(r.Context(), actor, source, target)
	if err != nil {
		mapActivityServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, r, result)
}

// PatchStatus handles PATCH /api/v1/activities/{id}/status
func (h *ActivityHandler) PatchStatus(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")

	var req statusPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", errInvalidRequestBody)
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
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, activityResponse{Activity: activity})
}
