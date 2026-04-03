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

const (
	testDrName    = "Dr. Test"
	testDrNew     = "Dr. New"
	testDrUpdated = "Dr. Updated"
	testDrImport  = "Dr. Import"
	pathImport    = "/import"
	pathTarget1   = "/target-1"
	pathAssign    = "/assign"
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
		Name:       testDrName,
		Fields:     map[string]any{},
		AssigneeID: actor.ID,
		TeamID:     testTeamID,
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
		Name:       testDrName,
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
			t.Fatalf(fmtEncodingBody, err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set(headerContentType, contentTypeJSON)
	// Inject an admin user into the context for auth.
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
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
			t.Fatalf(fmtEncodingBody, err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: "rep-1", Role: domain.RoleRep, TeamIDs: []string{testTeamID}}
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
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	if _, ok := resp["items"]; !ok {
		t.Error(msgExpectedItems)
	}
}

// --- Create tests ---

func TestTargetCreate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       testDrNew,
		"fields":     map[string]any{},
	}
	w := targetReq(t, "POST", "/", body)
	if w.Code != http.StatusCreated {
		t.Fatalf(fmtExpected201, w.Code, w.Body.String())
	}
}

func TestTargetCreate_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       testDrNew,
		"fields":     map[string]any{},
	}
	w := targetReqAsRep(t, "POST", "/", body)
	if w.Code != http.StatusForbidden {
		t.Fatalf(fmtExpected403, w.Code, w.Body.String())
	}
}

func TestTargetCreate_MissingName(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
	}
	w := targetReq(t, "POST", "/", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

// --- Get tests ---

func TestTargetGet_ReturnsOK(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", pathTarget1, nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
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
		"name":       testDrUpdated,
		"fields":     map[string]any{},
	}
	w := targetReq(t, "PUT", pathTarget1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestTargetUpdate_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       testDrUpdated,
		"fields":     map[string]any{},
	}
	w := targetReqAsRep(t, "PUT", pathTarget1, body)
	if w.Code != http.StatusForbidden {
		t.Fatalf(fmtExpected403, w.Code, w.Body.String())
	}
}

// --- Import tests ---

func TestTargetImport_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targets": []map[string]any{
			{"externalId": "ext-1", "targetType": "doctor", "name": testDrImport, "fields": map[string]any{}},
			{"externalId": "ext-2", "targetType": "pharmacy", "name": "Central Pharmacy", "fields": map[string]any{}},
		},
	}
	w := targetReq(t, "POST", pathImport, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	if int(resp["created"].(float64)) != 2 {
		t.Errorf("expected 2 created, got %v", resp["created"])
	}
}

func TestTargetImport_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targets": []map[string]any{
			{"externalId": "ext-1", "targetType": "doctor", "name": testDrImport},
		},
	}
	w := targetReqAsRep(t, "POST", pathImport, body)
	if w.Code != http.StatusForbidden {
		t.Fatalf(fmtExpected403, w.Code, w.Body.String())
	}
}

func TestTargetImport_EmptyTargets(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targets": []map[string]any{},
	}
	w := targetReq(t, "POST", pathImport, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

// --- Assign tests ---

func TestTargetAssign_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"assigneeId": "rep-1",
		"teamId":     testTeamID,
	}
	w := targetReq(t, "PATCH", pathTarget1+pathAssign, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestTargetAssign_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"assigneeId": "rep-2",
		"teamId":     testTeamID,
	}
	w := targetReqAsRep(t, "PATCH", pathTarget1+pathAssign, body)
	if w.Code != http.StatusForbidden {
		t.Fatalf(fmtExpected403, w.Code, w.Body.String())
	}
}

func TestTargetAssign_MissingAssigneeID(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"teamId": testTeamID,
	}
	w := targetReq(t, "PATCH", pathTarget1+pathAssign, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestTargetAssign_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("PATCH", pathTarget1+pathAssign, bytes.NewBufferString(invalidJSON))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

// --- VisitStatus tests ---

func TestTargetVisitStatus_ReturnsOK(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/visit-status", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	if _, ok := resp["items"]; !ok {
		t.Error(msgExpectedItems)
	}
}

// --- FrequencyStatus tests ---

func TestTargetFrequencyStatus_ReturnsOK(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/frequency-status?period=2026-03", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	if _, ok := resp["items"]; !ok {
		t.Error(msgExpectedItems)
	}
}

func TestTargetFrequencyStatus_InvalidPeriod(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/frequency-status?period=bad", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

// --- List filter tests ---

func TestTargetList_WithFilters(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/?type=doctor&assignee=rep-1&q=test", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestTargetList_PaginationDefaults(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/?page=0&limit=0", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestTargetList_PaginationLimitCap(t *testing.T) {
	t.Parallel()
	w := targetReq(t, "GET", "/?limit=999", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

// --- Create/Update edge cases ---

func TestTargetCreate_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(invalidJSON))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestTargetUpdate_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("PUT", pathTarget1, bytes.NewBufferString(invalidJSON))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestTargetUpdate_MissingName(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
	}
	w := targetReq(t, "PUT", pathTarget1, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestTargetImport_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("POST", pathImport, bytes.NewBufferString(invalidJSON))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

// --- NoAuth tests ---

func TestTargetList_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/", http.NoBody)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf(fmtExpected401, w.Code, w.Body.String())
	}
}

func TestTargetCreate_NoAuth(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"name": testDrNew, "targetType": "doctor"})
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set(headerContentType, contentTypeJSON)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf(fmtExpected401, w.Code, w.Body.String())
	}
}

func TestTargetVisitStatus_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/visit-status", http.NoBody)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf(fmtExpected401, w.Code, w.Body.String())
	}
}

func TestTargetFrequencyStatus_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/frequency-status?period=2026-03", http.NoBody)
	w := httptest.NewRecorder()
	targetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf(fmtExpected401, w.Code, w.Body.String())
	}
}

func TestTargetImport_NilFields(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targets": []map[string]any{
			{"externalId": "ext-1", "targetType": "doctor", "name": testDrImport},
		},
	}
	w := targetReq(t, "POST", pathImport, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestTargetCreate_NilFields(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       "Dr. NoFields",
	}
	w := targetReq(t, "POST", "/", body)
	if w.Code != http.StatusCreated {
		t.Fatalf(fmtExpected201, w.Code, w.Body.String())
	}
}

func TestTargetUpdate_NilFields(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"targetType": "doctor",
		"name":       testDrUpdated,
	}
	w := targetReq(t, "PUT", pathTarget1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}
