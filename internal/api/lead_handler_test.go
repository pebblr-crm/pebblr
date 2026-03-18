package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub LeadService ---

type stubLeadSvc struct{}

func (s *stubLeadSvc) Create(_ context.Context, actor *domain.User, lead *domain.Lead) (*domain.Lead, error) {
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	lead.ID = "lead-new"
	lead.Status = domain.LeadStatusNew
	return lead, nil
}

func (s *stubLeadSvc) Get(_ context.Context, _ *domain.User, id string) (*domain.Lead, error) {
	if id == "lead-1" {
		return &domain.Lead{ID: "lead-1", Title: "Stub Lead", AssigneeID: "rep-1", TeamID: "team-1"}, nil
	}
	return nil, store.ErrNotFound
}

func (s *stubLeadSvc) List(_ context.Context, _ *domain.User, _ store.LeadFilter, _, _ int) (*store.LeadPage, error) {
	return &store.LeadPage{Leads: []*domain.Lead{}, Total: 0, Page: 1, Limit: 20}, nil
}

func (s *stubLeadSvc) Update(_ context.Context, _ *domain.User, lead *domain.Lead) (*domain.Lead, error) {
	if lead.ID == "missing" {
		return nil, store.ErrNotFound
	}
	return lead, nil
}

func (s *stubLeadSvc) Delete(_ context.Context, actor *domain.User, id string) error {
	if actor.Role == domain.RoleRep {
		return service.ErrForbidden
	}
	if id == "missing" {
		return store.ErrNotFound
	}
	return nil
}

func (s *stubLeadSvc) PatchStatus(_ context.Context, _ *domain.User, _ string, status domain.LeadStatus) (*domain.Lead, error) {
	if !status.Valid() {
		return nil, service.ErrInvalidInput
	}
	return &domain.Lead{ID: "lead-1", Status: status}, nil
}

// --- helpers ---

func injectUser(user *domain.User, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := rbac.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func testAdminUser() *domain.User {
	return &domain.User{ID: "admin-1", Role: domain.RoleAdmin, TeamIDs: []string{"team-1"}}
}

func testRepUser() *domain.User {
	return &domain.User{ID: "rep-1", Role: domain.RoleRep, TeamIDs: []string{"team-1"}}
}

func newTestLeadRouter(user *domain.User) http.Handler {
	h := api.NewLeadHandler(&stubLeadSvc{})
	router := api.NewLeadRouter(h)
	if user != nil {
		return injectUser(user, router)
	}
	return router
}

// --- tests ---

func TestLeadList_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestLeadRouter(testAdminUser()).ServeHTTP(w, req)

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

func TestLeadList_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestLeadRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadCreate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"title": "Test Lead", "teamId": "team-1", "customerId": "cust-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestLeadRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadCreate_RepForbidden(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"title": "Test Lead", "teamId": "team-1", "customerId": "cust-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestLeadRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadCreate_MissingTitle_Returns400(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"teamId": "team-1", "customerId": "cust-1"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestLeadRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadGet_Found(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/lead-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestLeadRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadGet_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/missing-lead", http.NoBody)
	w := httptest.NewRecorder()
	newTestLeadRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadDelete_AdminSucceeds(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, "/lead-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestLeadRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadDelete_RepForbidden(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, "/lead-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestLeadRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadPatchStatus_ValidStatus(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"status": "in_progress"})
	req := httptest.NewRequest(http.MethodPatch, "/lead-1/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestLeadRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLeadPatchStatus_InvalidStatus_Returns400(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"status": "bogus"})
	req := httptest.NewRequest(http.MethodPatch, "/lead-1/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestLeadRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
