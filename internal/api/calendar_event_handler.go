package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// CalendarEventServicer is the interface the CalendarEventHandler depends on for business logic.
type CalendarEventServicer interface {
	Get(ctx context.Context, id string) (*domain.CalendarEvent, error)
	List(ctx context.Context, filter store.CalendarEventFilter, page, limit int) (*store.CalendarEventPage, error)
	Create(ctx context.Context, actor *domain.User, event *domain.CalendarEvent) (*domain.CalendarEvent, error)
	Update(ctx context.Context, actor *domain.User, event *domain.CalendarEvent) (*domain.CalendarEvent, error)
	Delete(ctx context.Context, actor *domain.User, id string) error
}

// CalendarEventHandler handles HTTP requests for calendar event CRUD operations.
type CalendarEventHandler struct {
	svc CalendarEventServicer
}

// NewCalendarEventHandler constructs a CalendarEventHandler backed by the given service.
func NewCalendarEventHandler(svc CalendarEventServicer) *CalendarEventHandler {
	return &CalendarEventHandler{svc: svc}
}

// NewCalendarEventRouter returns an http.Handler with all calendar event sub-routes mounted.
// Mount at /api/v1/events in the main router.
func NewCalendarEventRouter(h *CalendarEventHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// calendarEventRequest is the body for create and update operations.
type calendarEventRequest struct {
	Title     string                   `json:"title"`
	EventType domain.CalendarEventType `json:"eventType"`
	StartTime time.Time                `json:"startTime"`
	EndTime   *time.Time               `json:"endTime"`
	Client    string                   `json:"client"`
	TeamID    string                   `json:"teamId"`
}

// calendarEventResponse is the JSON envelope for a single calendar event.
type calendarEventResponse struct {
	Event *domain.CalendarEvent `json:"event"`
}

// calendarEventListResponse is the JSON envelope for a paginated calendar event list.
type calendarEventListResponse struct {
	Items []*domain.CalendarEvent `json:"items"`
	Total int                     `json:"total"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
}

func mapCalendarEventServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "calendar event not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	}
}

// List handles GET /api/v1/events
func (h *CalendarEventHandler) List(w http.ResponseWriter, r *http.Request) {
	_, err := rbac.UserFromContext(r.Context())
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

	var filter store.CalendarEventFilter
	if c := r.URL.Query().Get("creator"); c != "" {
		filter.CreatorID = &c
	}
	if t := r.URL.Query().Get("team"); t != "" {
		filter.TeamID = &t
	}

	result, err := h.svc.List(r.Context(), filter, page, limit)
	if err != nil {
		mapCalendarEventServiceError(w, err)
		return
	}

	evts := result.Events
	if evts == nil {
		evts = []*domain.CalendarEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(calendarEventListResponse{
		Items: evts,
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})
}

// Get handles GET /api/v1/events/{id}
func (h *CalendarEventHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	evt, err := h.svc.Get(r.Context(), id)
	if err != nil {
		mapCalendarEventServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(calendarEventResponse{Event: evt})
}

// Create handles POST /api/v1/events
func (h *CalendarEventHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	var req calendarEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "title is required")
		return
	}
	if !req.EventType.Valid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid eventType")
		return
	}
	if req.StartTime.IsZero() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "startTime is required")
		return
	}

	evt := &domain.CalendarEvent{
		Title:     req.Title,
		EventType: req.EventType,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Client:    req.Client,
		TeamID:    req.TeamID,
		CreatorID: actor.ID,
	}

	created, err := h.svc.Create(r.Context(), actor, evt)
	if err != nil {
		mapCalendarEventServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(calendarEventResponse{Event: created})
}

// Update handles PUT /api/v1/events/{id}
func (h *CalendarEventHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")

	var req calendarEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "title is required")
		return
	}
	if !req.EventType.Valid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid eventType")
		return
	}

	evt := &domain.CalendarEvent{
		ID:        id,
		Title:     req.Title,
		EventType: req.EventType,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Client:    req.Client,
		TeamID:    req.TeamID,
		CreatorID: actor.ID,
	}

	updated, err := h.svc.Update(r.Context(), actor, evt)
	if err != nil {
		mapCalendarEventServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(calendarEventResponse{Event: updated})
}

// Delete handles DELETE /api/v1/events/{id}
func (h *CalendarEventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), actor, id); err != nil {
		mapCalendarEventServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
