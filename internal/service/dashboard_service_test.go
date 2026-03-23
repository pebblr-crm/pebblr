package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub dashboard repo ---

type stubDashboardRepo struct {
	activityStats *store.ActivityStats
	coverageStats *store.CoverageStats
	userStats     []store.UserActivityStats
	err           error
}

func (r *stubDashboardRepo) ActivityStats(_ context.Context, _ rbac.ActivityScope, _ store.DashboardFilter) (*store.ActivityStats, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.activityStats, nil
}

func (r *stubDashboardRepo) CoverageStats(_ context.Context, _ rbac.ActivityScope, _ rbac.TargetScope, _ store.DashboardFilter, _ []string, _ string) (*store.CoverageStats, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.coverageStats, nil
}

func (r *stubDashboardRepo) UserStats(_ context.Context, _ rbac.ActivityScope, _ store.DashboardFilter) ([]store.UserActivityStats, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.userStats, nil
}

func dashboardTestConfig() *config.TenantConfig {
	return &config.TenantConfig{
		Activities: config.ActivitiesConfig{
			Statuses: []config.StatusDef{
				{Key: "planificat", Label: "Planned", Initial: true},
				{Key: "realizat", Label: "Realized"},
				{Key: "anulat", Label: "Cancelled"},
			},
			Types: []config.ActivityTypeConfig{
				{Key: "visit", Label: "Visit", Category: "field"},
				{Key: "administrative", Label: "Administrative", Category: "non_field"},
				{Key: "vacation", Label: "Vacation", Category: "non_field", BlocksFieldActivities: true},
			},
		},
	}
}

func newDashboardSvc(repo *stubDashboardRepo) *service.DashboardService {
	return service.NewDashboardService(repo, rbac.NewEnforcer(), dashboardTestConfig())
}

// --- ActivityStats tests ---

func TestDashboardActivityStats_Success(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		activityStats: &store.ActivityStats{
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
	}
	svc := newDashboardSvc(repo)

	result, err := svc.ActivityStats(context.Background(), adminUser(), "2026-03", store.DashboardFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Period != "2026-03" {
		t.Errorf("expected period 2026-03, got %s", result.Period)
	}
	if result.DateFrom != "2026-03-01" {
		t.Errorf("expected dateFrom 2026-03-01, got %s", result.DateFrom)
	}
	if result.DateTo != "2026-03-31" {
		t.Errorf("expected dateTo 2026-03-31, got %s", result.DateTo)
	}
	if result.Stats.Total != 50 {
		t.Errorf("expected total 50, got %d", result.Stats.Total)
	}

	// Check category breakdown computed from config.
	fieldCount, nonFieldCount := 0, 0
	for _, c := range result.ByCategory {
		switch c.Category {
		case "field":
			fieldCount = c.Count
		case "non_field":
			nonFieldCount = c.Count
		}
	}
	if fieldCount != 40 {
		t.Errorf("expected field count 40, got %d", fieldCount)
	}
	if nonFieldCount != 10 {
		t.Errorf("expected non_field count 10, got %d", nonFieldCount)
	}
}

func TestDashboardActivityStats_InvalidPeriod(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{}
	svc := newDashboardSvc(repo)

	_, err := svc.ActivityStats(context.Background(), adminUser(), "bad", store.DashboardFilter{})
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestDashboardActivityStats_RepScopedByRBAC(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		activityStats: &store.ActivityStats{
			Total:    10,
			ByStatus: []store.StatusCount{},
			ByType:   []store.TypeCount{},
		},
	}
	svc := newDashboardSvc(repo)

	result, err := svc.ActivityStats(context.Background(), repUser(), "2026-03", store.DashboardFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Rep should get scoped results (stub returns 10).
	if result.Stats.Total != 10 {
		t.Errorf("expected total 10 (rep scope), got %d", result.Stats.Total)
	}
}

// --- CoverageStats tests ---

func TestDashboardCoverageStats_Success(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		coverageStats: &store.CoverageStats{
			TotalTargets:    100,
			VisitedTargets:  75,
			CoveragePercent: 75.0,
		},
	}
	svc := newDashboardSvc(repo)

	result, err := svc.CoverageStats(context.Background(), adminUser(), "2026-03", store.DashboardFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalTargets != 100 {
		t.Errorf("expected 100 total targets, got %d", result.TotalTargets)
	}
	if result.VisitedTargets != 75 {
		t.Errorf("expected 75 visited targets, got %d", result.VisitedTargets)
	}
}

func TestDashboardCoverageStats_InvalidPeriod(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{}
	svc := newDashboardSvc(repo)

	_, err := svc.CoverageStats(context.Background(), adminUser(), "invalid", store.DashboardFilter{})
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

// --- UserStats tests ---

func TestDashboardUserStats_Success(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		userStats: []store.UserActivityStats{
			{
				UserID:   "rep-1",
				UserName: "Alice",
				Total:    20,
				ByStatus: []store.StatusCount{
					{Status: "realizat", Count: 15},
					{Status: "planificat", Count: 5},
				},
			},
			{
				UserID:   "rep-2",
				UserName: "Bob",
				Total:    15,
				ByStatus: []store.StatusCount{
					{Status: "realizat", Count: 10},
					{Status: "planificat", Count: 5},
				},
			},
		},
	}
	svc := newDashboardSvc(repo)

	result, err := svc.UserStats(context.Background(), adminUser(), "2026-03", store.DashboardFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 users, got %d", len(result))
	}
	if result[0].UserID != "rep-1" {
		t.Errorf("expected first user rep-1, got %s", result[0].UserID)
	}
	if result[0].Total != 20 {
		t.Errorf("expected first user total 20, got %d", result[0].Total)
	}
}

func TestDashboardUserStats_InvalidPeriod(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{}
	svc := newDashboardSvc(repo)

	_, err := svc.UserStats(context.Background(), adminUser(), "xxx", store.DashboardFilter{})
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestDashboardActivityStats_FebruaryPeriod(t *testing.T) {
	t.Parallel()
	repo := &stubDashboardRepo{
		activityStats: &store.ActivityStats{
			Total:    5,
			ByStatus: []store.StatusCount{},
			ByType:   []store.TypeCount{},
		},
	}
	svc := newDashboardSvc(repo)

	result, err := svc.ActivityStats(context.Background(), adminUser(), "2026-02", store.DashboardFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DateFrom != "2026-02-01" {
		t.Errorf("expected dateFrom 2026-02-01, got %s", result.DateFrom)
	}
	if result.DateTo != "2026-02-28" {
		t.Errorf("expected dateTo 2026-02-28, got %s", result.DateTo)
	}
}
