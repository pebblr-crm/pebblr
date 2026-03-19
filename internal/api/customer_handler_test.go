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
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub CustomerService ---

type stubCustomerSvc struct{}

func (s *stubCustomerSvc) Create(_ context.Context, actor *domain.User, c *domain.Customer) (*domain.Customer, error) {
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	c.ID = "cust-new"
	return c, nil
}

func (s *stubCustomerSvc) Get(_ context.Context, id string) (*service.CustomerDetail, error) {
	if id == "cust-1" {
		return &service.CustomerDetail{
			Customer: &domain.Customer{ID: "cust-1", Name: "Acme Corp", Type: domain.CustomerTypeRetail},
			Leads:    []*domain.Lead{},
		}, nil
	}
	return nil, store.ErrNotFound
}

func (s *stubCustomerSvc) List(_ context.Context, _ store.CustomerFilter, _, _ int) (*store.CustomerPage, error) {
	return &store.CustomerPage{Customers: []*domain.Customer{}, Total: 0, Page: 1, Limit: 20}, nil
}

func (s *stubCustomerSvc) Update(_ context.Context, actor *domain.User, c *domain.Customer) (*domain.Customer, error) {
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	if c.ID == "missing" {
		return nil, store.ErrNotFound
	}
	return c, nil
}

// --- helpers ---

func newTestCustomerRouter(user *domain.User) http.Handler {
	h := api.NewCustomerHandler(&stubCustomerSvc{})
	router := api.NewCustomerRouter(h)
	if user != nil {
		return injectUser(user, router)
	}
	return router
}

// --- tests ---

func TestCustomerList_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestCustomerRouter(testAdminUser()).ServeHTTP(w, req)

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

func TestCustomerList_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestCustomerRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerList_InvalidTypeFilter_Returns400(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/?type=bogus", http.NoBody)
	w := httptest.NewRecorder()
	newTestCustomerRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerCreate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name": "Test Customer", "type": "retail",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCustomerRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerCreate_RepForbidden(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name": "Test Customer", "type": "retail",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCustomerRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerCreate_MissingName_Returns400(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"type": "retail"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCustomerRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerCreate_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"name": "X", "type": "retail"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCustomerRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerGet_Found(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/cust-1", http.NoBody)
	w := httptest.NewRecorder()
	newTestCustomerRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["customer"]; !ok {
		t.Error("expected 'customer' key in response")
	}
	if _, ok := resp["leads"]; !ok {
		t.Error("expected 'leads' key in response")
	}
}

func TestCustomerGet_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/missing-cust", http.NoBody)
	w := httptest.NewRecorder()
	newTestCustomerRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerUpdate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"name": "Updated Corp", "type": "wholesale"})
	req := httptest.NewRequest(http.MethodPut, "/cust-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCustomerRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerUpdate_RepForbidden(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"name": "Updated Corp", "type": "wholesale"})
	req := httptest.NewRequest(http.MethodPut, "/cust-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCustomerRouter(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCustomerUpdate_MissingName_Returns400(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"type": "retail"})
	req := httptest.NewRequest(http.MethodPut, "/cust-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestCustomerRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
