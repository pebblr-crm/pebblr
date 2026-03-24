package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub TargetService ---

type stubTargetSvc struct{}

func (s *stubTargetSvc) Create(_ context.Context, actor *domain.User, target *domain.Target) (*domain.Target, error) {
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	target.ID = "target-new"
	return target, nil
}

func (s *stubTargetSvc) Get(_ context.Context, actor *domain.User, id string) (*domain.Target, error) {
	if id == "missing" {
		return nil, store.ErrNotFound
	}
	return &domain.Target{
		ID:         id,
		TargetType: "doctor",
		Name:       "Dr. Test",
		Fields:     map[string]any{},
		AssigneeID: actor.ID,
		TeamID:     "team-1",
	}, nil
}

func (s *stubTargetSvc) List(_ context.Context, _ *domain.User, _ store.TargetFilter, page, limit int) (*store.TargetPage, error) {
	return &store.TargetPage{
		Targets: []*domain.Target{
			{ID: "target-1", TargetType: "doctor", Name: "Dr. Test", Fields: map[string]any{}},
		},
		Total: 1,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *stubTargetSvc) Update(_ context.Context, actor *domain.User, target *domain.Target) (*domain.Target, error) {
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	return target, nil
}

func (s *stubTargetSvc) Assign(_ context.Context, actor *domain.User, targetID, assigneeID, teamID string) (*domain.Target, error) {
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	return &domain.Target{
		ID:         targetID,
		TargetType: "doctor",
		Name:       "Dr. Test",
		Fields:     map[string]any{},
		AssigneeID: assigneeID,
		TeamID:     teamID,
	}, nil
}

func (s *stubTargetSvc) Import(_ context.Context, actor *domain.User, targets []*domain.Target) (*store.ImportResult, error) {
	if actor.Role != domain.RoleAdmin {
		return nil, service.ErrForbidden
	}
	for i, t := range targets {
		t.ID = fmt.Sprintf("imported-%d", i+1)
	}
	return &store.ImportResult{Created: len(targets), Imported: targets}, nil
}

func (s *stubTargetSvc) VisitStatus(_ context.Context, _ *domain.User) ([]store.TargetVisitStatus, error) {
	return []store.TargetVisitStatus{}, nil
}

func (s *stubTargetSvc) FrequencyStatus(_ context.Context, _ *domain.User, _, _ time.Time) ([]service.TargetFrequencyItem, error) {
	return []service.TargetFrequencyItem{}, nil
}

func targetRouter() http.Handler {
	h := api.NewTargetHandler(&stubTargetSvc{})
	return api.NewTargetRouter(h)
}

func targetReq(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	// Inject an admin user into the context for auth.
	user := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, TeamIDs: []string{"team-1"}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	return w
}

func targetReqAsRep(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	user := &domain.User{ID: "rep-1", Role: domain.RoleRep, TeamIDs: []string{"team-1"}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	return w
}

// --- List tests ---

func TestTargetList_ReturnsOK(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if _, ok := resp["items"]; !ok {
		t.Error("expected 'items' key in response")
	}
}

// --- Create tests ---

func TestTargetCreate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       "Dr. New",
		"fields":     map[string]any{},
	}
	w := targetReq(t, "POST", "/", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTargetCreate_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       "Dr. New",
		"fields":     map[string]any{},
	}
	w := targetReqAsRep(t, "POST", "/", body)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTargetCreate_MissingName(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
	}
	w := targetReq(t, "POST", "/", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Get tests ---

func TestTargetGet_ReturnsOK(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/target-1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTargetGet_NotFound(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/missing", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Update tests ---

func TestTargetUpdate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       "Dr. Updated",
		"fields":     map[string]any{},
	}
	w := targetReq(t, "PUT", "/target-1", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTargetUpdate_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       "Dr. Updated",
		"fields":     map[string]any{},
	}
	w := targetReqAsRep(t, "PUT", "/target-1", body)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Import tests ---

func TestTargetImport_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targets": []map[string]any{
			{"externalId": "ext-1", "targetType": "doctor", "name": "Dr. Import", "fields": map[string]any{}},
			{"externalId": "ext-2", "targetType": "pharmacy", "name": "Central Pharmacy", "fields": map[string]any{}},
		},
	}
	w := targetReq(t, "POST", "/import", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if int(resp["created"].(float64)) != 2 {
		t.Errorf("expected 2 created, got %v", resp["created"])
	}
}

func TestTargetImport_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targets": []map[string]any{
			{"externalId": "ext-1", "targetType": "doctor", "name": "Dr. Import"},
		},
	}
	w := targetReqAsRep(t, "POST", "/import", body)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTargetImport_EmptyTargets(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targets": []map[string]any{},
	}
	w := targetReq(t, "POST", "/import", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
