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

// --- stub UserService ---

type stubUserSvc struct{}

func (s *stubUserSvc) List(_ context.Context) ([]*domain.User, error) {
	return []*domain.User{
		{ID: "user-1", Name: "Alice Admin", Role: domain.RoleAdmin},
	}, nil
}

func (s *stubUserSvc) Get(_ context.Context, id string) (*domain.User, error) {
	if id == "user-1" {
		return &domain.User{ID: "user-1", Name: "Alice Admin", Role: domain.RoleAdmin}, nil
	}
	return nil, store.ErrNotFound
}

// --- helpers ---

func newTestUserRouter(user *domain.User) http.Handler {
	h := api.NewUserHandler(&stubUserSvc{})
	router := api.NewUserRouter(h)
	if user != nil {
		return injectUser(user, router)
	}
	return router
}

// --- tests ---

func TestUserList_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestUserRouter(testAdminUser()).ServeHTTP(w, req)

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

func TestUserList_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestUserRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUserGet_Found(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/user-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestUserRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["user"]; !ok {
		t.Error("expected 'user' key in response")
	}
}

func TestUserGet_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/missing-user", http.NoBody)
	w := httptest.NewRecorder()
	newTestUserRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
