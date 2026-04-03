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

const testTeamID = "team-1"

// --- stub TerritoryService ---

type stubTerritorySvc struct {
	territories []*domain.Territory
}

func (s *stubTerritorySvc) List(_ context.Context, _ *domain.User) ([]*domain.Territory, error) {
	return s.territories, nil
}

func (s *stubTerritorySvc) Get(_ context.Context, _ *domain.User, id string) (*domain.Territory, error) {
	for _, t := range s.territories {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, store.ErrNotFound
}

func (s *stubTerritorySvc) Create(_ context.Context, actor *domain.User, t *domain.Territory) (*domain.Territory, error) {
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	t.ID = "new-territory-id"
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	s.territories = append(s.territories, t)
	return t, nil
}

func (s *stubTerritorySvc) Update(_ context.Context, actor *domain.User, t *domain.Territory) (*domain.Territory, error) {
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	for i, existing := range s.territories {
		if existing.ID == t.ID {
			t.CreatedAt = existing.CreatedAt
			t.UpdatedAt = time.Now()
			s.territories[i] = t
			return t, nil
		}
	}
	return nil, store.ErrNotFound
}

func (s *stubTerritorySvc) Delete(_ context.Context, actor *domain.User, id string) error {
	if actor.Role == domain.RoleRep {
		return service.ErrForbidden
	}
	for i, t := range s.territories {
		if t.ID == id {
			s.territories = append(s.territories[:i], s.territories[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func defaultTerritories() []*domain.Territory {
	return []*domain.Territory{
		{ID: "t-1", Name: "Bucharest North", TeamID: testTeamID, Region: "Bucharest"},
		{ID: "t-2", Name: "Bucharest South", TeamID: testTeamID, Region: "Bucharest"},
	}
}

func newTestTerritoryRouter(user *domain.User) http.Handler {
	h := api.NewTerritoryHandler(&stubTerritorySvc{territories: defaultTerritories()})
	router := api.NewTerritoryRouter(h)
	if user != nil {
		return injectUser(user, router)
	}
	return router
}

// --- tests ---

func TestTerritoryList_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatal("expected 'items' array in response")
	}
	if len(items) != 2 {
		t.Errorf("expected 2 territories, got %d", len(items))
	}
}

func TestTerritoryList_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTerritoryGet_Found(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/t-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryGet_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/missing", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTerritoryCreate_Admin_Returns201(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":   "Cluj Region",
		"teamId": testTeamID,
		"region": "Cluj",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryCreate_Rep_Returns403(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":   "Test",
		"teamId": testTeamID,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryDelete_Admin_Returns204(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, "/t-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryDelete_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, "/missing", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTerritoryDelete_Rep_Returns403(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, "/t-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryUpdate_Admin_ReturnsOK(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":   "Bucharest North Updated",
		"teamId": testTeamID,
		"region": "Bucharest",
	})
	req := httptest.NewRequest(http.MethodPut, "/t-1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryUpdate_Rep_Returns403(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":   "Test",
		"teamId": testTeamID,
	})
	req := httptest.NewRequest(http.MethodPut, "/t-1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryUpdate_NotFound(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":   "Test",
		"teamId": testTeamID,
	})
	req := httptest.NewRequest(http.MethodPut, "/missing", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryCreate_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryUpdate_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPut, "/t-1", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTerritoryGet_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/t-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTerritoryCreate_NoAuth(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"name": "Test"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTerritoryUpdate_NoAuth(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"name": "Test"})
	req := httptest.NewRequest(http.MethodPut, "/t-1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestTerritoryRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTerritoryDelete_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, "/t-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestTerritoryRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
