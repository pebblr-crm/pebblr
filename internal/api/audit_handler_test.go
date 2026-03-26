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

// --- stub AuditService ---

type stubAuditSvc struct {
	entries []*domain.AuditEntry
}

func (s *stubAuditSvc) List(_ context.Context, actor *domain.User, _ store.AuditFilter) ([]*domain.AuditEntry, int, error) {
	if actor.Role != domain.RoleAdmin {
		return nil, 0, service.ErrForbidden
	}
	return s.entries, len(s.entries), nil
}

func (s *stubAuditSvc) UpdateStatus(_ context.Context, actor *domain.User, id, status string) error {
	if actor.Role != domain.RoleAdmin {
		return service.ErrForbidden
	}
	for _, e := range s.entries {
		if e.ID == id {
			e.Status = status
			return nil
		}
	}
	return store.ErrNotFound
}

func defaultAuditEntries() []*domain.AuditEntry {
	return []*domain.AuditEntry{
		{
			ID:         "audit-1",
			EntityType: "activity",
			EntityID:   "act-1",
			EventType:  "status_changed",
			ActorID:    "rep-1",
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "audit-2",
			EntityType: "target",
			EntityID:   "tgt-1",
			EventType:  "field_updated",
			ActorID:    "rep-1",
			Status:     "accepted",
			CreatedAt:  time.Now(),
		},
	}
}

func newTestAuditRouter(user *domain.User) http.Handler {
	h := api.NewAuditHandler(&stubAuditSvc{entries: defaultAuditEntries()})
	router := api.NewAuditRouter(h)
	if user != nil {
		return injectUser(user, router)
	}
	return router
}

// --- tests ---

func TestAuditList_Admin_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestAuditRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatal("expected 'items' array")
	}
	if len(items) != 2 {
		t.Errorf("expected 2 entries, got %d", len(items))
	}
}

func TestAuditList_Rep_Returns403(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestAuditRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuditList_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestAuditRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuditUpdateStatus_Admin_Returns204(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]string{"status": "accepted"})
	req := httptest.NewRequest(http.MethodPatch, "/audit-1/status", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestAuditRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuditUpdateStatus_Rep_Returns403(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]string{"status": "accepted"})
	req := httptest.NewRequest(http.MethodPatch, "/audit-1/status", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestAuditRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuditUpdateStatus_NotFound(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]string{"status": "accepted"})
	req := httptest.NewRequest(http.MethodPatch, "/missing/status", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestAuditRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuditList_WithFilters(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/?entityType=activity&status=pending&page=1&limit=10", http.NoBody)
	w := httptest.NewRecorder()
	newTestAuditRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
