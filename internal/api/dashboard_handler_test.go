package api_test

import (
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

// --- stub dashboard service ---

type stubDashboardService struct {
	activityStats *service.ActivityStatsResponse
	coverage      *service.CoverageResponse
	frequency     *service.FrequencyResponse
	err           error
}

func (s *stubDashboardService) ActivityStats(_ context.Context, _ *domain.User, _ store.DashboardFilter) (*service.ActivityStatsResponse, error) {
	return s.activityStats, s.err
}

func (s *stubDashboardService) Coverage(_ context.Context, _ *domain.User, _ store.DashboardFilter) (*service.CoverageResponse, error) {
	return s.coverage, s.err
}

func (s *stubDashboardService) Frequency(_ context.Context, _ *domain.User, _ store.DashboardFilter) (*service.FrequencyResponse, error) {
	return s.frequency, s.err
}

func TestDashboardHandler_ActivityStats_OK(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardService{
		activityStats: &service.ActivityStatsResponse{
			ByStatus:   map[string]int{"planned": 5, "completed": 3},
			ByCategory: map[string]int{"field": 6, "non_field": 2},
			Total:      8,
		},
	}
	h := api.NewDashboardHandler(svc)
	router := api.NewDashboardRouter(h)
	handler := injectUser(testAdminUser(), router)

	req := httptest.NewRequest(http.MethodGet, "/activities?period=2026-03", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp service.ActivityStatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 8 {
		t.Errorf("total = %d, want 8", resp.Total)
	}
}

func TestDashboardHandler_ActivityStats_InvalidPeriod(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardService{}
	h := api.NewDashboardHandler(svc)
	router := api.NewDashboardRouter(h)
	handler := injectUser(testAdminUser(), router)

	req := httptest.NewRequest(http.MethodGet, "/activities?period=invalid", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDashboardHandler_ActivityStats_DefaultsToCurrentMonth(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardService{
		activityStats: &service.ActivityStatsResponse{
			ByStatus:   map[string]int{},
			ByCategory: map[string]int{},
			Total:      0,
		},
	}
	h := api.NewDashboardHandler(svc)
	router := api.NewDashboardRouter(h)
	handler := injectUser(testAdminUser(), router)

	req := httptest.NewRequest(http.MethodGet, "/activities", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDashboardHandler_ActivityStats_Unauthorized(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardService{}
	h := api.NewDashboardHandler(svc)
	router := api.NewDashboardRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/activities?period=2026-03", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestDashboardHandler_Coverage_OK(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardService{
		coverage: &service.CoverageResponse{
			TotalTargets:   20,
			VisitedTargets: 15,
			Percentage:     75,
		},
	}
	h := api.NewDashboardHandler(svc)
	router := api.NewDashboardRouter(h)
	handler := injectUser(testAdminUser(), router)

	req := httptest.NewRequest(http.MethodGet, "/coverage?period=2026-03", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp service.CoverageResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Percentage != 75 {
		t.Errorf("percentage = %f, want 75", resp.Percentage)
	}
}

func TestDashboardHandler_Frequency_OK(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardService{
		frequency: &service.FrequencyResponse{
			Items: []service.FrequencyItem{
				{Classification: "a", TargetCount: 10, TotalVisits: 40, Required: 4, Compliance: 100},
			},
		},
	}
	h := api.NewDashboardHandler(svc)
	router := api.NewDashboardRouter(h)
	handler := injectUser(testAdminUser(), router)

	req := httptest.NewRequest(http.MethodGet, "/frequency?period=2026-03", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp service.FrequencyResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(resp.Items))
	}
	if resp.Items[0].Compliance != 100 {
		t.Errorf("compliance = %f, want 100", resp.Items[0].Compliance)
	}
}

func TestDashboardHandler_DateRangeFilter(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardService{
		activityStats: &service.ActivityStatsResponse{
			ByStatus:   map[string]int{},
			ByCategory: map[string]int{},
			Total:      0,
		},
	}
	h := api.NewDashboardHandler(svc)
	router := api.NewDashboardRouter(h)
	handler := injectUser(testAdminUser(), router)

	req := httptest.NewRequest(http.MethodGet, "/activities?dateFrom=2026-01-01&dateTo=2026-03-31", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDashboardHandler_UserAndTeamFilter(t *testing.T) {
	t.Parallel()
	svc := &stubDashboardService{
		activityStats: &service.ActivityStatsResponse{
			ByStatus:   map[string]int{},
			ByCategory: map[string]int{},
			Total:      0,
		},
	}
	h := api.NewDashboardHandler(svc)
	router := api.NewDashboardRouter(h)
	handler := injectUser(testAdminUser(), router)

	req := httptest.NewRequest(http.MethodGet, "/activities?period=2026-03&userId=user-1&teamId=team-1", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
