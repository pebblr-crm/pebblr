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
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

const (
	testAdminID       = "admin-1"
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
	fmtExpected200    = "expected 200, got %d: %s"
	fmtDecodeErr      = "decoding response: %v"
	fmtExpected400    = "expected 400, got %d: %s"
	pathActivity1     = "/activity-1"
	pathSubmitted     = "/submitted"
	pathStatus        = "/status"
	fmtExpected409    = "expected 409, got %d: %s"
	testDate          = "2026-03-24"
)

// --- stub ActivityService ---

type stubActivitySvc struct {
	activity *domain.Activity
}

func defaultStubActivity() *domain.Activity {
	return &domain.Activity{
		ID:           "activity-1",
		ActivityType: "visit",
		Status:       "planificat",
		DueDate:      time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC),
		Duration:     "full_day",
		Fields:       map[string]any{},
		CreatorID:    testAdminID,
		TeamID:       testTeamID,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

func (s *stubActivitySvc) Create(_ context.Context, actor *domain.User, activity *domain.Activity) (*domain.Activity, error) {
	a := s.getActivity()
	a.CreatorID = actor.ID
	a.ActivityType = activity.ActivityType
	a.DueDate = activity.DueDate
	a.Duration = activity.Duration
	a.Fields = activity.Fields
	return a, nil
}

func (s *stubActivitySvc) Get(_ context.Context, _ *domain.User, id string) (*domain.Activity, error) {
	if id == "missing" {
		return nil, store.ErrNotFound
	}
	a := s.getActivity()
	a.ID = id
	return a, nil
}

func (s *stubActivitySvc) List(_ context.Context, _ *domain.User, _ store.ActivityFilter, page, limit int) (*store.ActivityPage, error) {
	return &store.ActivityPage{
		Activities: []*domain.Activity{s.getActivity()},
		Total:      1,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (s *stubActivitySvc) Update(_ context.Context, actor *domain.User, id string, activity *domain.Activity) (*domain.Activity, error) {
	if id == "submitted" {
		return nil, service.ErrSubmitted
	}
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	a := s.getActivity()
	a.ID = id
	a.ActivityType = activity.ActivityType
	a.DueDate = activity.DueDate
	return a, nil
}

func (s *stubActivitySvc) Delete(_ context.Context, actor *domain.User, id string) error {
	if id == "submitted" {
		return service.ErrSubmitted
	}
	if actor.Role == domain.RoleRep {
		return service.ErrForbidden
	}
	return nil
}

func (s *stubActivitySvc) Submit(_ context.Context, _ *domain.User, id string) (*domain.Activity, error) {
	if id == "submitted" {
		return nil, service.ErrSubmitted
	}
	a := s.getActivity()
	a.ID = id
	now := time.Now()
	a.SubmittedAt = &now
	return a, nil
}

func (s *stubActivitySvc) PartialUpdate(_ context.Context, actor *domain.User, id string, patch *domain.ActivityPatch) (*domain.Activity, error) {
	if id == "submitted" {
		return nil, service.ErrSubmitted
	}
	if actor.Role == domain.RoleRep {
		return nil, service.ErrForbidden
	}
	if id == "missing" {
		return nil, store.ErrNotFound
	}
	a := s.getActivity()
	a.ID = id
	patch.ApplyTo(a)
	return a, nil
}

func (s *stubActivitySvc) PatchStatus(_ context.Context, _ *domain.User, id, newStatus string) (*domain.Activity, error) {
	if id == "submitted" {
		return nil, service.ErrSubmitted
	}
	a := s.getActivity()
	a.ID = id
	a.Status = newStatus
	return a, nil
}

func (s *stubActivitySvc) CloneWeek(_ context.Context, _ *domain.User, _, _ time.Time) (*service.CloneWeekResult, error) {
	return &service.CloneWeekResult{Created: 5}, nil
}

func (s *stubActivitySvc) getActivity() *domain.Activity {
	if s.activity != nil {
		a := *s.activity
		return &a
	}
	return defaultStubActivity()
}

func activityRouter() http.Handler {
	h := api.NewActivityHandler(&stubActivitySvc{})
	return api.NewActivityRouter(h)
}

func activityReq(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	return w
}

func activityReqAsRep(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: "rep-1", Role: domain.RoleRep, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	return w
}

// --- List tests ---

func TestActivityList_ReturnsOK(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "GET", "/", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	if _, ok := resp["items"]; !ok {
		t.Error("expected 'items' key in response")
	}
}

func TestActivityList_WithFilters(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "GET", "/?activityType=visit&status=planificat&dateFrom=2026-03-01&dateTo=2026-03-31", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

// --- Create tests ---

func TestActivityCreate_Succeeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"activityType": "visit",
		"dueDate":      "2026-03-23",
		"duration":     "full_day",
		"fields":       map[string]any{},
	}
	w := activityReq(t, "POST", "/", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	if _, ok := resp["activity"]; !ok {
		t.Error("expected 'activity' key in response")
	}
}

func TestActivityCreate_MissingType(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"dueDate": "2026-03-23",
	}
	w := activityReq(t, "POST", "/", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityCreate_MissingDueDate(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"activityType": "visit",
	}
	w := activityReq(t, "POST", "/", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityCreate_InvalidDateFormat(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"activityType": "visit",
		"dueDate":      "03/23/2026",
	}
	w := activityReq(t, "POST", "/", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

// --- Get tests ---

func TestActivityGet_ReturnsOK(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "GET", pathActivity1, nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityGet_NotFound(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "GET", "/missing", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Update tests ---

func TestActivityUpdate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"activityType": "visit",
		"status":       "planificat",
		"dueDate":      testDate,
		"fields":       map[string]any{},
	}
	w := activityReq(t, "PUT", pathActivity1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityUpdate_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"activityType": "visit",
		"status":       "planificat",
		"dueDate":      testDate,
		"fields":       map[string]any{},
	}
	w := activityReqAsRep(t, "PUT", pathActivity1, body)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityUpdate_SubmittedConflict(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"activityType": "visit",
		"status":       "planificat",
		"dueDate":      testDate,
		"fields":       map[string]any{},
	}
	w := activityReq(t, "PUT", pathSubmitted, body)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

// --- Delete tests ---

func TestActivityDelete_AdminSucceeds(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "DELETE", pathActivity1, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityDelete_SubmittedConflict(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "DELETE", pathSubmitted, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

// --- Submit tests ---

func TestActivitySubmit_Succeeds(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "POST", pathActivity1 + "/submit", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	activity := resp["activity"].(map[string]any)
	if activity["submittedAt"] == nil {
		t.Error("expected submittedAt to be set")
	}
}

func TestActivitySubmit_AlreadySubmitted(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "POST", pathSubmitted + "/submit", nil)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

// --- PatchStatus tests ---

func TestActivityPatchStatus_Succeeds(t *testing.T) {
	t.Parallel()
	body := map[string]any{"status": "realizat"}
	w := activityReq(t, "PATCH", pathActivity1+pathStatus, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	activity := resp["activity"].(map[string]any)
	if activity["status"] != "realizat" {
		t.Errorf("expected realizat, got %v", activity["status"])
	}
}

func TestActivityPatchStatus_MissingStatus(t *testing.T) {
	t.Parallel()
	body := map[string]any{}
	w := activityReq(t, "PATCH", pathActivity1+pathStatus, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityPatchStatus_SubmittedConflict(t *testing.T) {
	t.Parallel()
	body := map[string]any{"status": "realizat"}
	w := activityReq(t, "PATCH", pathSubmitted+pathStatus, body)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

// --- Patch tests ---

func TestActivityPatch_StatusOnly(t *testing.T) {
	t.Parallel()
	body := map[string]any{"status": "realizat"}
	w := activityReq(t, "PATCH", pathActivity1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	activity := resp["activity"].(map[string]any)
	if activity["status"] != "realizat" {
		t.Errorf("expected realizat, got %v", activity["status"])
	}
}

func TestActivityPatch_DueDate(t *testing.T) {
	t.Parallel()
	body := map[string]any{"dueDate": "2026-03-30"}
	w := activityReq(t, "PATCH", pathActivity1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityPatch_InvalidDueDateFormat(t *testing.T) {
	t.Parallel()
	body := map[string]any{"dueDate": "30/03/2026"}
	w := activityReq(t, "PATCH", pathActivity1, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityPatch_FieldsMerge(t *testing.T) {
	t.Parallel()
	body := map[string]any{"fields": map[string]any{"notes": "updated note"}}
	w := activityReq(t, "PATCH", pathActivity1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	activity := resp["activity"].(map[string]any)
	fields := activity["fields"].(map[string]any)
	if fields["notes"] != "updated note" {
		t.Errorf("expected notes=updated note, got %v", fields["notes"])
	}
}

func TestActivityPatch_EmptyBodyIsNoOp(t *testing.T) {
	t.Parallel()
	body := map[string]any{}
	w := activityReq(t, "PATCH", pathActivity1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityPatch_RepForbidden(t *testing.T) {
	t.Parallel()
	body := map[string]any{"status": "realizat"}
	w := activityReqAsRep(t, "PATCH", pathActivity1, body)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityPatch_SubmittedConflict(t *testing.T) {
	t.Parallel()
	body := map[string]any{"status": "realizat"}
	w := activityReq(t, "PATCH", pathSubmitted, body)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

func TestActivityPatch_NotFound(t *testing.T) {
	t.Parallel()
	body := map[string]any{"status": "realizat"}
	w := activityReq(t, "PATCH", "/missing", body)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityPatch_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("PATCH", pathActivity1, bytes.NewBufferString("not-json"))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}
