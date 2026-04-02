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

const (
	fmtCollExpected200 = "expected 200, got %d: %s"
	pathCol1           = "/col-1"
	pathMissing        = "/missing"
	fmtExpected404     = "expected 404, got %d"
)

// --- stub CollectionService ---

type stubCollectionSvc struct {
	collections []*domain.Collection
}

func (s *stubCollectionSvc) Create(_ context.Context, actor *domain.User, name string, targetIDs []string) (*domain.Collection, error) {
	if name == "" {
		return nil, service.ErrInvalidInput
	}
	c := &domain.Collection{
		ID:        "col-new",
		Name:      name,
		CreatorID: actor.ID,
		TeamID:    actor.TeamIDs[0],
		TargetIDs: targetIDs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.collections = append(s.collections, c)
	return c, nil
}

func (s *stubCollectionSvc) List(_ context.Context, _ *domain.User) ([]*domain.Collection, error) {
	return s.collections, nil
}

func (s *stubCollectionSvc) Get(_ context.Context, actor *domain.User, id string) (*domain.Collection, error) {
	for _, c := range s.collections {
		if c.ID != id {
			continue
		}
		if actor.Role == domain.RoleRep && actor.ID != c.CreatorID {
			return nil, service.ErrForbidden
		}
		return c, nil
	}
	return nil, store.ErrNotFound
}

func (s *stubCollectionSvc) Update(_ context.Context, actor *domain.User, id, name string, targetIDs []string) (*domain.Collection, error) {
	if name == "" {
		return nil, service.ErrInvalidInput
	}
	for i, c := range s.collections {
		if c.ID != id {
			continue
		}
		if actor.Role == domain.RoleRep && actor.ID != c.CreatorID {
			return nil, service.ErrForbidden
		}
		c.Name = name
		c.TargetIDs = targetIDs
		s.collections[i] = c
		return c, nil
	}
	return nil, store.ErrNotFound
}

func (s *stubCollectionSvc) Delete(_ context.Context, actor *domain.User, id string) error {
	for i, c := range s.collections {
		if c.ID == id {
			if actor.Role == domain.RoleRep && actor.ID != c.CreatorID {
				return service.ErrForbidden
			}
			s.collections = append(s.collections[:i], s.collections[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func defaultCollections() []*domain.Collection {
	return []*domain.Collection{
		{
			ID:        "col-1",
			Name:      "Monday Route",
			CreatorID: "rep-1",
			TeamID:    "team-1",
			TargetIDs: []string{"t1", "t2"},
		},
	}
}

func newTestCollectionRouter(user *domain.User) http.Handler {
	h := api.NewCollectionHandler(&stubCollectionSvc{collections: defaultCollections()})
	router := api.NewCollectionRouter(h)
	if user != nil {
		return injectUser(user, router)
	}
	return router
}

// --- tests ---

func TestCollectionList_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf(fmtCollExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatal("expected 'items' array in response")
	}
	if len(items) != 1 {
		t.Errorf("expected 1 collection, got %d", len(items))
	}
}

func TestCollectionList_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	newTestCollectionRouter(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCollectionGet_Found(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, pathCol1, http.NoBody)
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf(fmtCollExpected200, w.Code, w.Body.String())
	}
}

func TestCollectionGet_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, pathMissing, http.NoBody)
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf(fmtExpected404, w.Code)
	}
}

func TestCollectionCreate_Returns201(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":      "New Collection",
		"targetIds": []string{"t1"},
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCollectionCreate_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{bad")))
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCollectionCreate_EmptyName_Returns400(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":      "",
		"targetIds": []string{},
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCollectionUpdate_Returns200(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":      "Renamed",
		"targetIds": []string{"t1", "t3"},
	})
	req := httptest.NewRequest(http.MethodPut, pathCol1, bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf(fmtCollExpected200, w.Code, w.Body.String())
	}
}

func TestCollectionUpdate_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPut, pathCol1, bytes.NewReader([]byte("{bad")))
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCollectionUpdate_NotFound(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"name":      "Update",
		"targetIds": []string{},
	})
	req := httptest.NewRequest(http.MethodPut, pathMissing, bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf(fmtExpected404, w.Code)
	}
}

func TestCollectionDelete_Returns204(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, pathCol1, http.NoBody)
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCollectionDelete_NotFound(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodDelete, pathMissing, http.NoBody)
	w := httptest.NewRecorder()
	newTestCollectionRouter(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf(fmtExpected404, w.Code)
	}
}
