package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/config"
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
	activity  *domain.Activity
	createErr error
	listErr   error
	getErr    error
	cloneErr  error
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
	if s.createErr != nil {
		return nil, s.createErr
	}
	a := s.getActivity()
	a.CreatorID = actor.ID
	a.ActivityType = activity.ActivityType
	a.DueDate = activity.DueDate
	a.Duration = activity.Duration
	a.Fields = activity.Fields
	return a, nil
}

func (s *stubActivitySvc) Get(_ context.Context, _ *domain.User, id string) (*domain.Activity, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if id == "missing" {
		return nil, store.ErrNotFound
	}
	a := s.getActivity()
	a.ID = id
	return a, nil
}

func (s *stubActivitySvc) List(_ context.Context, _ *domain.User, _ store.ActivityFilter, page, limit int) (*store.ActivityPage, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
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
	if s.cloneErr != nil {
		return nil, s.cloneErr
	}
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
	return activityRouterWithSvc(&stubActivitySvc{})
}

func activityRouterWithSvc(svc api.ActivityServicer) http.Handler {
	h := api.NewActivityHandler(svc)
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

// --- BatchCreate tests ---

func TestActivityBatchCreate_Success(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"items": []map[string]any{
			{"targetId": "t-1", "dueDate": "2026-03-24"},
			{"targetId": "t-2", "dueDate": "2026-03-25"},
		},
	}
	w := activityReq(t, "POST", "/batch", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	created := resp["created"].([]any)
	if len(created) != 2 {
		t.Errorf("expected 2 created, got %d", len(created))
	}
}

func TestActivityBatchCreate_EmptyItems(t *testing.T) {
	t.Parallel()
	body := map[string]any{"items": []map[string]any{}}
	w := activityReq(t, "POST", "/batch", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityBatchCreate_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("POST", "/batch", bytes.NewBufferString("not-json"))
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

func TestActivityBatchCreate_InvalidDateInItem(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"items": []map[string]any{
			{"targetId": "t-1", "dueDate": "bad-date"},
		},
	}
	w := activityReq(t, "POST", "/batch", body)
	// All items failed -> 400
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	errs := resp["errors"].([]any)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

func TestActivityBatchCreate_MixedResults_MultiStatus(t *testing.T) {
	t.Parallel()
	// One valid, one with bad date -> multi status
	body := map[string]any{
		"items": []map[string]any{
			{"targetId": "t-1", "dueDate": "2026-03-24"},
			{"targetId": "t-2", "dueDate": "bad-date"},
		},
	}
	w := activityReq(t, "POST", "/batch", body)
	if w.Code != http.StatusMultiStatus {
		t.Fatalf("expected 207, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityBatchCreate_WithFields(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"items": []map[string]any{
			{"targetId": "t-1", "dueDate": "2026-03-24", "fields": map[string]any{"notes": "test"}},
		},
	}
	w := activityReq(t, "POST", "/batch", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityBatchCreate_NoAuth(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{
			{"targetId": "t-1", "dueDate": "2026-03-24"},
		},
	})
	req := httptest.NewRequest("POST", "/batch", bytes.NewReader(body))
	req.Header.Set(headerContentType, contentTypeJSON)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// --- CloneWeek tests ---

func TestActivityCloneWeek_Success(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"sourceWeekStart": "2026-03-16",
		"targetWeekStart": "2026-03-23",
	}
	w := activityReq(t, "POST", "/clone-week", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	if int(resp["created"].(float64)) != 5 {
		t.Errorf("expected 5 created, got %v", resp["created"])
	}
}

func TestActivityCloneWeek_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("POST", "/clone-week", bytes.NewBufferString("not-json"))
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

func TestActivityCloneWeek_InvalidSourceDate(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"sourceWeekStart": "bad",
		"targetWeekStart": "2026-03-23",
	}
	w := activityReq(t, "POST", "/clone-week", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityCloneWeek_InvalidTargetDate(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"sourceWeekStart": "2026-03-16",
		"targetWeekStart": "bad",
	}
	w := activityReq(t, "POST", "/clone-week", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityCloneWeek_NoAuth(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{
		"sourceWeekStart": "2026-03-16",
		"targetWeekStart": "2026-03-23",
	})
	req := httptest.NewRequest("POST", "/clone-week", bytes.NewReader(body))
	req.Header.Set(headerContentType, contentTypeJSON)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Error mapping tests ---

func TestActivityCreate_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString("not-json"))
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

func TestActivityUpdate_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("PUT", pathActivity1, bytes.NewBufferString("not-json"))
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

func TestActivityPatchStatus_InvalidBody(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("PATCH", pathActivity1+pathStatus, bytes.NewBufferString("not-json"))
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

func TestActivityDelete_RepForbidden(t *testing.T) {
	t.Parallel()
	w := activityReqAsRep(t, "DELETE", pathActivity1, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Filter tests ---

func TestActivityList_WithCreatorAndTargetFilters(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "GET", "/?creatorId=user-1&targetId=target-1&teamId=team-1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityList_PaginationDefaults(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "GET", "/?page=0&limit=0", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	if int(resp["page"].(float64)) != 1 {
		t.Errorf("expected page 1, got %v", resp["page"])
	}
}

func TestActivityList_PaginationLimitCap(t *testing.T) {
	t.Parallel()
	w := activityReq(t, "GET", "/?limit=999", nil)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityList_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/", http.NoBody)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Patch field edge cases ---

func TestActivityPatch_FieldsWithJointVisitUID(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"fields": map[string]any{
			"joint_visit_user_id": "user-2",
			"notes":              "test",
		},
	}
	w := activityReq(t, "PATCH", pathActivity1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityPatch_FieldsNull(t *testing.T) {
	t.Parallel()
	// Send {"fields": null} — should be treated as fields present but nil
	req := httptest.NewRequest("PATCH", pathActivity1, bytes.NewBufferString(`{"fields": null}`))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityPatch_DueDateNull(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("PATCH", pathActivity1, bytes.NewBufferString(`{"dueDate": null}`))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityPatch_StatusNull(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("PATCH", pathActivity1, bytes.NewBufferString(`{"status": null}`))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityPatch_DurationAndRouting(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"duration": "half_day",
		"routing":  "week_1",
		"targetId": "target-2",
	}
	w := activityReq(t, "PATCH", pathActivity1, body)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

func TestActivityPatch_InvalidDueDateType(t *testing.T) {
	t.Parallel()
	// dueDate as a number, not string — should fail parseRawDueDate
	req := httptest.NewRequest("PATCH", pathActivity1, bytes.NewBufferString(`{"dueDate": 12345}`))
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

func TestActivityPatch_InvalidFieldsType(t *testing.T) {
	t.Parallel()
	// fields as a string, not object
	req := httptest.NewRequest("PATCH", pathActivity1, bytes.NewBufferString(`{"fields": "bad"}`))
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

func TestActivityPatch_InvalidStatusType(t *testing.T) {
	t.Parallel()
	// status as number, not string
	req := httptest.NewRequest("PATCH", pathActivity1, bytes.NewBufferString(`{"status": 123}`))
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

func TestActivityPatch_JointVisitUIDNull(t *testing.T) {
	t.Parallel()
	// fields with joint_visit_user_id set to null
	req := httptest.NewRequest("PATCH", pathActivity1, bytes.NewBufferString(`{"fields": {"joint_visit_user_id": null}}`))
	req.Header.Set(headerContentType, contentTypeJSON)
	user := &domain.User{ID: testAdminID, Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, w.Code, w.Body.String())
	}
}

// --- Create with joint visit user ---

func TestActivityCreate_WithJointVisitUID(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"activityType": "visit",
		"dueDate":      "2026-03-23",
		"fields":       map[string]any{"joint_visit_user_id": "user-2"},
	}
	w := activityReq(t, "POST", "/", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityCreate_NilFields(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"activityType": "visit",
		"dueDate":      "2026-03-23",
	}
	w := activityReq(t, "POST", "/", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

// --- NoAuth tests for additional endpoints ---

func TestActivityGet_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", pathActivity1, http.NoBody)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityCreate_NoAuth(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"activityType": "visit", "dueDate": "2026-03-23"})
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set(headerContentType, contentTypeJSON)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityUpdate_NoAuth(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"activityType": "visit", "dueDate": "2026-03-23"})
	req := httptest.NewRequest("PUT", pathActivity1, bytes.NewReader(body))
	req.Header.Set(headerContentType, contentTypeJSON)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityDelete_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("DELETE", pathActivity1, http.NoBody)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivitySubmit_NoAuth(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("POST", pathActivity1+"/submit", http.NoBody)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Error mapping coverage via custom stubs ---

func activityReqWithSvc(t *testing.T, svc api.ActivityServicer, method, path string, body any) *httptest.ResponseRecorder {
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
	activityRouterWithSvc(svc).ServeHTTP(w, req)
	return w
}

func TestActivityCreate_MaxActivitiesError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrMaxActivities}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

func TestActivityCreate_BlockedDayError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrBlockedDay}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

func TestActivityCreate_TargetRequiredError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrTargetRequired}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityCreate_InvalidJointVisitorError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrInvalidJointVisitor}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityCreate_TargetNotAccessibleError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrTargetNotAccessible}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityCreate_StatusNotSubmittableError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrStatusNotSubmittable}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

func TestActivityCreate_NoRecoveryBalanceError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrNoRecoveryBalance}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

func TestActivityCreate_DuplicateActivityError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrDuplicateActivity}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusConflict {
		t.Fatalf(fmtExpected409, w.Code, w.Body.String())
	}
}

func TestActivityCreate_InvalidInputError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrInvalidInput}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
}

func TestActivityCreate_UnexpectedError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: errors.New("unexpected")}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityCreate_ValidationError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: &service.ValidationErrors{Errors: []config.FieldError{{Field: "f", Message: "bad"}}}}
	body := map[string]any{"activityType": "visit", "dueDate": "2026-03-23"}
	w := activityReqWithSvc(t, svc, "POST", "/", body)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityList_ServiceError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{listErr: errors.New("db error")}
	w := activityReqWithSvc(t, svc, "GET", "/", nil)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityGet_ServiceError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{getErr: service.ErrForbidden}
	w := activityReqWithSvc(t, svc, "GET", pathActivity1, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityCloneWeek_ServiceError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{cloneErr: service.ErrForbidden}
	body := map[string]any{
		"sourceWeekStart": "2026-03-16",
		"targetWeekStart": "2026-03-23",
	}
	w := activityReqWithSvc(t, svc, "POST", "/clone-week", body)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActivityBatchCreate_ServiceError(t *testing.T) {
	t.Parallel()
	svc := &stubActivitySvc{createErr: service.ErrForbidden}
	body := map[string]any{
		"items": []map[string]any{
			{"targetId": "t-1", "dueDate": "2026-03-24"},
		},
	}
	w := activityReqWithSvc(t, svc, "POST", "/batch", body)
	// All items errored -> 400
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	errs := resp["errors"].([]any)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
	errMap := errs[0].(map[string]any)
	if errMap["error"] != "access denied" {
		t.Errorf("expected 'access denied', got %v", errMap["error"])
	}
}

func TestActivityBatchCreate_MixedServiceErrors(t *testing.T) {
	t.Parallel()
	// Test safeBatchError for various error types through batch
	// Use a stub that fails on specific targets
	svc := &stubActivitySvc{createErr: service.ErrDuplicateActivity}
	body := map[string]any{
		"items": []map[string]any{
			{"targetId": "t-1", "dueDate": "2026-03-24"},
		},
	}
	w := activityReqWithSvc(t, svc, "POST", "/batch", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf(fmtExpected400, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(fmtDecodeErr, err)
	}
	errs := resp["errors"].([]any)
	errMap := errs[0].(map[string]any)
	if errMap["error"] != "activity for this target on this date already exists" {
		t.Errorf("expected duplicate error message, got %v", errMap["error"])
	}
}

func TestActivityPatchStatus_NoAuth(t *testing.T) {
	t.Parallel()
	body, _ := json.Marshal(map[string]any{"status": "realizat"})
	req := httptest.NewRequest("PATCH", pathActivity1+pathStatus, bytes.NewReader(body))
	req.Header.Set(headerContentType, contentTypeJSON)
	w := httptest.NewRecorder()
	activityRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
