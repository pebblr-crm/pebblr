package api_test

import (
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

// --- stub DashboardService ---

type stubDashboardSvc struct {
	statsResult    *service.DashboardStatsResponse
	coverageResult *store.CoverageStats
	userResult     []store.UserActivityStats
	err            error
}

func (s *stubDashboardSvc) ActivityStats(_ context.Context, _ *domain.User, _ string, _ store.DashboardFilter) (*service.DashboardStatsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.statsResult, nil
}

func (s *stubDashboardSvc) CoverageStats(_ context.Context, _ *domain.User, _ string, _ store.DashboardFilter) (*store.CoverageStats, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.coverageResult, nil
}

func (s *stubDashboardSvc) UserStats(_ context.Context, _ *domain.User, _ string, _ store.DashboardFilter) ([]store.UserActivityStats, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.userResult, nil
}

func defaultDashboardSvc() *stubDashboardSvc {
	return &stubDashboardSvc{
		statsResult: &service.DashboardStatsResponse{
			Period:   "2026-03",
			DateFrom: "2026-03-01",
			DateTo:   "2026-03-31",
			Stats: &store.ActivityStats{
				Total:          50,
				SubmittedCount: 30,
				ByStatus: []store.StatusCount{
					{Status: "realizat", Count: 30},
					{Status: "planificat", Count: 15},
					{Status: "anulat", Count: 5},
				},
				ByType: []store.TypeCount{
					{ActivityType: "visit", Count: 40},
					{ActivityType: "administrative", Count: 10},
				},
			},
			ByCategory: []service.CategoryCount{
				{Category: "field", Count: 40},
				{Category: "non_field", Count: 10},
			},
		},
		coverageResult: &store.CoverageStats{
			TotalTargets:    100,
			VisitedTargets:  75,
			CoveragePercent: 75.0,
		},
		userResult: []store.UserActivityStats{
			{
				UserID:   "rep-1",
				UserName: "Alice",
				Total:    20,
				ByStatus: []store.StatusCount{{Status: "realizat", Count: 15}},
			},
		},
	}
}

func dashboardRouter(svc *stubDashboardSvc) http.Handler {
	h := api.NewDashboardHandler(svc)
	return api.NewDashboardRouter(h)
}

func dashboardReq(t *testing.T, router http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, http.NoBody)
	user := &domain.User{ID: "admin-1", Role: domain.RoleAdmin, TeamIDs: []string{"team-1"}}
	ctx := rbac.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// --- Stats endpoint tests ---

func TestDashboardStats_Success(t *testing.T) {
	t.Parallel()
	svc := defaultDashboardSvc()
	r := dashboardRouter(svc)
	w := dashboardReq(t, r, "GET", "/stats?period=2026-03")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result service.DashboardStatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if result.Period != "2026-03" {
		t.Errorf("expected period 2026-03, got %s", result.Period)
	}
	if result.Stats.Total != 50 {
		t.Errorf("expected total 50, got %d", result.Stats.Total)
	}
}

func TestDashboardStats_MissingPeriod(t *testing.T) {
	t.Parallel()
	svc := defaultDashboardSvc()
	r := dashboardRouter(svc)
	w := dashboardReq(t, r, "GET", "/stats")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDashboardStats_InvalidPeriod(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardSvc{err: service.ErrInvalidInput}
	r := dashboardRouter(svc)
	w := dashboardReq(t, r, "GET", "/stats?period=bad")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDashboardStats_WithFilters(t *testing.T) {
	t.Parallel()
	svc := defaultDashboardSvc()
	r := dashboardRouter(svc)
	w := dashboardReq(t, r, "GET", "/stats?period=2026-03&teamId=team-1&creatorId=rep-1")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Coverage endpoint tests ---

func TestDashboardCoverage_Success(t *testing.T) {
	t.Parallel()
	svc := defaultDashboardSvc()
	r := dashboardRouter(svc)
	w := dashboardReq(t, r, "GET", "/coverage?period=2026-03")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result store.CoverageStats
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if result.TotalTargets != 100 {
		t.Errorf("expected 100 total targets, got %d", result.TotalTargets)
	}
	if result.CoveragePercent != 75.0 {
		t.Errorf("expected 75%% coverage, got %.1f", result.CoveragePercent)
	}
}

func TestDashboardCoverage_MissingPeriod(t *testing.T) {
	t.Parallel()
	svc := defaultDashboardSvc()
	r := dashboardRouter(svc)
	w := dashboardReq(t, r, "GET", "/coverage")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- UserStats endpoint tests ---

func TestDashboardUserStats_Success(t *testing.T) {
	t.Parallel()
	svc := defaultDashboardSvc()
	r := dashboardRouter(svc)
	w := dashboardReq(t, r, "GET", "/user-stats?period=2026-03")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result struct {
		Users []store.UserActivityStats `json:"users"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(result.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(result.Users))
	}
	if result.Users[0].UserID != "rep-1" {
		t.Errorf("expected user rep-1, got %s", result.Users[0].UserID)
	}
}

func TestDashboardUserStats_MissingPeriod(t *testing.T) {
	t.Parallel()
	svc := defaultDashboardSvc()
	r := dashboardRouter(svc)
	w := dashboardReq(t, r, "GET", "/user-stats")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDashboardStats_NoAuth(t *testing.T) {
	t.Parallel()
	svc := defaultDashboardSvc()
	r := dashboardRouter(svc)

	req := httptest.NewRequest("GET", "/stats?period=2026-03", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
