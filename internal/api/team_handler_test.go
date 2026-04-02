package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub TeamService ---

type stubTeamSvc struct{}

func (s *stubTeamSvc) List(_ context.Context, _ *domain.User) ([]*domain.Team, error) {
	return []*domain.Team{
		{ID: testTeamID, Name: "Alpha Team", ManagerID: "admin-1"},
	}, nil
}

func (s *stubTeamSvc) Get(_ context.Context, _ *domain.User, id string) (*domain.Team, error) {
	if id == testTeamID {
		return &domain.Team{ID: testTeamID, Name: "Alpha Team", ManagerID: "admin-1"}, nil
	}
	return nil, store.ErrNotFound
}

func (s *stubTeamSvc) ListMembers(_ context.Context, _ *domain.User, _ string) ([]*domain.User, error) {
	return []*domain.User{
		{ID: "rep-1", Name: "Alice Rep", Role: domain.RoleRep},
	}, nil
}

// --- helpers ---

func newTestTeamRouter(user *domain.User) http.Handler {
	h := api.NewTeamHandler(&stubTeamSvc{})
	router := api.NewTeamRouter(h)
	if user != nil {
		return injectUser(user, router)
	}
	return router
}

// --- tests ---

func TestTeamList_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestTeamRouter(testAdminUser()).ServeHTTP(w, req)

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

func TestTeamList_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestTeamRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTeamGet_Found(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/team-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestTeamRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["team"]; !ok {
		t.Error("expected 'team' key in response")
	}
	if _, ok := resp["members"]; !ok {
		t.Error("expected 'members' key in response")
	}
}

func TestTeamGet_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/missing-team", http.NoBody)
	w := httptest.NewRecorder()
	newTestTeamRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
