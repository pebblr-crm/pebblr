package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub CalendarEventService ---

type stubCalendarEventSvc struct{}

func (s *stubCalendarEventSvc) Get(_ context.Context, id string) (*domain.CalendarEvent, error) {
	if id == "evt-1" {
		return &domain.CalendarEvent{ID: "evt-1", Title: "Stub Event"}, nil
	}
	return nil, store.ErrNotFound
}

func (s *stubCalendarEventSvc) List(_ context.Context, _ store.CalendarEventFilter, _, _ int) (*store.CalendarEventPage, error) {
	return &store.CalendarEventPage{Events: []*domain.CalendarEvent{}, Total: 0, Page: 1, Limit: 20}, nil
}

func (s *stubCalendarEventSvc) Create(_ context.Context, actor *domain.User, event *domain.CalendarEvent) (*domain.CalendarEvent, error) {
	event.ID = "evt-new"
	event.CreatorID = actor.ID
	return event, nil
}

func (s *stubCalendarEventSvc) Update(_ context.Context, _ *domain.User, event *domain.CalendarEvent) (*domain.CalendarEvent, error) {
	if event.ID == "missing" {
		return nil, store.ErrNotFound
	}
	return event, nil
}

func (s *stubCalendarEventSvc) Delete(_ context.Context, actor *domain.User, id string) error {
	if id == "missing" {
		return store.ErrNotFound
	}
	if actor.Role == domain.RoleRep {
		return service.ErrForbidden
	}
	return nil
}

// --- helpers ---

func newTestCalendarEventRouter(user *domain.User) http.Handler {
	h := api.NewCalendarEventHandler(&stubCalendarEventSvc{})
	router := api.NewCalendarEventRouter(h)
	if user != nil {
		return injectUser(user, router)
	}
	return router
}

// --- tests ---

func TestCalendarEventList_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["items"]; !ok {
		t.Error("expected 'items' key in response")
	}
}

func TestCalendarEventList_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCalendarEventGet_Found(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/evt-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCalendarEventGet_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/missing-evt", http.NoBody)
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCalendarEventCreate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"title":     "Team Sync",
		"eventType": "sync",
		"startTime": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCalendarEventCreate_MissingTitle_Returns400(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"eventType": "sync",
		"startTime": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCalendarEventCreate_InvalidEventType_Returns400(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"title":     "Meeting",
		"eventType": "bogus",
		"startTime": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCalendarEventDelete_AdminSucceeds(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, "/evt-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCalendarEventDelete_RepForbidden(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, "/evt-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestCalendarEventRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}
